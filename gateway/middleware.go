package main

import (
	"compress/gzip"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// gzipResponseWriter wraps http.ResponseWriter to add gzip compression
type gzipResponseWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.writer.Write(b)
}

func (g *gzipResponseWriter) Close() error {
	return g.writer.Close()
}

// enableGzip - Compresses responses with gzip (70-90% size reduction)
func enableGzip(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip compression for streaming endpoints
		if strings.Contains(r.URL.Path, "stream") || strings.Contains(r.URL.Path, "bidirectional") {
			next(w, r)
			return
		}

		// Check if client supports gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next(w, r)
			return
		}

		// Create gzip writer
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()

		gzw := &gzipResponseWriter{ResponseWriter: w, writer: gz}
		next(gzw, r)
	}
}

// Rate limiter per IP
type rateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

func newRateLimiter(rps int, burst int) *rateLimiter {
	return &rateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(rps),
		burst:    burst,
	}
}

func (rl *rateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[ip]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = limiter
		rl.mu.Unlock()
	}

	return limiter
}

// Global rate limiter: 100 requests/second, burst of 200
var globalRateLimiter = newRateLimiter(100, 200)

// rateLimitMiddleware - Prevents abuse and ensures fair usage
func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}

		limiter := globalRateLimiter.getLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}

// requestLogger - Logs request timing and status
func requestLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap response writer to capture status code
		lw := &responseLogger{ResponseWriter: w, statusCode: http.StatusOK}
		
		next(lw, r)
		
		duration := time.Since(start)
		log.Printf("[%s] %s %s - %d - %v", r.Method, r.URL.Path, r.RemoteAddr, lw.statusCode, duration)
	}
}

type responseLogger struct {
	http.ResponseWriter
	statusCode int
}

func (rl *responseLogger) WriteHeader(code int) {
	rl.statusCode = code
	rl.ResponseWriter.WriteHeader(code)
}

// Flush implements http.Flusher to support Server-Sent Events
func (rl *responseLogger) Flush() {
	if flusher, ok := rl.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

