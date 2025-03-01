package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

func Encrypt(plainText string, key []byte) (string, error) {
	if !isValidAESKey(key) {
		return "", errors.New("invalid AES key length (must be 16, 24, or 32 bytes long)")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func Decrypt(cipherTextBase64 string, key []byte) (string, error) {
	if !isValidAESKey(key) {
		return "", errors.New("invalid AES key length (must be 16, 24, or 32 bytes long)")
	}

	cipherText, err := base64.StdEncoding.DecodeString(cipherTextBase64)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(cipherText) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce := cipherText[:nonceSize]
	encryptedData := cipherText[nonceSize:]
	plainText, err := aesGCM.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

func isValidAESKey(key []byte) bool {
	validLengths := []int{16, 24, 32} // AES-128, AES-192, AES-256
	for _, length := range validLengths {
		if len(key) == length {
			return true
		}
	}
	return false
}
