package middlewares

import (
	"net/http"
)

// AuthCheck представляет middleware для проверки авторизации пользователя.
// Если токен пользователя отсутствует в запросе, возвращает статус http.StatusUnauthorized.
// В противном случае передает запрос следующему обработчику.
func (m Middleware) AuthCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		str := r.Header.Get("Authorization")
		if str == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
