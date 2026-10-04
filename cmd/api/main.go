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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret != "" {
		auth.SetJWTSecret(jwtSecret)
		log.Println("JWT_SECRET configured")
	}

	var database = dbpkg.MustOpenFromEnv()

	if database != nil {
		defer database.Close()

		if err := dbpkg.InitSchema(database); err != nil {
			log.Fatalf("failed to initialize database schema: %v", err)
		}

		log.Println("database schema initialized")
	}

	server := api.NewServer(database)
	mux := http.NewServeMux()

	// Static files
	mux.Handle("/", api.StaticHandler())

	// Public health endpoints
	mux.HandleFunc("/api/v1/health", server.HealthCheck)
	mux.HandleFunc("/api/v1/ready", server.ReadyCheck)

	// Wallet authentication
	mux.HandleFunc(
		"/api/v1/auth/wallet/challenge",
		server.HandleWalletChallenge,
	)
	mux.HandleFunc(
		"/api/v1/auth/wallet/verify",
		server.HandleWalletVerify,
	)

	// Core endpoints
	mux.HandleFunc("/api/v1/accounts", server.HandleAccounts)
	mux.HandleFunc("/api/v1/transactions", server.HandleTransactions)

	// Authenticated KYC endpoints
	mux.Handle(
		"/api/v1/kyc/submit",
		auth.Middleware(http.HandlerFunc(server.HandleKYCSubmit)),
	)
	mux.Handle(
		"/api/v1/kyc/status",
		auth.Middleware(http.HandlerFunc(server.HandleKYCStatus)),
	)

	// Admin KYC endpoints
	mux.Handle(
		"/api/v1/admin/kyc/pending",
		auth.AdminOnly(http.HandlerFunc(server.HandleAdminPendingKYC)),
	)
	mux.Handle(
		"/api/v1/admin/kyc/review",
		auth.AdminOnly(http.HandlerFunc(server.HandleAdminReviewKYC)),
	)

	// Bank card endpoints
	mux.Handle(
		"/api/v1/cards",
		auth.Middleware(http.HandlerFunc(server.HandleCards)),
	)
	mux.Handle(
		"/api/v1/cards/",
		auth.Middleware(http.HandlerFunc(server.HandleCardItem)),
	)

	handler := api.RateLimit(
		api.Recovery(
			api.SecurityHeaders(mux),
		),
	)

	log.Printf("BRICS Pay API listening on port %s", port)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
