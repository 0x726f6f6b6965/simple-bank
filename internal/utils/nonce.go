package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func GenerateNonce(n int) (string, error) {
	nonceBytes := make([]byte, n)
	_, err := rand.Read(nonceBytes)
	if err != nil {
		return "", fmt.Errorf("could not generate nonce")
	}

	return base64.URLEncoding.EncodeToString(nonceBytes), nil
}
