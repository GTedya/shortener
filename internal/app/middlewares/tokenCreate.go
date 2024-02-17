package middlewares

import (
	"context"
	"errors"
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
	TokenExp                   = 30 * time.Millisecond
	tokenCookie                = "token"
)

func (m Middleware) TokenCreate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var cooks *http.Cookie

		cooks, err := r.Cookie(tokenCookie)
		if !errors.Is(err, http.ErrNoCookie) && err != nil {
			m.Log.Errorf("cookie receiving error: %w", err)
			return
		}
		if errors.Is(err, http.ErrNoCookie) {
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
		}

		ctx := context.WithValue(r.Context(), TokenContextKey, cooks.Value)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
