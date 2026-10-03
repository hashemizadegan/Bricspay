package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"bricspay/internal/api"
	"bricspay/internal/auth"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not configured")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret != "" {
		auth.SetJWTSecret(jwtSecret)
	}

	database, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()

	database.SetMaxOpenConns(25)
	database.SetMaxIdleConns(10)
	database.SetConnMaxLifetime(5 * time.Minute)

	pingContext, cancelPing := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancelPing()

	if err := database.PingContext(pingContext); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	// نسخه واقعی v11 فقط دیتابیس را به NewServer می‌دهد.
	server := api.NewServer(database)

	mux := http.NewServeMux()

	// فایل‌های استاتیک و رابط کاربری
	mux.Handle("/", api.StaticHandler())

	// Health check برای Railway
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"version": "11.0.0",
		})
	})

	// وضعیت KYC با احراز هویت JWT
	mux.Handle(
		"/api/kyc/status",
		auth.Middleware(http.HandlerFunc(server.HandleKYCStatus)),
	)

	// Wallet authentication
	mux.HandleFunc(
		"/api/wallet/nonce",
		server.HandleWalletNonce,
	)

	mux.HandleFunc(
		"/api/wallet/verify",
		server.HandleWalletVerify,
	)

	// میان‌افزارهای موجود در نسخه v11
	var handler http.Handler = mux
	handler = api.RateLimit(handler)
	handler = api.Recovery(handler)
	handler = api.SecurityHeaders(handler)

	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf(
			"BRICS Pay v11.0.0 listening on port %s",
			port,
		)

		if err := httpServer.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(
		signalChannel,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	select {
	case err := <-serverErrors:
		log.Fatalf("HTTP server error: %v", err)

	case signalValue := <-signalChannel:
		log.Printf("shutdown signal received: %s", signalValue)
	}

	shutdownContext, cancelShutdown := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancelShutdown()

	if err := httpServer.Shutdown(shutdownContext); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}

	log.Println("server stopped")
}
