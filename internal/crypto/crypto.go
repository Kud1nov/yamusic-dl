// Package crypto предоставляет функциональность для шифрования и дешифрования данных.
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

// NonceSize определяет размер nonce для AES-CTR.
const NonceSize = 12

// GenerateSignature генерирует подпись для запроса к API Яндекс Музыки.
// Использует HMAC-SHA256 для вычисления подписи из строки данных.
// Строку данных нужно формировать в нужном порядке до вызова этой функции.
func GenerateSignature(dataString string, signKey string) string {
	// Удаляем запятые из строки данных
	dataString = strings.ReplaceAll(dataString, ",", "")

	// Вычисляем HMAC-SHA256
	h := hmac.New(sha256.New, []byte(signKey))
	h.Write([]byte(dataString))

	// Кодируем результат в Base64 и удаляем последний символ (=)
	sign := base64.StdEncoding.EncodeToString(h.Sum(nil))
	sign = strings.TrimRight(sign, "=")

	return sign
}

// GenerateSignatureFromParams генерирует подпись из параметров запроса.
// Параметры должны быть переданы в правильном порядке.
// Порядок важен: ts, trackId, quality, codecs, transports
func GenerateSignatureFromParams(ts, trackId, quality, codecs, transports, signKey string) string {
	// Формируем строку из параметров в нужном порядке
	dataString := ts + trackId + quality + codecs + transports

	return GenerateSignature(dataString, signKey)
}

// DecryptAesCtr расшифровывает данные, зашифрованные алгоритмом AES в режиме CTR.
func DecryptAesCtr(encryptedData []byte, hexKey string) ([]byte, error) {
	// Преобразуем ключ из hex в байты
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("error decoding key: %w", err)
	}

	// Создаем AES шифр
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("error creating cipher: %w", err)
	}

	// Создаем nonce из 12 нулевых байтов и 4 нулевых байта для счетчика
	// В сумме получаем 16 байтов (размер блока AES)
	iv := make([]byte, aes.BlockSize)
	// Первые 12 байтов - nonce (могут быть все нули)
	// Последние 4 байта - счетчик (начинается с 0)

	// Создаем CTR режим с нашим IV
	stream := cipher.NewCTR(block, iv)

	// Расшифровываем данные
	decrypted := make([]byte, len(encryptedData))
	stream.XORKeyStream(decrypted, encryptedData)

	return decrypted, nil
}
