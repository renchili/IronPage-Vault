package app

import (
    "strconv"
    "strings"
)

func redactedVersionPath(currentPath string, newVersion int) string {
    return strings.TrimSuffix(currentPath, ".pdf") + "_redacted_v" + strconv.Itoa(newVersion) + ".pdf"
}

func redactionTransformMarker() string {
    return "redaction burn-in confirmed"
}
