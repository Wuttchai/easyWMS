package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"strings"
	"time"
)

type Claims struct {
	UserID   string `json:"uid"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Exp      int64  `json:"exp"`
}

func HashPassword(s string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	} // Callers validate the bcrypt length limit before hashing.
	return string(hash)
}

func VerifyPassword(hash, password string) bool {
	if len(password) > 72 {
		return false
	}
	if strings.HasPrefix(hash, "$2") {
		return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
	}
	// Upgrade legacy hashes only after a successful login.
	legacy := sha256.Sum256([]byte(password))
	return len(hash) == 64 && subtle.ConstantTimeCompare([]byte(hash), []byte(fmt.Sprintf("%x", legacy[:]))) == 1
}

func Sign(secret string, c Claims) (string, error) {
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	enc := base64.RawURLEncoding
	body := enc.EncodeToString(header) + "." + enc.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	return body + "." + enc.EncodeToString(mac.Sum(nil)), nil
}

func Parse(secret, token string) (Claims, error) {
	var c Claims
	p := strings.Split(token, ".")
	if len(p) != 3 {
		return c, errors.New("invalid token")
	}
	body := p[0] + "." + p[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	sig, err := base64.RawURLEncoding.DecodeString(p[2])
	if err != nil || !hmac.Equal(sig, mac.Sum(nil)) {
		return c, errors.New("invalid signature")
	}
	b, err := base64.RawURLEncoding.DecodeString(p[1])
	if err != nil {
		return c, err
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return c, err
	}
	if c.Exp <= time.Now().Unix() {
		return c, errors.New("token expired")
	}
	return c, nil
}

func Bearer(v string) string {
	if len(v) > 7 && strings.EqualFold(v[:7], "Bearer ") {
		return strings.TrimSpace(v[7:])
	}
	return ""
}

func ExpiresIn(hours int) int64 { return time.Now().Add(time.Duration(hours) * time.Hour).Unix() }
func Itoa(v int64) string       { return strconv.FormatInt(v, 10) }
