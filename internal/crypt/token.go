package crypt


import (
	"crypto/rand"
	"encoding/base64"
	"log"


	// "crypto/sha256"
    // "encoding/hex"
)

func GenerateToken() (string, error) {
    tokenBytes := make([]byte, 32)
    if _, err := rand.Read(tokenBytes); err != nil {
        log.Printf("CRITICAL: Entropy failure: %v", err)
        return "", err
    }
    return base64.RawURLEncoding.EncodeToString(tokenBytes), nil
}