package main

import (
	"bytes"
	"net/http"
)

type CacheResponseWriter struct {
	StatusCode     int
	ResponseHeader http.Header
	Body           bytes.Buffer
}

func (c *CacheResponseWriter) WriteHeader(statusCode int) {
	c.StatusCode = statusCode
}

func (c *CacheResponseWriter) Header() http.Header {
	return c.ResponseHeader
}

func (c *CacheResponseWriter) Write(b []byte) (int, error) {
	return c.Body.Write(b)
}

func (s ProxyServer) CacheMiddlware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if s.cache.cache == nil {
				s.logger.Warn("cache is nil, no responses will be cached!")
				next.ServeHTTP(w, r)
				return
			}
			if r.Method != "GET" {
				s.logger.Info("forwarding non-GET url")
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("X-Cache", "HIT")
			cacheResponse, err := s.cache.Get(r.URL.String())
			if err != nil {
				// response := s.ForwardRequest(r.URL.Path, r.Method)
				tempWriter := CacheResponseWriter{
					ResponseHeader: http.Header{},
					Body:           bytes.Buffer{},
				}
				next.ServeHTTP(&tempWriter, r)
				var err error
				cacheResponse, err = s.cache.Add(r.URL.String(), &tempWriter)
				if err != nil {
					s.logger.Error("could not add response to cache", "error", err)
				}
				w.Header().Set("X-Cache", "MISS")
			}
			copyHeader(w.Header(), cacheResponse.Header)
			w.WriteHeader(cacheResponse.StatusCode)
			_, err = w.Write(cacheResponse.Body)
			if err != nil {
				s.logger.Error("error getting the hit", "error", err)
			}
		},
	)
}
