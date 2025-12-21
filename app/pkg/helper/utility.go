package helper

import (
	"crypto/rand"
	"encoding/hex"
)

// Generates a 8-byte hex string
func TraceID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)

	return hex.EncodeToString(bytes)
}
