package middlewares

import (
	"errors"
	"net/http"

	"github.com/GTedya/shortener/internal/app/tokenutils"
)

// AuthCheck представляет middleware для проверки авторизации пользователя.
// Если токен пользователя отсутствует в запросе, возвращает статус http.StatusUnauthorized.
// В противном случае передает запрос следующему обработчику.
func (m Middleware) AuthCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie(tokenutils.UserIDCookieName)
		if errors.Is(err, http.ErrNoCookie) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
