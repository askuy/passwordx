package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// SaltSize is the size of the salt used for key derivation
	SaltSize = 32
	// KeySize is the size of the AES-256 key
	KeySize = 32
	// NonceSize is the size of the GCM nonce
	NonceSize = 12
	// PBKDF2Iterations is the number of iterations for PBKDF2
	PBKDF2Iterations = 100000
)

// GenerateSalt generates a cryptographically secure random salt
func GenerateSalt() (string, error) {
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(salt), nil
}

// DeriveKey derives an AES-256 key from a password using PBKDF2
func DeriveKey(password, saltBase64 string) ([]byte, error) {
	salt, err := base64.StdEncoding.DecodeString(saltBase64)
	if err != nil {
		return nil, err
	}
	key := pbkdf2.Key([]byte(password), salt, PBKDF2Iterations, KeySize, sha256.New)
	return key, nil
}

// Encrypt encrypts plaintext using AES-256-GCM
func Encrypt(plaintext string, key []byte) (string, error) {
	if len(key) != KeySize {
		return "", errors.New("invalid key size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
func Decrypt(ciphertextBase64 string, key []byte) (string, error) {
	if len(key) != KeySize {
		return "", errors.New("invalid key size")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < NonceSize {
		return "", errors.New("ciphertext too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := ciphertext[:NonceSize]
	ciphertext = ciphertext[NonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// HashPassword hashes a password using SHA-256 (for storage, use bcrypt in auth)
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// GenerateRandomPassword generates a random password of specified length
func GenerateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;:,.<>?"
	password := make([]byte, length)
	charsetLen := len(charset)

	randomBytes := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, randomBytes); err != nil {
		return "", err
	}

	for i := 0; i < length; i++ {
		password[i] = charset[int(randomBytes[i])%charsetLen]
	}

	return string(password), nil
}
