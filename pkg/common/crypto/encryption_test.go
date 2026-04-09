package crypto

import (
	"strings"
	"testing"
)

func TestNewAESEncryptor_ValidKeys(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"16 bytes key (AES-128)", "1234567890123456", false},
		{"24 bytes key (AES-192)", "123456789012345678901234", false},
		{"32 bytes key (AES-256)", "12345678901234567890123456789012", false},
		{"15 bytes key (invalid)", "123456789012345", true},
		{"17 bytes key (invalid)", "12345678901234567", true},
		{"Empty key (invalid)", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAESEncryptor(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAESEncryptor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAESEncryptor_EncryptDecrypt(t *testing.T) {
	key := "12345678901234567890123456789012"
	encryptor, err := NewAESEncryptor(key)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{"Simple string", "hello world"},
		{"Empty string", ""},
		{"JSON token", `{"access_token":"abc123","expires_in":3600}`},
		{"Long string", strings.Repeat("a", 1000)},
		{"Unicode string", "Hello 世界 🌍"},
		{"Special characters", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := encryptor.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			if tt.plaintext != "" && encrypted == tt.plaintext {
				t.Error("Encrypted text should be different from plaintext")
			}

			decrypted, err := encryptor.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if decrypted != tt.plaintext {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestAESEncryptor_DecryptInvalid(t *testing.T) {
	key := "12345678901234567890123456789012"
	encryptor, _ := NewAESEncryptor(key)

	tests := []struct {
		name       string
		ciphertext string
		wantErr    bool
	}{
		{"Invalid base64", "not-valid-base64!!!", true},
		{"Too short ciphertext", "YWJj", true},
		{"Tampered ciphertext", "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo=", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := encryptor.Decrypt(tt.ciphertext)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAESEncryptor_DifferentEncryptions(t *testing.T) {
	key := "12345678901234567890123456789012"
	encryptor, _ := NewAESEncryptor(key)
	plaintext := "test data"

	encrypted1, _ := encryptor.Encrypt(plaintext)
	encrypted2, _ := encryptor.Encrypt(plaintext)

	if encrypted1 == encrypted2 {
		t.Error("Same plaintext should produce different ciphertexts due to random nonce")
	}

	decrypted1, _ := encryptor.Decrypt(encrypted1)
	decrypted2, _ := encryptor.Decrypt(encrypted2)

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Error("Both ciphertexts should decrypt to the same plaintext")
	}
}

func TestNoOpEncryptor(t *testing.T) {
	encryptor := NewNoOpEncryptor()

	plaintext := "test data"
	encrypted, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	if encrypted != plaintext {
		t.Errorf("NoOpEncryptor.Encrypt() = %v, want %v", encrypted, plaintext)
	}

	decrypted, err := encryptor.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("NoOpEncryptor.Decrypt() = %v, want %v", decrypted, plaintext)
	}
}

func BenchmarkAESEncryptor_Encrypt(b *testing.B) {
	key := "12345678901234567890123456789012"
	encryptor, _ := NewAESEncryptor(key)
	plaintext := "This is a test access token that needs to be encrypted"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encryptor.Encrypt(plaintext)
	}
}

func BenchmarkAESEncryptor_Decrypt(b *testing.B) {
	key := "12345678901234567890123456789012"
	encryptor, _ := NewAESEncryptor(key)
	plaintext := "This is a test access token that needs to be encrypted"
	encrypted, _ := encryptor.Encrypt(plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encryptor.Decrypt(encrypted)
	}
}
