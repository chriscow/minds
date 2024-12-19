package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func SHA256Hash(data []byte) (string, error) {
	h := sha256.New()
	_, err := h.Write(data)
	if err != nil {
		return "", fmt.Errorf("write data: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
