package api

import (
	"embed"
	"io/fs"
	"log"
	"net"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
)

//go:embed static web
var staticFS embed.FS

// StaticHandler serves the embedded application files.
func StaticHandler() http.Handler {
	staticRoot, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Printf("unable to load embedded static directory: %v", err)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(
				w,
				"Embedded static files are unavailable",
				http.StatusInternalServerError,
			)
		})
	}

	staticServer := http.FileServer(http.FS(staticRoot))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath := path.Clean("/" + r.URL.Path)

		if requestPath != "/" {
			requestPath = strings.TrimRight(requestPath, "/")
		}

		switch requestPath {
		case "/":
			serveEmbeddedHTML(w, r, "web/fa/index.html")
			return

		case "/fa":
			serveEmbeddedHTML(w, r, "web/fa/index.html")
			return

		case "/en":
			serveEmbeddedHTML(w, r, "web/en/index.html")
			return

		case "/ru":
			serveEmbeddedHTML(w, r, "web/ru/index.html")
			return

		case "/zh":
			serveEmbeddedHTML(w, r, "web/zh/index.html")
			return

		case "/app/register", "/register", "/signup":
			serveEmbeddedHTML(w, r, "static/register.html")
			return

		case "/app/login", "/login":
			serveEmbeddedHTML(w, r, "static/index.html")
			return

		case "/settings/id_verification", "/id_verification", "/app/id_verification":
			serveEmbeddedHTML(w, r, "static/id_verification.html")
			return

		default:
			// Serve paths such as:
			// /static/css/styles.css
			// /static/js/auth.js
			if strings.HasPrefix(requestPath, "/static/") {
				assetPath := strings.TrimPrefix(requestPath, "/static/")

				if assetPath == "" {
					http.NotFound(w, r)
					return
				}

				rewrittenRequest := r.Clone(r.Context())
				rewrittenRequest.URL.Path = "/" + assetPath

				staticServer.ServeHTTP(w, rewrittenRequest)
				return
			}

			// Serve ordinary assets such as:
			// /css/styles.css
			// /js/auth.js
			// /vtb-workflow.jpeg
			staticServer.ServeHTTP(w, r)
			return
		}
	})
}

// serveEmbeddedHTML serves an HTML file from the embedded filesystem.
func serveEmbeddedHTML(
	w http.ResponseWriter,
	r *http.Request,
	fileName string,
) {
	content, err := fs.ReadFile(staticFS, fileName)
	if err != nil {
		log.Printf("embedded file %q was not found: %v", fileName, err)
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write(content)
}

// SecurityHeaders adds security-related response headers.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set(
			"Referrer-Policy",
			"strict-origin-when-cross-origin",
		)

		w.Header().Set(
			"Content-Security-Policy",
			"default-src 'self'; "+
				"base-uri 'self'; "+
				"object-src 'none'; "+
				"frame-ancestors 'none'; "+
				"img-src 'self' data: blob: https:; "+
				"font-src 'self' data: https:; "+
				"style-src 'self' 'unsafe-inline' https:; "+
				"script-src 'self' 'unsafe-inline' 'unsafe-eval' https:; "+
				"connect-src 'self' https:; "+
				"form-action 'self'",
		)

		next.ServeHTTP(w, r)
	})
}

// Recovery prevents a panic from terminating the server.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf(
					"panic recovered: method=%s path=%s error=%v",
					r.Method,
					r.URL.Path,
					recovered,
				)

				http.Error(
					w,
					"Internal Server Error",
					http.StatusInternalServerError,
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// RateLimit applies a lightweight in-memory per-IP rate limit.
func RateLimit(next http.Handler) http.Handler {
	var (
		mu      sync.Mutex
		clients = make(map[string]*rateLimitClient)
	)

	const (
		windowDuration = time.Minute
		maxRequests    = 30
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := requestIP(r)
		now := time.Now()

		mu.Lock()

		client, exists := clients[clientIP]

		if !exists || now.Sub(client.windowStart) >= windowDuration {
			clients[clientIP] = &rateLimitClient{
				windowStart: now,
				requests:    1,
			}

			mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}

		client.requests++
		requestsExceeded := client.requests > maxRequests

		mu.Unlock()

		if requestsExceeded {
			w.Header().Set("Retry-After", "60")
			http.Error(
				w,
				"Too many requests",
				http.StatusTooManyRequests,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type rateLimitClient struct {
	windowStart time.Time
	requests    int
}

// requestIP returns the best available client IP.
func requestIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		firstIP := strings.TrimSpace(
			strings.Split(forwardedFor, ",")[0],
		)

		if firstIP != "" {
			return firstIP
		}
	}

	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}

	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}

	return "unknown"
}

// writeJSONError is kept for the existing API handlers.
func writeJSONError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	http.Error(w, message, status)
}
