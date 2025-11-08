package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptDecryptData(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
		password  string
	}{
		{
			name:      "simple text",
			plaintext: "Hello, World!",
			password:  "test-password",
		},
		{
			name:      "empty string",
			plaintext: "",
			password:  "test-password",
		},
		{
			name:      "long text",
			plaintext: string(make([]byte, 10000)),
			password:  "secure-password-123",
		},
		{
			name:      "special characters",
			plaintext: "Test with 中文 and émojis 🎮",
			password:  "pássword-with-spëcial",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultEncryptionConfig(tt.password)

			// Encrypt
			encrypted, err := EncryptData([]byte(tt.plaintext), config)
			if err != nil {
				t.Fatalf("EncryptData() error = %v", err)
			}

			// Verify encrypted data is different from plaintext
			if bytes.Equal(encrypted, []byte(tt.plaintext)) {
				t.Error("Encrypted data should be different from plaintext")
			}

			// Decrypt
			decrypted, err := DecryptData(encrypted, config)
			if err != nil {
				t.Fatalf("DecryptData() error = %v", err)
			}

			// Verify decrypted matches original
			if string(decrypted) != tt.plaintext {
				t.Errorf("Decrypted data = %q, want %q", string(decrypted), tt.plaintext)
			}
		})
	}
}

func TestDecryptDataWrongPassword(t *testing.T) {
	plaintext := []byte("secret message")
	config := DefaultEncryptionConfig("correct-password")

	encrypted, err := EncryptData(plaintext, config)
	if err != nil {
		t.Fatalf("EncryptData() error = %v", err)
	}

	// Try to decrypt with wrong password
	wrongConfig := DefaultEncryptionConfig("wrong-password")
	_, err = DecryptData(encrypted, wrongConfig)
	if err == nil {
		t.Error("DecryptData() with wrong password should fail")
	}
}

func TestEncryptDataNoPassword(t *testing.T) {
	plaintext := []byte("test data")
	config := &EncryptionConfig{Password: ""}

	_, err := EncryptData(plaintext, config)
	if err == nil {
		t.Error("EncryptData() with no password should fail")
	}
}

func TestDecryptDataCorrupted(t *testing.T) {
	plaintext := []byte("test data")
	config := DefaultEncryptionConfig("password")

	encrypted, err := EncryptData(plaintext, config)
	if err != nil {
		t.Fatalf("EncryptData() error = %v", err)
	}

	// Corrupt the encrypted data
	encrypted[len(encrypted)-1] ^= 0xFF

	_, err = DecryptData(encrypted, config)
	if err == nil {
		t.Error("DecryptData() with corrupted data should fail")
	}
}

func TestEncryptDecryptFile(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	sourcePath := filepath.Join(tmpDir, "source.txt")
	encryptedPath := filepath.Join(tmpDir, "encrypted.enc")
	decryptedPath := filepath.Join(tmpDir, "decrypted.txt")

	// Create source file
	originalContent := []byte("This is a test file with some content\nMultiple lines\n")
	if err := os.WriteFile(sourcePath, originalContent, 0o600); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	password := "file-encryption-password"
	config := DefaultEncryptionConfig(password)

	// Encrypt file
	if err := EncryptFile(sourcePath, encryptedPath, config); err != nil {
		t.Fatalf("EncryptFile() error = %v", err)
	}

	// Verify encrypted file exists
	if _, err := os.Stat(encryptedPath); os.IsNotExist(err) {
		t.Fatal("Encrypted file was not created")
	}

	// Verify encrypted file has magic header
	isEnc, err := IsEncrypted(encryptedPath)
	if err != nil {
		t.Fatalf("IsEncrypted() error = %v", err)
	}
	if !isEnc {
		t.Error("Encrypted file should have magic header")
	}

	// Decrypt file
	if err := DecryptFile(encryptedPath, decryptedPath, config); err != nil {
		t.Fatalf("DecryptFile() error = %v", err)
	}

	// Verify decrypted content matches original
	decryptedContent, err := os.ReadFile(decryptedPath)
	if err != nil {
		t.Fatalf("Failed to read decrypted file: %v", err)
	}

	if !bytes.Equal(decryptedContent, originalContent) {
		t.Errorf("Decrypted content does not match original")
	}
}

func TestDecryptFileWrongPassword(t *testing.T) {
	tmpDir := t.TempDir()

	sourcePath := filepath.Join(tmpDir, "source.txt")
	encryptedPath := filepath.Join(tmpDir, "encrypted.enc")
	decryptedPath := filepath.Join(tmpDir, "decrypted.txt")

	// Create and encrypt file
	originalContent := []byte("secret content")
	if err := os.WriteFile(sourcePath, originalContent, 0o600); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	correctPassword := "correct-password"
	if err := EncryptFile(sourcePath, encryptedPath, DefaultEncryptionConfig(correctPassword)); err != nil {
		t.Fatalf("EncryptFile() error = %v", err)
	}

	// Try to decrypt with wrong password
	wrongPassword := "wrong-password"
	err := DecryptFile(encryptedPath, decryptedPath, DefaultEncryptionConfig(wrongPassword))
	if err == nil {
		t.Error("DecryptFile() with wrong password should fail")
	}
}

func TestIsEncrypted(t *testing.T) {
	tmpDir := t.TempDir()

	// Create encrypted file
	encryptedPath := filepath.Join(tmpDir, "encrypted.enc")
	plainPath := filepath.Join(tmpDir, "plain.txt")

	plainContent := []byte("test content")
	if err := os.WriteFile(plainPath, plainContent, 0o600); err != nil {
		t.Fatalf("Failed to create plain file: %v", err)
	}

	config := DefaultEncryptionConfig("password")
	if err := EncryptFile(plainPath, encryptedPath, config); err != nil {
		t.Fatalf("EncryptFile() error = %v", err)
	}

	// Test encrypted file
	isEnc, err := IsEncrypted(encryptedPath)
	if err != nil {
		t.Fatalf("IsEncrypted() error = %v", err)
	}
	if !isEnc {
		t.Error("Encrypted file should be detected as encrypted")
	}

	// Test plain file
	isEnc, err = IsEncrypted(plainPath)
	if err != nil {
		t.Fatalf("IsEncrypted() error = %v", err)
	}
	if isEnc {
		t.Error("Plain file should not be detected as encrypted")
	}
}

func TestEncodeDecodePassword(t *testing.T) {
	tests := []string{
		"simple",
		"with spaces",
		"with-special-chars!@#$%",
		"中文密码",
		"",
	}

	for _, password := range tests {
		encoded := EncodePassword(password)
		decoded, err := DecodePassword(encoded)
		if err != nil {
			t.Fatalf("DecodePassword() error = %v", err)
		}
		if decoded != password {
			t.Errorf("Decoded password = %q, want %q", decoded, password)
		}
	}
}

func TestArgon2Parameters(t *testing.T) {
	plaintext := []byte("test data")
	password := "password"

	// Test with custom Argon2 parameters
	config := &EncryptionConfig{
		Password:      password,
		Argon2Time:    2,
		Argon2Memory:  128 * 1024,
		Argon2Threads: 8,
	}

	encrypted, err := EncryptData(plaintext, config)
	if err != nil {
		t.Fatalf("EncryptData() with custom config error = %v", err)
	}

	decrypted, err := DecryptData(encrypted, config)
	if err != nil {
		t.Fatalf("DecryptData() error = %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("Decrypted data does not match plaintext with custom Argon2 parameters")
	}
}

func TestEncryptionDeterminism(t *testing.T) {
	// Encrypting the same data twice should produce different ciphertexts
	// (because of random salt and nonce)
	plaintext := []byte("test data")
	password := "password"
	config := DefaultEncryptionConfig(password)

	encrypted1, err := EncryptData(plaintext, config)
	if err != nil {
		t.Fatalf("EncryptData() error = %v", err)
	}

	encrypted2, err := EncryptData(plaintext, config)
	if err != nil {
		t.Fatalf("EncryptData() error = %v", err)
	}

	if bytes.Equal(encrypted1, encrypted2) {
		t.Error("Encrypting same data twice should produce different ciphertexts")
	}

	// But both should decrypt to the same plaintext
	decrypted1, _ := DecryptData(encrypted1, config)
	decrypted2, _ := DecryptData(encrypted2, config)

	if !bytes.Equal(decrypted1, plaintext) || !bytes.Equal(decrypted2, plaintext) {
		t.Error("Both encrypted versions should decrypt to the same plaintext")
	}
}
