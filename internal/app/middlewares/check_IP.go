package middlewares

import (
	"net"
	"net/http"
)

func (m Middleware) IPCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Real-IP")
		if ip == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if m.TrustedSubnet == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		_, ipnetA, _ := net.ParseCIDR(m.TrustedSubnet)
		ipB, _, _ := net.ParseCIDR(ip)

		if !ipnetA.Contains(ipB) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
