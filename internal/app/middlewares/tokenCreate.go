package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Claims представляет пользовательские утверждения для токена JWT.
type Claims struct {
	jwt.RegisteredClaims
}

// ContextKey определяет тип для ключа контекста.
type ContextKey string

const (
	// TokenContextKey используется для доступа к токену в контексте запроса.
	TokenContextKey ContextKey = "token"
	// SecretKey представляет секретный ключ для подписи токена JWT.
	SecretKey = "some_key"
	// TokenExp представляет срок действия токена JWT.
	TokenExp = 3 * time.Hour
	// tokenCookie определяет имя куки, в которой хранится токен.
	tokenCookie = "token"
)

// TokenCreate представляет middleware для создания и добавления токена JWT в куку.
// Созданный токен добавляется в куку и сохраняется в контексте запроса.
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

// TokenCreate создает токен JWT и возвращает его в виде куки.
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
