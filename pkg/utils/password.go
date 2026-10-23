package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

func VerfiyPassword(password, encodedHash string) error {
	parts := strings.Split(encodedHash, ".")
	if len(parts) != 2 {
		return ErrorHandler(errors.New("invalid encoded hash format"), "Internal server error")
		// http.Error(w, "invalid encoded format", http.StatusForbidden)
		// return true

	}
	saltBase64 := parts[0]
	hashedPasswordBase64 := parts[1]
	salt, err := base64.StdEncoding.DecodeString(saltBase64)
	if err != nil {
		return ErrorHandler(err, "Internal Server error")
		// http.Error(w, "failed to decode the slat", http.StatusForbidden)
		// return true
	}
	hashedPassword, err := base64.StdEncoding.DecodeString(hashedPasswordBase64)
	if err != nil {
		return ErrorHandler(err, "Internal server error")
		// http.Error(w, "failed to decode the hashed password", http.StatusForbidden)
		// return true
	}
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	if len(hash) != len(hashedPassword) {
		return ErrorHandler(errors.New("incorrect password"), "incorrect password")

	}
	if subtle.ConstantTimeCompare(hash, hashedPassword) == 1 {
		return nil
		//do nothing
	}
	return ErrorHandler(errors.New("incorrect password"), "incorrect password")

}

func HashPassword(Password string) (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", ErrorHandler(errors.New("failed to the generate salt"), "error adding data")

	}
	hash := argon2.IDKey([]byte(Password), salt, 1, 64*1024, 4, 32)
	saltBase64 := base64.StdEncoding.EncodeToString(salt)
	hashBase64 := base64.StdEncoding.EncodeToString(hash)
	encodedHash := fmt.Sprintf("%s.%s", saltBase64, hashBase64)
	return encodedHash, nil
}
