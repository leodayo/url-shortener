package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/leodayo/url-shortener/internal/compression/gzip"
	"github.com/leodayo/url-shortener/internal/logger"
	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size

	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		timeTaken := time.Since(start)

		logger.Log.Info("Got incoming HTTP request",
			zap.String("method", r.Method),
			zap.String("URI", r.RequestURI),
			zap.Duration("took", timeTaken),
		)
	})
}

func ResponseLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   &responseData{},
		}
		h.ServeHTTP(&lw, r)

		logger.Log.Info("Response sent",
			zap.Int("status", lw.responseData.status),
			zap.Int("size", lw.responseData.size),
		)
	})
}

// Increases Content-Length for small responses
// [i.e. for /api/shorten which on success returns ~49 bytes without compression tends to return ~73 bytes with gzip applied]
// TODO:
// -> consider making 'compressWriter' write into a buffer first so that the response could be analysed after h.ServeHTTP(wrappedWriter, r)
// -> apply gzip compression only if content-type is either 'application/json' or 'text/html' and if content length is > ~1000-1500 bytes.
func GzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrappedWriter := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		if strings.Contains(acceptEncoding, "gzip") {
			cw := gzip.NewCompressWriter(w)
			defer cw.Close()

			wrappedWriter = cw
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		if contentEncoding == "gzip" {
			cr, err := gzip.NewCompressReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer cr.Close()
			r.Body = cr
		}

		h.ServeHTTP(wrappedWriter, r)
	})
}
