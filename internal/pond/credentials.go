package pond

import "golang.org/x/crypto/argon2"

// HashPassword derives an Argon2id hash from password and salt.
// Parameters follow OWASP recommendations: m=19456, t=2, p=1, keyLen=32.
func HashPassword(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, 2, 19456, 1, 32)
}
