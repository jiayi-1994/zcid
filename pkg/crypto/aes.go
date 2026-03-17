package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	EnvEncryptionKey = "ZCID_ENCRYPTION_KEY"
	keyLength        = 32 // AES-256
	nonceLength      = 12 // GCM standard nonce
)

var (
	ErrKeyNotSet   = errors.New("encryption key not set")
	ErrKeyTooShort = errors.New("encryption key must be 32 bytes")
	ErrDecryptFail = errors.New("failed to decrypt value")
	ErrInvalidData = errors.New("invalid encrypted data")
)

type AESCrypto struct {
	gcm cipher.AEAD
}

func NewAESCrypto(key []byte) (*AESCrypto, error) {
	if len(key) != keyLength {
		return nil, ErrKeyTooShort
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	return &AESCrypto{gcm: gcm}, nil
}

func NewAESCryptoFromEnv() (*AESCrypto, error) {
	keyStr := os.Getenv(EnvEncryptionKey)
	if keyStr == "" {
		return nil, ErrKeyNotSet
	}
	key := []byte(keyStr)
	if len(key) != keyLength {
		return nil, fmt.Errorf("%w: got %d bytes", ErrKeyTooShort, len(key))
	}
	return NewAESCrypto(key)
}

func (c *AESCrypto) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, nonceLength)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := c.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (c *AESCrypto) Decrypt(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", ErrInvalidData
	}

	if len(data) < nonceLength {
		return "", ErrInvalidData
	}

	nonce := data[:nonceLength]
	ciphertext := data[nonceLength:]

	plaintext, err := c.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrDecryptFail
	}

	return string(plaintext), nil
}
