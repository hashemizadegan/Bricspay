package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bricspay/internal/api"
	"bricspay/internal/db"

	_ "github.com/lib/pq"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	// db.New returns *db.DB, which embeds *sql.DB.
	database, err := db.New(databaseURL)
	if err != nil {
		log.Fatalf("database initialization failed: %v", err)
	}
	defer database.Close()

	// api.NewServer accepts the embedded *sql.DB.
	// It creates and manages the ledger service internally.
	serverAPI := api.NewServer(database.DB)

	mux := http.NewServeMux()

	// Health and readiness endpoints.
	mux.HandleFunc("/health", serverAPI.HealthCheck)
	mux.HandleFunc("/ready", serverAPI.ReadyCheck)

	// Ledger endpoints.
	mux.HandleFunc("/accounts", serverAPI.HandleAccounts)
	mux.HandleFunc("/transactions", serverAPI.HandleTransactions)

	// Authentication endpoints.
	mux.HandleFunc("/api/v1/auth/register", serverAPI.HandleRegister)
	mux.HandleFunc("/api/v1/auth/login", serverAPI.HandleLogin)

	// KYC endpoints.
	mux.HandleFunc("/api/v1/kyc/upload", serverAPI.HandleKYCUpload)

	// Admin and compliance endpoints.
	mux.HandleFunc("/api/v1/admin/kyc/list", serverAPI.HandleAdminProfiles)
	mux.HandleFunc("/api/v1/admin/kyc/decide", serverAPI.HandleAdminDecision)
	mux.HandleFunc("/api/v1/admin/audit/list", serverAPI.HandleAdminAudit)

	// Wallet challenge/verification endpoints.
	mux.HandleFunc("/api/v1/wallet/challenge", serverAPI.HandleWalletChallenge)
	mux.HandleFunc("/api/v1/wallet/verify", serverAPI.HandleWalletVerify)

	// Static assets.
	//
	// Browser URL:
	//   /static/css/styles.css
	//
	// Actual file:
	//   internal/api/static/css/styles.css
	staticFiles := http.FileServer(
		http.Dir("internal/api/static"),
	)

	mux.Handle(
		"/static/",
		http.StripPrefix("/static/", staticFiles),
	)

	// Serve the application entry point.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		http.ServeFile(
			w,
			r,
			"internal/api/static/index.html",
		)
	})

	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           securityHeaders(mux),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	shutdownDone := make(chan struct{})

	go func() {
		defer close(shutdownDone)

		log.Printf("BRICS Pay API listening on :%s", port)

		if err := httpServer.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
	)

	<-stop

	log.Println("Shutdown signal received")

	shutdownContext, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := httpServer.Shutdown(shutdownContext); err != nil {
		log.Printf("HTTP graceful shutdown failed: %v", err)
	}

	<-shutdownDone

	log.Println("BRICS Pay API stopped")
}

// securityHeaders adds basic browser security headers.
// Authentication and authorization must still be enforced
// inside the protected handlers or middleware.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set(
			"Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data:; "+
				"font-src 'self' data:; "+
				"connect-src 'self'; "+
				"frame-ancestors 'none'",
		)

		next.ServeHTTP(w, r)
	})
}
