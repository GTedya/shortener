package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
}

type ContextKey string

const (
	TokenContextKey ContextKey = "token"
	SecretKey                  = "some_key"
	TokenExp                   = 3 * time.Hour
	tokenCookie                = "token"
)

func (m Middleware) TokenCreate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var cooks *http.Cookie

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
			},
		})

		tokenString, err := token.SignedString([]byte(SecretKey))
		if err != nil {
			m.Log.Errorf("token signed error: %w", err)
			return
		}

		cooks = &http.Cookie{Name: tokenCookie, Value: tokenString}
		http.SetCookie(w, cooks)

		ctx := context.WithValue(r.Context(), TokenContextKey, cooks.Value)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func TokenCreate() (*http.Cookie, error) {
	var cooks *http.Cookie
	var err error

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return nil, fmt.Errorf("token signed error: %w", err)
	}

	cooks = &http.Cookie{Name: tokenCookie, Value: tokenString}

	return cooks, nil
}
