package config

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipWrite struct {
	http.ResponseWriter
	gzWriter io.Writer
}

func (gz gzipWrite) WriteHeader(code int) {
	gz.ResponseWriter.WriteHeader(code)
}

func (gz gzipWrite) Write(buf []byte) (int, error) {
	return gz.gzWriter.Write(buf)
}

func GzipHandle(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gzw := gzip.NewWriter(w)
		defer gzw.Close()
		gz := gzipWrite{
			ResponseWriter: w,
			gzWriter:       gzw,
		}

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gz, r)
	}
	return http.HandlerFunc(fn)
}
