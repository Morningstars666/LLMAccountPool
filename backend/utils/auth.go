package utils

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
)

func HashPassword(password string) (string, error) {
	salt := generateSalt()
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	combined := make([]byte, len(salt)+len(hash))
	copy(combined, salt)
	copy(combined[len(salt):], hash)

	encodedHash := base64.StdEncoding.EncodeToString(combined)
	return encodedHash, nil
}

func CheckPassword(password, encodedHash string) bool {
	combined, err := base64.StdEncoding.DecodeString(encodedHash)
	if err != nil || len(combined) < 16 {
		return false
	}

	salt := combined[:16]
	hash := combined[16:]

	newHash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return string(hash) == string(newHash)
}

func generateSalt() []byte {
	salt := make([]byte, 16)
	rand.Read(salt)
	return salt
}

func encodeHash(salt, hash []byte) string {
	combined := make([]byte, len(salt)+len(hash))
	copy(combined, salt)
	copy(combined[len(salt):], hash)
	return base64.StdEncoding.EncodeToString(combined)
}

func decodeHash(encoded string) ([]byte, []byte) {
	combined, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(combined) < 16 {
		return nil, nil
	}
	salt := combined[:16]
	hash := combined[16:]
	return salt, hash
}

func GenerateJWT(secret string, userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateJWT(tokenString, secret string) (bool, map[string]interface{}) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return false, nil
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return true, claims
	}
	return false, nil
}
