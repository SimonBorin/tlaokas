package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

var key = []byte("this_is_32_bytes_secret_key!!!!!") // 32 bytes key for AES-256 encryption

// Encrypt encrypts a plaintext string using AES-CTR and returns the ciphertext and IV.
func Encrypt(plaintext string) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, nil, err
	}

	ciphertext := make([]byte, len(plaintext))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext, []byte(plaintext))
	return ciphertext, iv, nil
}

// Decrypt decrypts ciphertext using AES-CTR with the given IV.
func Decrypt(ciphertext, iv []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(ciphertext) == 0 || len(iv) != aes.BlockSize {
		return "", errors.New("invalid input")
	}

	plaintext := make([]byte, len(ciphertext))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plaintext, ciphertext)
	return string(plaintext), nil
}
