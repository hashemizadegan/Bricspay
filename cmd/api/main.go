package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bricspay/internal/api"
	"bricspay/internal/auth"
	"bricspay/internal/db"
	"bricspay/internal/ledger"
)

//go:embed static/*
var staticFS embed.FS

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Println("WARNING: DATABASE_URL not set, operating with fallback configuration")
	}

	// 1. Initialize Database & Migrations
	database, err := db.InitDB(databaseURL)
	if err != nil {
		log.Printf("Database initialization warning: %v", err)
	} else {
		defer database.Close()
		if err := db.RunKYCMigrations(database); err != nil {
			log.Printf("KYC Schema Migration warning: %v", err)
		}
	}

	// 2. Initialize Core Services
	ledgerService := ledger.NewService(database)
	authService := auth.NewService(os.Getenv("JWT_SECRET"))

	// 3. Initialize API Server
	server := api.NewServer(database, ledgerService, authService, staticFS)

	// 4. Setup Router & Routes
	mux := http.NewServeMux()

	// Static Assets & UI
	mux.Handle("/", server.StaticFileServer())

	// Core API Gateway
	mux.HandleFunc("/api", server.HandleRoot)
	mux.HandleFunc("/health", server.HealthCheck)

	// Ledger API Routes
	mux.HandleFunc("/api/v1/accounts", server.HandleAccounts)
	mux.HandleFunc("/api/v1/transactions", server.HandleTransactions)
	mux.HandleFunc("/api/v1/banks", server.HandleBanks)

	// Auth & KYC API Routes
	mux.HandleFunc("/api/v1/auth/login", server.HandleLogin)
	mux.HandleFunc("/api/v1/auth/register", server.HandleRegister)
	mux.HandleFunc("/api/v1/kyc/submit", server.AuthMiddleware(server.HandleKYCSubmit))
	mux.HandleFunc("/api/v1/kyc/status", server.AuthMiddleware(server.HandleKYCStatus))
	mux.HandleFunc("/api/v1/admin/kyc/requests", server.AdminMiddleware(server.HandleAdminKYCList))
	mux.HandleFunc("/api/v1/admin/kyc/approve", server.AdminMiddleware(server.HandleAdminKYCApprove))

	// CORS & Security Handler Wrapper
	handler := server.CorsMiddleware(mux)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful Shutdown Setup
	go func() {
		log.Printf("BRICS Pay Settlement Engine running on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down BRICS Pay Settlement Engine gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server stopped successfully.")
}
