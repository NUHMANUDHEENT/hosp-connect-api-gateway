package middleware

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("my_secret_key")

type Claims struct {
	UserId string `json:"id"`
	Role  string `json:"role"`
	jwt.StandardClaims
}

// CreateJWTToken generates a new JWT token for the user with a role
func CreateJWTToken(UserId string, role string) (string, error) {
	expirationTime := time.Now().Add(2 * time.Hour)
	claims := &Claims{
		UserId:   UserId,
		Role: role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func SetJWTToken(w http.ResponseWriter, token string, role string) {
	http.SetCookie(w, &http.Cookie{
		Name:     role + "token",
		Path:     "/",
		Value:    token,
		Expires:  time.Now().Add(2 * time.Hour),
		HttpOnly: true,
	})
}
