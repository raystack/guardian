package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
)

// AES does encrypt and decrypt using AES algorithm
type AES struct {
	key []byte
}

// NewAES returns string cipher
func NewAES(key string) *AES {
	hasher := md5.New()
	hasher.Write([]byte(key))
	hashedKey := hasher.Sum(nil)

	return &AES{
		key: hashedKey,
	}
}

// Encrypt encrypts a plain text into an encrypted text
func (c *AES) Encrypt(value string) (string, error) {
	stringToEncrypt := []byte(value)

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, stringToEncrypt, nil)

	return hex.EncodeToString(ciphertext), nil
}

// Decrypt decrypts an encrypted text into a plain text
func (c *AES) Decrypt(value string) (string, error) {
	stringToDecrypt, err := hex.DecodeString(value)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := stringToDecrypt[:nonceSize], stringToDecrypt[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
