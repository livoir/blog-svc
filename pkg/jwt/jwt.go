package jwt

import (
	"crypto/rsa"
	"os"
	"path/filepath"

	"github.com/golang-jwt/jwt/v5"
)

func NewJWT(privateKeyPath, publicKeyPath string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	filepath.Clean(privateKeyPath)
	filepath.Clean(publicKeyPath)
	privBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, nil, err
	}
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		return nil, nil, err
	}
	pubBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, nil, err
	}
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		return nil, nil, err
	}

	return privKey, pubKey, nil
}
