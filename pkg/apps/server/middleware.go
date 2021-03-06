package server

import (
	"highload_hw3/pkg/responses"
	"highload_hw3/pkg/session"
	"context"
	"net/http"

	"github.com/google/uuid"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", config.AllowedIP)
			w.Header().Set("Access-Control-Allow-Methods", config.AllowedMethods)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("X-Vasily", "58")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
}

// re-implementation of ResponseWriter to access status code
type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusWriter) WriteHeader(code int) {
	if code != 0 {
		w.statusCode = code
	} else {
		w.statusCode = http.StatusOK
	}

	w.ResponseWriter.WriteHeader(code)
}

func (srv *Server) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writter := statusWriter{w, http.StatusOK}
		next.ServeHTTP(&writter, r)
	})
}

func (srv *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var isAuth bool
			var sess *session.Session
			var id uuid.UUID
			ctx := r.Context()
			value, err := responses.GetValueFromCookie(r, "sessionID")
			if err == http.ErrNoCookie {
				isAuth = false
			} else {
				id, err = uuid.Parse(value)
				if err != nil {
					srv.log.Warnln("can't parse ")
					ctx = context.WithValue(ctx, "isAuth", false)
					next.ServeHTTP(w, r.WithContext(ctx))
				}
				sess, err = srv.sm.Check(ctx, &session.SessionID{ID: id.String()})
				if err != nil {
					srv.log.Println("can't check session ID: ", err)
				}
				if sess != nil {
					isAuth = true
					ctx = context.WithValue(ctx, "userID", sess.ID)
				}
			}
			ctx = context.WithValue(ctx, "sessionID", id.String())
			ctx = context.WithValue(ctx, "isAuth", isAuth)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
}

func (srv *Server) logginigMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			srv.log.Infoln(r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
}

func (srv *Server) authRequierMiddleware(next http.Handler) http.Handler {
	srv.log.Println("auth requier miiddleware")
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if !getIsAuth(r) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
}

func getIsAuth(r *http.Request) bool {
	return r.Context().Value("isAuth").(bool)
}

func getUserID(r *http.Request) int32 {
	return r.Context().Value("userID").(int32)
}

func getSessionID(r *http.Request) string {
	return r.Context().Value("sessionID").(string)
}
