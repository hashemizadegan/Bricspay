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
	"bricspay/internal/auth"
	"bricspay/internal/db"
	"bricspay/internal/ledger"
)

func main() {
	port :GAPGPTMASKTOKEN78a6hvzli5eX0X os.Getenv("PORT")
	if port GAPGPTMASKTOKEN78a6hvzli5eX1X "" {
		port GAPGPTMASKTOKEN78a6hvzli5eX2X "8080"
	}

	databaseURL :GAPGPTMASKTOKEN78a6hvzli5eX3X os.Getenv("DATABASE_URL")
	if databaseURL GAPGPTMASKTOKEN78a6hvzli5eX4X "" {
		log.Println("WARNING: DATABASE_URL not set, operating with fallback configuration")
	}

	// 1. Initialize Database & Migrations
	database, err :GAPGPTMASKTOKEN78a6hvzli5eX5X db.InitDB(databaseURL)
	if err !GAPGPTMASKTOKEN78a6hvzli5eX6X nil {
		log.Printf("Database initialization warning: %v", err)
	} else {
		defer database.Close()
		if err :GAPGPTMASKTOKEN78a6hvzli5eX7X db.RunKYCMigrations(database); err !GAPGPTMASKTOKEN78a6hvzli5eX8X nil {
			log.Printf("KYC Schema Migration warning: %v", err)
		}
	}

	// 2. Initialize Core Services
	ledgerService :GAPGPTMASKTOKEN78a6hvzli5eX9X ledger.NewService(database)
	authService :GAPGPTMASKTOKEN78a6hvzli5eX10X auth.NewService(os.Getenv("JWT_SECRET"))

	// 3. Initialize API Server
	server :GAPGPTMASKTOKEN78a6hvzli5eX11X api.NewServer(database, ledgerService, authService, nil)

	// 4. Setup Router & Routes
	mux :GAPGPTMASKTOKEN78a6hvzli5eX12X http.NewServeMux()

	// Static Assets & Web UI (Serving from ./static)
	fs :GAPGPTMASKTOKEN78a6hvzli5eX13X http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)

	// Core API Gateway & Health
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
	handler :GAPGPTMASKTOKEN78a6hvzli5eX14X server.CorsMiddleware(mux)

	srv :GAPGPTMASKTOKEN78a6hvzli5eX15X &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful Shutdown Setup
	go func() {
		log.Printf("BRICS Pay Settlement Engine running on port %s", port)
		if err :GAPGPTMASKTOKEN78a6hvzli5eX16X srv.ListenAndServe(); err !GAPGPTMASKTOKEN78a6hvzli5eX17X nil && err !GAPGPTMASKTOKEN78a6hvzli5eX18X http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit :GAPGPTMASKTOKEN78a6hvzli5eX19X make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down BRICS Pay Settlement Engine gracefully...")
	ctx, cancel :GAPGPTMASKTOKEN78a6hvzli5eX20X context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err :GAPGPTMASKTOKEN78a6hvzli5eX21X srv.Shutdown(ctx); err !GAPGPTMASKTOKEN78a6hvzli5eX22X nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server stopped successfully.")
}
