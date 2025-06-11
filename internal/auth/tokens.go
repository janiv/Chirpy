package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	issued_at := time.Now()
	expires_at := issued_at.Add(expiresIn)
	claims := &jwt.RegisteredClaims{
		Issuer:    "chirpy",
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(issued_at),
		ExpiresAt: jwt.NewNumericDate(expires_at),
	}

	tk := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := tk.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return ss, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	usr, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}
	usrUUID, err := uuid.Parse(usr)
	if err != nil {
		return uuid.Nil, err
	}
	return usrUUID, nil

}

func GetBearerToken(headers http.Header) (string, error) {
	bearerString := headers.Get("Authorization")
	if len(bearerString) < 1 {
		return "", errors.New("missing auth")
	}
	return strings.TrimPrefix(bearerString, "Bearer "), nil
}

func MakeRefreshToken() (string, error) {
	bytesKey := make([]byte, 64)
	rand.Read(bytesKey)
	stringKey := hex.EncodeToString(bytesKey)
	return stringKey, nil
}
