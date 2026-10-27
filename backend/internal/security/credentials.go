package security

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

func RandomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func ValidUsername(s string) bool {
	return regexp.MustCompile(`^[\p{L}\p{N}_\-.]{2,32}$`).MatchString(s)
}
func ValidPassword(s string) bool { return len(s) >= 6 && len(s) <= 72 }
func HashPassword(s string) string {
	b, _ := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	return string(b)
}
