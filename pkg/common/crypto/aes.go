package crypto

import (
	"crypto/sha256"
)

// KeyFromString derives a 32-byte AES-256 key from an arbitrary string
// using SHA-256. This allows using JWT secrets or other variable-length
// strings as encryption keys.
func KeyFromString(secret string) []byte {
	hash := sha256.Sum256([]byte(secret))
	return hash[:]
}

// NewAESEncryptorFromSecret creates an AES-256-GCM Encryptor from an arbitrary
// string secret by deriving a 32-byte key via SHA-256.
func NewAESEncryptorFromSecret(secret string) (Encryptor, error) {
	key := KeyFromString(secret)
	return &aesEncryptor{key: key}, nil
}
