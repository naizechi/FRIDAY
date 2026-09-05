package middleware

import (
	"net/http"
	"runtime/debug"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/naizechi/friday/internal/core/logger"
)

func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return chiMiddleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if requestID := chiMiddleware.GetReqID(r.Context()); requestID != "" {
				w.Header().Set(chiMiddleware.RequestIDHeader, requestID)
			}

			next.ServeHTTP(w, r)
		}))
	}
}

func Logger(log *logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := chiMiddleware.GetReqID(r.Context())

			l := log.With(
				"request_id", requestID,
				"url", r.URL.String(),
			)

			next.ServeHTTP(w, r.WithContext(logger.WithContext(r.Context(), l)))
		})
	}
}

func Recoverer() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if p := recover(); p != nil {
					if p == http.ErrAbortHandler {
						panic(p)
					}

					log := logger.FromContext(r.Context())
					log.Error(
						"panic during HTTP request",
						"panic", p,
						"stack", string(debug.Stack()),
					)

					w.WriteHeader(http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func Trace() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := logger.FromContext(r.Context())
			rw := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			started := time.Now()

			log.Debug(">>> incoming HTTP request", "time", started.UTC())

			next.ServeHTTP(rw, r)

			statusCode := rw.Status()
			if statusCode == 0 {
				statusCode = http.StatusOK
			}

			log.Debug(
				"<<< done HTTP request",
				"status_code", statusCode,
				"latency", time.Since(started),
			)
		})
	}
}
