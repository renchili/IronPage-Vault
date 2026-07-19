#!/usr/bin/env python3
"""Exercise the acceptance-only browser probe through Chrome DevTools Protocol.

Uses only the Python standard library so the air-gapped project does not acquire
an npm, Playwright, Selenium, or remote browser dependency.
"""

from __future__ import annotations

import argparse
import base64
import hashlib
import json
import os
import secrets
import socket
import struct
import subprocess
import sys
import tempfile
import time
import urllib.request
from pathlib import Path
from typing import Any
from urllib.parse import urlparse


class CDPError(RuntimeError):
    pass


class WebSocket:
    def __init__(self, url: str, timeout: float = 15.0) -> None:
        parsed = urlparse(url)
        if parsed.scheme != "ws":
            raise CDPError(f"unsupported DevTools websocket scheme: {parsed.scheme}")
        host = parsed.hostname or "127.0.0.1"
        port = parsed.port or 80
        path = parsed.path or "/"
        if parsed.query:
            path += "?" + parsed.query

        self.sock = socket.create_connection((host, port), timeout=timeout)
        self.sock.settimeout(timeout)
        key = base64.b64encode(os.urandom(16)).decode("ascii")
        request = (
            f"GET {path} HTTP/1.1\r\n"
            f"Host: {host}:{port}\r\n"
            "Upgrade: websocket\r\n"
            "Connection: Upgrade\r\n"
            f"Sec-WebSocket-Key: {key}\r\n"
            "Sec-WebSocket-Version: 13\r\n\r\n"
        )
        self.sock.sendall(request.encode("ascii"))
        response = self._read_until(b"\r\n\r\n")
        status_line = response.split(b"\r\n", 1)[0]
        if b" 101 " not in status_line:
            raise CDPError(f"websocket upgrade failed: {status_line.decode(errors='replace')}")
        accept_expected = base64.b64encode(
            hashlib.sha1((key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11").encode("ascii")).digest()
        )
        headers = response.decode("latin1").split("\r\n")
        accept_actual = ""
        for header in headers[1:]:
            if header.lower().startswith("sec-websocket-accept:"):
                accept_actual = header.split(":", 1)[1].strip()
                break
        if accept_actual != accept_expected.decode("ascii"):
            raise CDPError("invalid websocket accept header")
        self._next_id = 1
        self._continuation = bytearray()

    def _read_until(self, marker: bytes) -> bytes:
        data = bytearray()
        while marker not in data:
            chunk = self.sock.recv(4096)
            if not chunk:
                raise CDPError("websocket closed during handshake")
            data.extend(chunk)
        return bytes(data)

    def _read_exact(self, length: int) -> bytes:
        data = bytearray()
        while len(data) < length:
            chunk = self.sock.recv(length - len(data))
            if not chunk:
                raise CDPError("websocket closed")
            data.extend(chunk)
        return bytes(data)

    def send_text(self, text: str) -> None:
        payload = text.encode("utf-8")
        mask = os.urandom(4)
        first = 0x81
        length = len(payload)
        if length < 126:
            header = bytes((first, 0x80 | length))
        elif length < 65536:
            header = bytes((first, 0x80 | 126)) + struct.pack("!H", length)
        else:
            header = bytes((first, 0x80 | 127)) + struct.pack("!Q", length)
        masked = bytes(value ^ mask[index % 4] for index, value in enumerate(payload))
        self.sock.sendall(header + mask + masked)

    def _send_control(self, opcode: int, payload: bytes) -> None:
        mask = os.urandom(4)
        header = bytes((0x80 | opcode, 0x80 | len(payload)))
        masked = bytes(value ^ mask[index % 4] for index, value in enumerate(payload))
        self.sock.sendall(header + mask + masked)

    def recv_text(self) -> str:
        while True:
            first, second = self._read_exact(2)
            fin = bool(first & 0x80)
            opcode = first & 0x0F
            masked = bool(second & 0x80)
            length = second & 0x7F
            if length == 126:
                length = struct.unpack("!H", self._read_exact(2))[0]
            elif length == 127:
                length = struct.unpack("!Q", self._read_exact(8))[0]
            mask = self._read_exact(4) if masked else b""
            payload = self._read_exact(length)
            if masked:
                payload = bytes(value ^ mask[index % 4] for index, value in enumerate(payload))

            if opcode == 0x8:
                raise CDPError("DevTools websocket closed")
            if opcode == 0x9:
                self._send_control(0xA, payload)
                continue
            if opcode == 0xA:
                continue
            if opcode == 0x1:
                self._continuation = bytearray(payload)
                if fin:
                    text = self._continuation.decode("utf-8")
                    self._continuation.clear()
                    return text
                continue
            if opcode == 0x0:
                self._continuation.extend(payload)
                if fin:
                    text = self._continuation.decode("utf-8")
                    self._continuation.clear()
                    return text

    def command(self, method: str, params: dict[str, Any] | None = None, session_id: str | None = None) -> dict[str, Any]:
        command_id = self._next_id
        self._next_id += 1
        message: dict[str, Any] = {"id": command_id, "method": method}
        if params:
            message["params"] = params
        if session_id:
            message["sessionId"] = session_id
        self.send_text(json.dumps(message, separators=(",", ":")))
        while True:
            response = json.loads(self.recv_text())
            if response.get("id") != command_id:
                continue
            if "error" in response:
                raise CDPError(f"{method}: {response['error']}")
            return response.get("result", {})

    def close(self) -> None:
        try:
            self._send_control(0x8, b"")
        except OSError:
            pass
        self.sock.close()


class BrowserProbe:
    def __init__(self, browser: str, base_url: str, output_dir: Path) -> None:
        self.browser = browser
        self.base_url = base_url.rstrip("/")
        self.output_dir = output_dir
        self.output_dir.mkdir(parents=True, exist_ok=True)
        self.profile_dir = tempfile.TemporaryDirectory(prefix="ironpage-ui-profile-")
        self.process: subprocess.Popen[str] | None = None
        self.ws: WebSocket | None = None
        self.session_id = ""
        self.assertions: list[dict[str, Any]] = []

    def start(self) -> None:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as probe:
            probe.bind(("127.0.0.1", 0))
            port = probe.getsockname()[1]
        log_path = self.output_dir / "browser.log"
        log_handle = log_path.open("w", encoding="utf-8")
        self.process = subprocess.Popen(
            [
                self.browser,
                "--headless=new",
                "--no-sandbox",
                "--disable-gpu",
                "--disable-dev-shm-usage",
                f"--remote-debugging-port={port}",
                f"--user-data-dir={self.profile_dir.name}",
                "--window-size=1440,1100",
                "about:blank",
            ],
            stdout=log_handle,
            stderr=subprocess.STDOUT,
            text=True,
        )
        version_url = f"http://127.0.0.1:{port}/json/version"
        metadata: dict[str, Any] | None = None
        for _ in range(100):
            if self.process.poll() is not None:
                raise CDPError(f"browser exited before DevTools became ready; see {log_path}")
            try:
                with urllib.request.urlopen(version_url, timeout=0.5) as response:
                    metadata = json.load(response)
                break
            except Exception:
                time.sleep(0.1)
        if not metadata:
            raise CDPError(f"DevTools endpoint did not become ready; see {log_path}")

        self.browser_metadata = {
            "browser": metadata.get("Browser"),
            "protocol_version": metadata.get("Protocol-Version"),
            "user_agent": metadata.get("User-Agent"),
        }
        self.ws = WebSocket(str(metadata["webSocketDebuggerUrl"]))
        target = self.ws.command("Target.createTarget", {"url": "about:blank"})
        attached = self.ws.command("Target.attachToTarget", {"targetId": target["targetId"], "flatten": True})
        self.session_id = str(attached["sessionId"])
        self.command("Page.enable")
        self.command("Runtime.enable")
        self.command("Network.enable")
        navigation = self.command("Page.navigate", {"url": f"{self.base_url}/ui/"})
        if navigation.get("errorText"):
            raise CDPError(f"navigation failed: {navigation['errorText']}")
        expected_url = json.dumps(f"{self.base_url}/ui/")
        self.wait_until(f"location.href === {expected_url} && document.readyState === 'complete'", timeout=15)
        self.wait_until("document.getElementById('health-text') !== null", timeout=10)
        self.wait_until("document.getElementById('health-text').textContent !== 'checking /healthz'", timeout=10)

    def command(self, method: str, params: dict[str, Any] | None = None) -> dict[str, Any]:
        if not self.ws:
            raise CDPError("browser is not started")
        return self.ws.command(method, params, self.session_id)

    def evaluate(self, expression: str) -> Any:
        result = self.command(
            "Runtime.evaluate",
            {"expression": expression, "returnByValue": True, "awaitPromise": True},
        )
        remote = result.get("result", {})
        if remote.get("subtype") == "error":
            raise CDPError(f"JavaScript evaluation failed: {remote}")
        return remote.get("value")

    def wait_until(self, expression: str, timeout: float = 10.0) -> None:
        deadline = time.monotonic() + timeout
        while time.monotonic() < deadline:
            if self.evaluate(expression):
                return
            time.sleep(0.1)
        raise CDPError(f"timed out waiting for: {expression}")

    def record(self, name: str, condition: bool, detail: Any) -> None:
        entry = {"name": name, "passed": bool(condition), "detail": detail}
        self.assertions.append(entry)
        if not condition:
            raise AssertionError(f"{name}: {detail}")
        print(f"PASS ui: {name}")

    def set_input(self, element_id: str, value: str) -> None:
        encoded = json.dumps(value)
        self.evaluate(
            f"""(() => {{
                const element = document.getElementById({json.dumps(element_id)});
                element.value = {encoded};
                element.dispatchEvent(new Event('input', {{bubbles:true}}));
                element.dispatchEvent(new Event('change', {{bubbles:true}}));
                return true;
            }})()"""
        )

    def element_center(self, element_id: str) -> tuple[float, float]:
        rect = self.evaluate(
            f"""(() => {{
                const r = document.getElementById({json.dumps(element_id)}).getBoundingClientRect();
                return {{x:r.left+r.width/2, y:r.top+r.height/2}};
            }})()"""
        )
        return float(rect["x"]), float(rect["y"])

    def click(self, element_id: str) -> None:
        x, y = self.element_center(element_id)
        self.command("Input.dispatchMouseEvent", {"type": "mouseMoved", "x": x, "y": y})
        self.command("Input.dispatchMouseEvent", {"type": "mousePressed", "x": x, "y": y, "button": "left", "clickCount": 1})
        self.command("Input.dispatchMouseEvent", {"type": "mouseReleased", "x": x, "y": y, "button": "left", "clickCount": 1})

    def key(self, key: str, code: str, key_code: int, text: str | None = None) -> None:
        base = {"key": key, "code": code, "windowsVirtualKeyCode": key_code, "nativeVirtualKeyCode": key_code}
        self.command("Input.dispatchKeyEvent", {"type": "keyDown", **base})
        if text is not None:
            self.command("Input.dispatchKeyEvent", {"type": "char", **base, "text": text, "unmodifiedText": text})
        self.command("Input.dispatchKeyEvent", {"type": "keyUp", **base})

    def screenshot(self, name: str) -> str:
        result = self.command("Page.captureScreenshot", {"format": "png", "captureBeyondViewport": True})
        path = self.output_dir / f"{name}.png"
        path.write_bytes(base64.b64decode(result["data"]))
        if path.stat().st_size == 0:
            raise CDPError(f"empty screenshot: {path}")
        return path.name

    def close(self) -> None:
        if self.ws:
            self.ws.close()
        if self.process and self.process.poll() is None:
            self.process.terminate()
            try:
                self.process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                self.process.kill()
                self.process.wait(timeout=5)
        self.profile_dir.cleanup()


def run(args: argparse.Namespace) -> None:
    out = Path(args.output_dir)
    probe = BrowserProbe(args.browser, args.base_url, out)
    screenshots: list[str] = []
    try:
        probe.start()
        probe.record("health status rendered", probe.evaluate("document.getElementById('health-text').textContent") == "healthy", probe.evaluate("document.getElementById('health-text').textContent"))
        probe.record("login output is live status", probe.evaluate("document.getElementById('login-output').getAttribute('aria-live')") == "polite", probe.evaluate("document.getElementById('login-output').outerHTML"))

        probe.click("login-button")
        probe.wait_until("document.getElementById('login-output').dataset.state === 'error'")
        probe.record("missing username reported", probe.evaluate("document.getElementById('login-output').textContent") == "enter an acceptance username", probe.evaluate("document.getElementById('login-output').textContent"))
        probe.record("missing username receives focus", probe.evaluate("document.activeElement.id") == "username", probe.evaluate("document.activeElement.id"))
        probe.record("missing username marked invalid", probe.evaluate("document.getElementById('username').getAttribute('aria-invalid')") == "true", probe.evaluate("document.getElementById('username').getAttribute('aria-invalid')"))
        screenshots.append(probe.screenshot("01-missing-input"))

        probe.set_input("username", args.username)
        probe.set_input("password", "incorrect-" + secrets.token_hex(8))
        probe.click("login-button")
        probe.wait_until("document.getElementById('login-output').dataset.state === 'error' && document.getElementById('login-output').textContent.includes('INVALID_CREDENTIALS')")
        probe.record("incorrect credentials expose API error code", "INVALID_CREDENTIALS" in str(probe.evaluate("document.getElementById('login-output').textContent")), probe.evaluate("document.getElementById('login-output').textContent"))
        screenshots.append(probe.screenshot("02-invalid-credentials"))

        probe.set_input("password", args.password)
        probe.click("login-button")
        probe.wait_until("document.getElementById('login-output').dataset.state === 'success'")
        success_text = str(probe.evaluate("document.getElementById('login-output').textContent"))
        probe.record("successful click login reports role", '"role": "Editor"' in success_text, success_text)
        probe.record("successful click login reports token type", '"token_type": "Bearer"' in success_text, success_text)
        screenshots.append(probe.screenshot("03-successful-login"))

        probe.click("username")
        probe.record("username is focusable", probe.evaluate("document.activeElement.id") == "username", probe.evaluate("document.activeElement.id"))
        probe.key("Tab", "Tab", 9)
        probe.record("Tab moves to password", probe.evaluate("document.activeElement.id") == "password", probe.evaluate("document.activeElement.id"))
        probe.key("Tab", "Tab", 9)
        probe.record("Tab moves to submit button", probe.evaluate("document.activeElement.id") == "login-button", probe.evaluate("document.activeElement.id"))

        probe.set_input("password", args.password)
        probe.click("password")
        probe.key("Enter", "Enter", 13, "\r")
        probe.wait_until("document.getElementById('login-output').dataset.state === 'success'")
        probe.record("Enter submits the form", '"role": "Editor"' in str(probe.evaluate("document.getElementById('login-output').textContent")), probe.evaluate("document.getElementById('login-output').textContent"))
        screenshots.append(probe.screenshot("04-keyboard-submit"))

        probe.command(
            "Network.emulateNetworkConditions",
            {"offline": True, "latency": 0, "downloadThroughput": 0, "uploadThroughput": 0, "connectionType": "none"},
        )
        probe.click("login-button")
        probe.wait_until("document.getElementById('login-output').dataset.state === 'error' && document.getElementById('login-output').textContent.includes('request failed')")
        probe.record("network failure gives retry guidance", "retry" in str(probe.evaluate("document.getElementById('login-output').textContent")), probe.evaluate("document.getElementById('login-output').textContent"))
        screenshots.append(probe.screenshot("05-network-failure"))

        probe.command(
            "Network.emulateNetworkConditions",
            {"offline": False, "latency": 0, "downloadThroughput": -1, "uploadThroughput": -1, "connectionType": "wifi"},
        )
        probe.click("login-button")
        probe.wait_until("document.getElementById('login-output').dataset.state === 'success'")
        probe.record("retry succeeds after network recovery", '"role": "Editor"' in str(probe.evaluate("document.getElementById('login-output').textContent")), probe.evaluate("document.getElementById('login-output').textContent"))
        screenshots.append(probe.screenshot("06-retry-success"))

        payload = {
            "generated_at_epoch": time.time(),
            "base_url": probe.base_url,
            "browser_binary": args.browser,
            "browser_metadata": probe.browser_metadata,
            "username": args.username,
            "password_recorded": False,
            "assertions": probe.assertions,
            "screenshots": screenshots,
            "overall_status": "passed",
        }
        (out / "interaction.json").write_text(json.dumps(payload, indent=2), encoding="utf-8")
        print(f"PASS ui-suite: interaction evidence written to {out / 'interaction.json'}")
    finally:
        probe.close()


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--browser", required=True)
    parser.add_argument("--base-url", required=True)
    parser.add_argument("--username", required=True)
    parser.add_argument("--password", required=True)
    parser.add_argument("--output-dir", required=True)
    return parser.parse_args()


if __name__ == "__main__":
    try:
        run(parse_args())
    except Exception as exc:
        print(f"FAIL ui-suite: {exc}", file=sys.stderr)
        raise
