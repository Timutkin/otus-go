package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/urfave/negroni"
)

func httpLoggingMiddleware(h http.Handler, lg Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := negroni.NewResponseWriter(w)
		h.ServeHTTP(lrw, r)
		latency := time.Since(start)
		lg.InfoWithParams("", map[string]string{
			"protocol": "http",
			"method":   r.Method,
			"path":     r.RequestURI,
			"status":   strconv.Itoa(lrw.Status()),
			"latency":  latency.String(),
			"IP":       r.RemoteAddr,
		})
	})
}
