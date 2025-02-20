package main

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

var contentTypesForCompress = []string{"application/json", "text/html"}

// compressWriter implements interface http.ResponseWriter
// and allows compressing data and setting proper headers
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close closes gzip.Writer and flushes data from the buffer.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader implements interface io.ReadCloser
// and allows decompressing data.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func gzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// original http.ResponseWriter by default
		ow := w

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") &&
			isSuitedForGzip(r.Header.Get("Content-Type")) {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		h.ServeHTTP(ow, r)
	}
}

func isSuitedForGzip(contentType string) bool {
	for _, acceptEncoding := range contentTypesForCompress {
		if contentType == acceptEncoding {
			return true
		}
	}
	return false
}
