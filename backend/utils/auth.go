package utils

import (
	"crypto/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
)

func HashPassword(password string) (string, error) {
	salt := generateSalt()
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return encodeHash(salt, hash), nil
}

func CheckPassword(password, encodedHash string) bool {
	salt, hash := decodeHash(encodedHash)
	if salt == nil || hash == nil {
		return false
	}
	newHash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return string(hash) == string(newHash)
}

func generateSalt() []byte {
	salt := make([]byte, 16)
	rand.Read(salt)
	return salt
}

func encodeHash(salt, hash []byte) string {
	encoded := make([]byte, len(salt)+len(hash))
	copy(encoded, salt)
	copy(encoded[len(salt):], hash)
	return string(encoded)
}

func decodeHash(encoded string) ([]byte, []byte) {
	if len(encoded) < 16 {
		return nil, nil
	}
	salt := []byte(encoded[:16])
	hash := []byte(encoded[16:])
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
