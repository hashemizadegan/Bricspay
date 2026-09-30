package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bricspay/internal/api"
	dbpkg "bricspay/internal/db"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	database, err := dbpkg.InitDB(dbURL)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer database.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := api.NewServer(database)
	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir("internal/api/static"))
	mux.Handle("/", fs)

	// Health
	mux.HandleFunc("/api/v1/health", srv.HealthCheck)

	// Auth
	mux.HandleFunc("/api/v1/auth/register", srv.HandleRegister)
	mux.HandleFunc("/api/v1/auth/login", srv.HandleLogin)

	// KYC (authenticated)
	mux.Handle("/api/v1/kyc/profile", srv.AuthMiddleware(http.HandlerFunc(srv.HandleKYCProfile)))
	mux.Handle("/api/v1/kyc/documents", srv.AuthMiddleware(http.HandlerFunc(srv.HandleKYCUpload)))

	// Admin (authenticated + admin role)
	mux.Handle("/api/v1/admin/kyc/list", srv.AdminMiddleware(http.HandlerFunc(srv.HandleAdminProfiles)))
	mux.Handle("/api/v1/admin/kyc/evaluate", srv.AdminMiddleware(http.HandlerFunc(srv.HandleAdminDecision)))

	// Accounts & Transactions (authenticated)
	mux.Handle("/api/v1/accounts",.HandlerFunc(srv.HandleAdminDecision)))

	// Accounts & Transactions (authenticated)
	mux.Handle("/api/v1/accounts", srv.AuthMiddleware(http.HandlerFunc(srv.HandleAccounts)))
	mux.Handle("/api/v1/transactions", srv.AuthMiddleware(http.HandlerFunc(srv.HandleTransactions)))

	httpServer := &http.Server{IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("server listening on :%s", port)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http. httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
