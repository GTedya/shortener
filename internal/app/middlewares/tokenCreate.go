package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// Claims представляет пользовательские утверждения для токена JWT.
type Claims struct {
	userID string
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
		claims := &Claims{
			userID: uuid.NewString(),
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		tokenString, err := token.SignedString([]byte(SecretKey))
		if err != nil {
			m.Log.Errorf("token signed error: %w", err)
			return
		}
		w.Header().Set("Authorization", tokenString)
		r.Header.Set("Authorization", tokenString)

		next.ServeHTTP(w, r)
	})
}

func ExtractIDFromToken(requestToken string, secret string) (string, error) {
	token, err := jwt.Parse(requestToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok && !token.Valid {
		return "", fmt.Errorf("Invalid Token")
	}

	return claims["userID"].(string), nil
}
