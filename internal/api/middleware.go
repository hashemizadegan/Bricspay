package api

import (
	"embed"
	"io/fs"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

//go:embed static
var staticFS embed.FS

// StaticHandler فایلهای استاتیک را از داخل باینری سرو میکند (نه از CWD).
func StaticHandler() http.Handler {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(sub))
}

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy",
    "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; frame-ancestors 'none'; object-src 'none'")

		next.ServeHTTP(w, r)
	})
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered: %v", rec)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RateLimit محدودکنندهٔ سادهٔ درونحافظه است؛ برای چند-اینستنس از Redis استفاده کن.
func RateLimit(next http.Handler) http.Handler {
	type bucket struct {
		count int
		reset time.Time
	}
	var mu sync.Mutex
	buckets := map[string]*bucket{}
	const limit = 30
	const window = time.Minute

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		now := time.Now()

		mu.Lock()
		b, ok := buckets[host]
		if !ok || now.After(b.reset) {
			b = &bucket{reset: now.Add(window)}
			buckets[host] = b
		}
		b.count++
		over := b.count > limit
		mu.Unlock()

		if over {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// writeJSONError از writeJSON موجود در kyc_handlers.go استفاده میکند.
// نام یکتا انتخاب شد تا خطای "writeJSON redeclared" تکرار نشود.
func writeJSONError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
