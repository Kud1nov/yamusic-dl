// Package crypto provides functionality for encrypting and decrypting data.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

// NonceSize defines the nonce size for AES-CTR.
const NonceSize = 12

// GenerateSignature generates a signature for a request to the Yandex Music API.
// Uses HMAC-SHA256 to calculate the signature from a data string.
// The data string must be formed in the correct order before calling this function.
func GenerateSignature(dataString string, signKey string) string {
	// Remove commas from the data string
	dataString = strings.ReplaceAll(dataString, ",", "")

	// Calculate HMAC-SHA256
	h := hmac.New(sha256.New, []byte(signKey))
	h.Write([]byte(dataString))

	// Encode the result in Base64 and remove the last character (=)
	sign := base64.StdEncoding.EncodeToString(h.Sum(nil))
	sign = strings.TrimRight(sign, "=")

	return sign
}

// GenerateSignatureFromParams generates a signature from request parameters.
// Parameters must be passed in the correct order.
// The order is important: ts, trackId, quality, codecs, transports
func GenerateSignatureFromParams(ts, trackId, quality, codecs, transports, signKey string) string {
	// Form a string from parameters in the correct order
	dataString := ts + trackId + quality + codecs + transports

	return GenerateSignature(dataString, signKey)
}

// DecryptAesCtr decrypts data encrypted with the AES algorithm in CTR mode.
func DecryptAesCtr(encryptedData []byte, hexKey string) ([]byte, error) {
	// Convert key from hex to bytes
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("error decoding key: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("error creating cipher: %w", err)
	}

	// Create nonce from 12 zero bytes and 4 zero bytes for the counter
	// In total, we get 16 bytes (AES block size)
	iv := make([]byte, aes.BlockSize)
	// First 12 bytes - nonce (can be all zeros)
	// Last 4 bytes - counter (starts from 0)

	// Create CTR mode with our IV
	stream := cipher.NewCTR(block, iv)

	// Decrypt data
	decrypted := make([]byte, len(encryptedData))
	stream.XORKeyStream(decrypted, encryptedData)

	return decrypted, nil
}
