package main

import (
	"log"
	"net/http"
	"os"

	"bricspay/internal/api"
	"bricspay/internal/auth"
	dbpkg "bricspay/internal/db"

	_ "github.com/lib/pq"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// در v12 اگر JWT_SECRET خالی باشد SetJWTSecret ممکن است panic کند.
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set")
	}
	auth.SetJWTSecret(jwtSecret)

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	database, err := dbpkg.Open(dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := dbpkg.InitSchema(database); err != nil {
		log.Fatalf("failed to initialize schema: %v", err)
	}

	server := api.NewServer(database)
	mux := http.NewServeMux()

	// Static
	mux.Handle("/", api.StaticHandler())

	// Health / readiness
	mux.HandleFunc("/api/v1/health", server.HealthCheck)
	mux.HandleFunc("/api/v1/ready", server.ReadyCheck)

	// Wallet auth (جایگزین register/login)
	mux.HandleFunc("/api/v1/auth/wallet/challenge", server.HandleWalletChallenge)
	mux.HandleFunc("/api/v1/auth/wallet/verify", server.HandleWalletVerify)

	// Core
	mux.HandleFunc("/api/v1/accounts", server.HandleAccounts)
	mux.HandleFunc("/api/v1/transactions", server.HandleTransactions)

	// KYC (authenticated)
	mux.Handle("/api/v1/kyc/submit", auth.Middleware(http.HandlerFunc(server.HandleKYCSubmit)))
	mux.Handle("/api/v1/kyc/status", auth.Middleware(http.HandlerFunc(server.HandleKYCStatus)))

	// Admin KYC (admin only)
	mux.Handle("/api/v1/admin/kyc/pending", auth.AdminOnly(http.HandlerFunc(server.HandleAdminPendingKYC)))
	mux.Handle("/api/v1/admin/kyc/review", auth.AdminOnly(http.HandlerFunc(server.HandleAdminReviewKYC)))

	// Cards (authenticated)
	mux.Handle("/api/v1/cards", auth.Middleware(http.HandlerFunc(server.HandleCards)))
	mux.Handle("/api/v1/cards/", auth.Middleware(http.HandlerFunc(server.HandleCardItem)))

	handler := api.RateLimit(api.Recovery(api.SecurityHeaders(mux)))

	log.Printf("BRICSPAY listening on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
