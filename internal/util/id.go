package util

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateID generates a random ID string
func GenerateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
