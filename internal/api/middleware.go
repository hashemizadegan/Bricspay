package api

import (
	"bytes"
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

// Embed both application assets and multilingual landing pages.
//
// Directory structure:
//
// internal/api/
// ├── middleware.go
// ├── static/
// │   ├── index.html
// │   ├── register.html
// │   ├── css/
// │   └── js/
// └── web/
//     ├── fa/index.html
//     ├── en/index.html
//     ├── ru/index.html
//     └── zh/index.html
//
//go:embed static web
var staticFS embed.FS

// StaticHandler serves the landing pages, registration page,
// login page and static assets.
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

	staticHandler := http.FileServer(http.FS(staticRoot))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath := path.Clean("/" + r.URL.Path)

		// Remove a trailing slash except for the root path.
		if len(requestPath) > 1 {
			requestPath = strings.TrimRight(requestPath, "/")
		}

		switch requestPath {
		// ------------------------------------------------------------
		// Main landing page
		// ------------------------------------------------------------
		case "/":
			serveEmbeddedHTML(w, r, "web/fa/index.html")
			return

		// ------------------------------------------------------------
		// Multilingual landing pages
		// ------------------------------------------------------------
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

		// ------------------------------------------------------------
		// Registration routes
		// ------------------------------------------------------------
		// All of these URLs serve the same embedded registration page.
		case "/app/register", "/register", "/signup":
			serveEmbeddedHTML(w, r, "static/register.html")
			return

		// ------------------------------------------------------------
		// Login routes
		// ------------------------------------------------------------
		// The current project does not contain a separate login.html.
		// Therefore the existing application page is served here.
		case "/app/login", "/login":
			serveEmbeddedHTML(w, r, "static/index.html")
			return

		// ------------------------------------------------------------
		// Static assets
		// ------------------------------------------------------------
		// Support URLs such as:
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

			staticHandler.ServeHTTP(w, rewrittenRequest)
			return
		}

		// Also support direct asset URLs such as:
		// /css/styles.css
		// /js/auth.js
		// /vtb-workflow.jpeg
		//
		// Existing HTML files may use either form.
		staticHandler.ServeHTTP(w, r)
	})
}

// serveEmbeddedHTML reads an HTML file from the embedded filesystem
// and writes it to the HTTP response.
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

// SecurityHeaders adds security-related HTTP headers.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME-type sniffing.
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent the application from being embedded in an iframe.
		w.Header().Set("X-Frame-Options", "DENY")

		// Enable browser XSS protection where supported.
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Limit referrer information.
		w.Header().Set(
			"Referrer-Policy",
			"strict-origin-when-cross-origin",
		)

		// Allow the inline CSS and JavaScript used by the existing
		// landing and registration pages.
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

// Recovery prevents a panic in one request from terminating the server.
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

// RateLimit implements a lightweight in-memory per-IP rate limiter.
//
// This limiter is intentionally fail-open: if the limiter encounters
// an internal problem, the request is still passed to the application.
func RateLimit(next http.Handler) http.Handler {
	var (
		mu      sync.Mutex
		clients = make(map[string]*rateLimitClient)
	)

	const (
		windowDuration = time.Minute
		maxRequests    = 120
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

// rateLimitClient stores request information for one client IP.
type rateLimitClient struct {
	windowStart time.Time
	requests    int
}

// requestIP extracts the best available client IP.
//
// Railway and reverse proxies may provide the original IP through
// X-Forwarded-For or X-Real-IP.
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

// Keep bytes imported for compatibility with projects that previously
// used an embedded content buffer in this middleware file.
var _ = bytes.NewReader
