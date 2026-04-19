package pond_test

import (
	"bytes"
	"testing"

	"golang.org/x/crypto/argon2"

	"github.com/arturovm/pond/internal/pond"
)

func TestHashPassword_DifferentSalts_ReturnDifferentHashes(t *testing.T) {
	password := "secret"

	hash1 := pond.HashPassword(password, []byte("saltsaltsaltsalt"))
	hash2 := pond.HashPassword(password, []byte("othersaltothersalt"))

	if bytes.Equal(hash1, hash2) {
		t.Error("expected different hashes for different salts, got equal hashes")
	}
}

func TestHashPassword_DifferentPasswords_ReturnDifferentHashes(t *testing.T) {
	salt := []byte("saltsaltsaltsalt")

	hash1 := pond.HashPassword("password1", salt)
	hash2 := pond.HashPassword("password2", salt)

	if bytes.Equal(hash1, hash2) {
		t.Error("expected different hashes for different passwords, got equal hashes")
	}
}

func TestHashPassword_KnownInputs_ReturnsExpectedArgon2idHash(t *testing.T) {
	password := "secret"
	salt := []byte("saltsaltsaltsalt") // 16 bytes

	got := pond.HashPassword(password, salt)

	want := argon2.IDKey([]byte(password), salt, 2, 19456, 1, 32)
	if !bytes.Equal(got, want) {
		t.Errorf("HashPassword returned unexpected hash\ngot:  %x\nwant: %x", got, want)
	}
}
