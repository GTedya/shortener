package middlewares

import (
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
	// SecretKey представляет секретный ключ для подписи токена JWT.
	SecretKey = "some_key"
	// TokenExp представляет срок действия токена JWT.
	TokenExp = 3 * time.Hour
)

// TokenCreate представляет middleware для создания и добавления токена JWT в куку.
// Созданный токен добавляется в куку и сохраняется в контексте запроса.
func (m Middleware) TokenCreate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
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
			w.Header().Add("Authorization", "Bearer "+tokenString)
			r.Header.Add("Authorization", "Bearer "+tokenString)
		}
		next.ServeHTTP(w, r)
	})
}
