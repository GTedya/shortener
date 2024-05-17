package middlewares

import (
	"github.com/stretchr/testify/assert"

	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthCheck(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/api/user/urls", nil)
	assert.NoError(t, err)

	req.AddCookie(&http.Cookie{Name: "token", Value: "fake_token"})

	middleware := Middleware{}
	authMiddleware := middleware.AuthCheck(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	reqWithoutToken, err := http.NewRequest(http.MethodGet, "/api/user/urls", nil)
	assert.NoError(t, err)

	wWithoutToken := httptest.NewRecorder()

	authMiddleware.ServeHTTP(wWithoutToken, reqWithoutToken)

	assert.Equal(t, http.StatusUnauthorized, wWithoutToken.Code, "Статус код ответа должен быть 401 (Unauthorized)")
}
