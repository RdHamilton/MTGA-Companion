package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/argon2"
)

const (
	// EncryptionMagicHeader is prepended to encrypted files for identification
	EncryptionMagicHeader = "MTGAENC1"

	// Default Argon2 parameters (RFC 9106 recommendations)
	defaultArgon2Time    = 1
	defaultArgon2Memory  = 64 * 1024 // 64 MB
	defaultArgon2Threads = 4
	defaultArgon2KeyLen  = 32 // 256 bits for AES-256

	// Salt length for key derivation
	saltLength = 32
)

// EncryptionConfig holds configuration for encryption operations.
type EncryptionConfig struct {
	// Password is the encryption password/passphrase
	Password string

	// Argon2Time is the number of iterations for Argon2
	// Higher = more secure but slower
	// Default: 1
	Argon2Time uint32

	// Argon2Memory is the amount of memory to use in KB
	// Higher = more secure but uses more RAM
	// Default: 64 MB (65536 KB)
	Argon2Memory uint32

	// Argon2Threads is the number of threads to use
	// Default: 4
	Argon2Threads uint8
}

// DefaultEncryptionConfig returns encryption config with secure defaults.
func DefaultEncryptionConfig(password string) *EncryptionConfig {
	return &EncryptionConfig{
		Password:      password,
		Argon2Time:    defaultArgon2Time,
		Argon2Memory:  defaultArgon2Memory,
		Argon2Threads: defaultArgon2Threads,
	}
}

// deriveKey derives an encryption key from a password using Argon2id.
// Argon2id is the recommended variant - resistant to both side-channel and GPU attacks.
func deriveKey(password string, salt []byte, config *EncryptionConfig) []byte {
	if config == nil {
		config = DefaultEncryptionConfig(password)
	}

	return argon2.IDKey(
		[]byte(password),
		salt,
		config.Argon2Time,
		config.Argon2Memory,
		config.Argon2Threads,
		defaultArgon2KeyLen,
	)
}

// EncryptData encrypts data using AES-256-GCM with password-based key derivation.
// Returns: salt + nonce + ciphertext + auth tag
func EncryptData(plaintext []byte, config *EncryptionConfig) ([]byte, error) {
	if config == nil || config.Password == "" {
		return nil, fmt.Errorf("encryption config with password required")
	}

	// Generate random salt for key derivation
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive encryption key from password
	key := deriveKey(config.Password, salt, config)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode (Galois/Counter Mode - provides authentication)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Format: salt || nonce || ciphertext (includes auth tag)
	result := make([]byte, 0, len(salt)+len(nonce)+len(ciphertext))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return result, nil
}

// DecryptData decrypts data that was encrypted with EncryptData.
// Expects format: salt + nonce + ciphertext + auth tag
func DecryptData(encrypted []byte, config *EncryptionConfig) ([]byte, error) {
	if config == nil || config.Password == "" {
		return nil, fmt.Errorf("encryption config with password required")
	}

	// Minimum size: salt + nonce + auth tag
	// GCM auth tag is 16 bytes, nonce is 12 bytes
	minSize := saltLength + 12 + 16
	if len(encrypted) < minSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	// Extract salt
	salt := encrypted[:saltLength]
	encrypted = encrypted[saltLength:]

	// Derive key from password
	key := deriveKey(config.Password, salt, config)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, fmt.Errorf("encrypted data too short for nonce")
	}
	nonce := encrypted[:nonceSize]
	ciphertext := encrypted[nonceSize:]

	// Decrypt and verify authentication tag
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (wrong password or corrupted data): %w", err)
	}

	return plaintext, nil
}

// EncryptFile encrypts a file and writes the encrypted data to a new file.
func EncryptFile(sourcePath, destPath string, config *EncryptionConfig) error {
	// Read source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() { _ = sourceFile.Close() }() //nolint:errcheck // Ignore error on cleanup

	plaintext, err := io.ReadAll(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Encrypt data
	encrypted, err := EncryptData(plaintext, config)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// Write magic header + encrypted data
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { _ = destFile.Close() }() //nolint:errcheck // Ignore error on cleanup

	// Write magic header
	if _, err := destFile.Write([]byte(EncryptionMagicHeader)); err != nil {
		return fmt.Errorf("failed to write magic header: %w", err)
	}

	// Write encrypted data
	if _, err := destFile.Write(encrypted); err != nil {
		return fmt.Errorf("failed to write encrypted data: %w", err)
	}

	return nil
}

// DecryptFile decrypts an encrypted file and writes the plaintext to a new file.
func DecryptFile(sourcePath, destPath string, config *EncryptionConfig) error {
	// Read encrypted file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open encrypted file: %w", err)
	}
	defer func() { _ = sourceFile.Close() }() //nolint:errcheck // Ignore error on cleanup

	data, err := io.ReadAll(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to read encrypted file: %w", err)
	}

	// Check for magic header
	if len(data) < len(EncryptionMagicHeader) {
		return fmt.Errorf("file too short to be encrypted")
	}

	header := string(data[:len(EncryptionMagicHeader)])
	if header != EncryptionMagicHeader {
		return fmt.Errorf("file is not encrypted or has wrong format")
	}

	// Remove header
	encrypted := data[len(EncryptionMagicHeader):]

	// Decrypt data
	plaintext, err := DecryptData(encrypted, config)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	// Write to destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { _ = destFile.Close() }() //nolint:errcheck // Ignore error on cleanup

	if _, err := destFile.Write(plaintext); err != nil {
		return fmt.Errorf("failed to write decrypted file: %w", err)
	}

	return nil
}

// IsEncrypted checks if a file is encrypted by checking for the magic header.
func IsEncrypted(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer func() { _ = file.Close() }() //nolint:errcheck // Ignore error on cleanup

	header := make([]byte, len(EncryptionMagicHeader))
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return false, err
	}

	return n == len(EncryptionMagicHeader) && string(header) == EncryptionMagicHeader, nil
}

// EncodePassword encodes a password to base64 for storage (NOT encryption, just encoding).
func EncodePassword(password string) string {
	return base64.StdEncoding.EncodeToString([]byte(password))
}

// DecodePassword decodes a base64-encoded password.
func DecodePassword(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode password: %w", err)
	}
	return string(decoded), nil
}
