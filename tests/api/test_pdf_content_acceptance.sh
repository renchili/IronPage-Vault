#!/usr/bin/env bash
set -u -o pipefail
. tests/api/lib.sh
FAIL=0
: "${EDITOR_TOKEN:?set EDITOR_TOKEN}"

command -v python3 >/dev/null 2>&1 || { echo "FAIL api: python3 missing"; exit 1; }

pdftotext_to_file() {
  local input="$1"
  local output="$2"

  if command -v pdftotext >/dev/null 2>&1; then
    pdftotext "$input" "$output"
    return
  fi

  if command -v docker >/dev/null 2>&1 && [ -n "${IRONPAGE_ENV_FILE:-}" ]; then
    if docker compose --env-file "$IRONPAGE_ENV_FILE" ps ironpage >/dev/null 2>&1; then
      docker compose --env-file "$IRONPAGE_ENV_FILE" exec -T ironpage sh -c 'cat >/tmp/ironpage_pdf_probe.pdf && pdftotext /tmp/ironpage_pdf_probe.pdf -' < "$input" > "$output"
      return
    fi
  fi

  echo "FAIL api: pdftotext missing and generated Compose environment is unavailable"
  return 1
}

python3 - <<'PY'
from reportlab.pdfgen import canvas
c = canvas.Canvas("/tmp/ironpage_secret_probe.pdf", pagesize=(612, 792))
c.setFont("Helvetica", 14)
c.drawString(72, 700, "SECRET_NEVER_APPEAR")
c.drawString(72, 650, "SAFE_PUBLIC_TEXT")
c.save()
PY

code=$(curl -s -o "$BODY" -w "%{http_code}" "$BASE_URL/api/documents" -X POST -H "Authorization: Bearer $EDITOR_TOKEN" -H "X-Request-ID: $(reqid)" -H "X-Request-Timestamp: $(ts)" -F "title=PDF Content Probe" -F "file=@/tmp/ironpage_secret_probe.pdf")
expect_code "upload pdf content probe" 201 "$code" || FAIL=$((FAIL+1))
DOC_ID="$(json_field data.id)"
if [ -z "$DOC_ID" ]; then echo "FAIL api: content probe doc id missing"; exit 1; fi

code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/redactions" '{"page":1,"x":60,"y":70,"width":260,"height":60,"reason":"remove secret"}')
expect_code "stage content redaction" 201 "$code" || FAIL=$((FAIL+1))
RED_ID="$(json_field id)"
if [ -z "$RED_ID" ]; then echo "FAIL api: redaction id missing"; exit 1; fi
code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/redactions/$RED_ID/confirm" '{}')
expect_code "confirm content redaction" 200 "$code" || FAIL=$((FAIL+1))

curl -s -o /tmp/ironpage_redacted_probe.pdf -H "Authorization: Bearer $EDITOR_TOKEN" -H "X-Request-ID: $(reqid)" -H "X-Request-Timestamp: $(ts)" "$BASE_URL/api/documents/$DOC_ID/file"
pdftotext_to_file /tmp/ironpage_redacted_probe.pdf /tmp/ironpage_redacted_probe.txt
if grep -q "SECRET_NEVER_APPEAR" /tmp/ironpage_redacted_probe.txt; then
  echo "FAIL api: redacted PDF still exposes target text"
  cat /tmp/ironpage_redacted_probe.txt
  FAIL=$((FAIL+1))
else
  echo "PASS api: redacted PDF no longer exposes target text"
fi

code=$(auth_post_json "$EDITOR_TOKEN" "/api/documents/$DOC_ID/bates" '{"prefix":"CNT-","suffix":"","zero_padding":3,"start":1}')
expect_code "apply content Bates" 201 "$code" || FAIL=$((FAIL+1))
curl -s -o /tmp/ironpage_bates_probe.pdf -H "Authorization: Bearer $EDITOR_TOKEN" -H "X-Request-ID: $(reqid)" -H "X-Request-Timestamp: $(ts)" "$BASE_URL/api/documents/$DOC_ID/file"
pdftotext_to_file /tmp/ironpage_bates_probe.pdf /tmp/ironpage_bates_probe.txt
if grep -q "CNT-001" /tmp/ironpage_bates_probe.txt; then
  echo "PASS api: Bates label is extractable from generated PDF"
else
  echo "FAIL api: Bates label missing from generated PDF text"
  cat /tmp/ironpage_bates_probe.txt
  FAIL=$((FAIL+1))
fi

exit "$FAIL"
