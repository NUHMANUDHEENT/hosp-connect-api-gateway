package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/nuhmanudheent/hosp-connect-api-gateway/internal/utils"
)

// JWTMiddleware checks if the request has a valid JWT token and the correct role
func JWTMiddleware(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(role + "token")
			if err != nil {
				if err == http.ErrNoCookie {
					utils.JSONStandardResponse(w, "fail", "Unauthorized", "", http.StatusBadRequest, r)
					return
				}
				utils.JSONStandardResponse(w, "fail", "Bad Request", "", http.StatusBadRequest, r)
				return
			}

			// Get token from cookie
			tokenStr := c.Value

			// Parse the token
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				return jwtKey, nil
			})

			if err != nil || !token.Valid {
				utils.JSONStandardResponse(w, "fail", "Unauthorized", "", http.StatusBadRequest, r)
				return
			}

			// Check if the role matches
			if claims.Role != role {
				utils.JSONStandardResponse(w, "fail", "Forbidden", "", http.StatusBadRequest, r)
				return
			}

			// Token is valid and role matches, pass to the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// VerifyToken parses the JWT token string and returns the claims
func VerifyToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// ExtractClaimsFromCookie extracts the JWT token from the cookie, parses it, and returns the claims
func ExtractClaimsFromCookie(r *http.Request, role string) (*Claims, error) {
	cookie, err := r.Cookie(role + "token")
	if err != nil {
		return nil, fmt.Errorf("could not find cookie: %v", err)
	}

	tokenString := cookie.Value

	claims, err := VerifyToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	return claims, nil
}
