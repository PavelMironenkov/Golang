package middleware

import (
	"hw5/pkg/session"
	"net/http"
)

func Auth(sm *session.SessionsManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := sm.Check(w, r)
		if sess != nil && err == nil {
			ctx := session.ContextWithSession(r.Context(), sess)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
