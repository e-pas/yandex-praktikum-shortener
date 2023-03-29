package config

import (
	"compress/gzip"
	"io"
	"log"
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

type gzipRead struct {
	io.Reader
	gzReader io.Reader
}

func (gzr gzipRead) Read(buf []byte) (n int, err error) {
	return gzr.gzReader.Read(buf)
}

func GzipResponse(next http.Handler) http.Handler {
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

func GunzipRequest(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gzr, err := gzip.NewReader(r.Body)
		if err != nil {
			log.Println(ErrInvalidGZip)
		}
		defer gzr.Close()
		gzbody := gzipRead{
			gzReader: gzr,
		}
		r.Body = io.NopCloser(gzbody)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)

}
