package app

import (
	"io"

	"ironpage-vault/internal/platform"
)

func fileDigest(r io.Reader) (string, error) { return platform.FileDigest(r) }
