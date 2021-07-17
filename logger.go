package chilogrus

import (
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

const timeFormat = "02/Jan/2006:15:04:05 -0700"

// Logger is the logrus logger handler
func Logger(log logrus.FieldLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if r.URL.RawQuery != "" {
				path += "?" + r.URL.RawQuery
			}
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			stop := time.Since(start)
			statusCode := ww.Status()
			clientIP := r.RemoteAddr
			clientUserAgent := r.UserAgent()
			if clientUserAgent == "" {
				clientUserAgent = "-"
			}
			referer := r.Referer()
			if referer == "" {
				referer = "-"
			}
			dataLength := ww.BytesWritten()
			if dataLength < 0 {
				dataLength = 0
			}

			entry := log.WithFields(logrus.Fields{
				"statusCode":     statusCode,
				"duration":       stop.Nanoseconds(), // in nanoseconds
				"durationPretty": stop.String(),
				"clientIP":       clientIP,
				"method":         r.Method,
				"path":           path,
				"proto":          r.Proto,
				"referer":        referer,
				"dataLength":     dataLength,
				"userAgent":      clientUserAgent,
			})

			msg := fmt.Sprintf(
				"%s - - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\" %s",
				clientIP,
				time.Now().Format(timeFormat),
				r.Method,
				path,
				r.Proto,
				statusCode,
				dataLength,
				referer,
				clientUserAgent,
				stop.String(),
			)
			if statusCode >= http.StatusInternalServerError {
				entry.Error(msg)
			} else if statusCode >= http.StatusBadRequest {
				entry.Warn(msg)
			} else {
				entry.Info(msg)
			}
		})
	}
}
