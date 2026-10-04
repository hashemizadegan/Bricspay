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

	var database = openDatabase()

	if database != nil {
		defer database.Close()

		if err := dbpkg.InitSchema(database); err != nil {
			log.Fatalf("failed to initialize database schema: %v", err)
		}

		log.Println("database schema initialized")
	}

	server := api.NewServer(database)
	mux := http.NewServeMux()

	mux.Handle("/", api.StaticHandler())

	mux.HandleFunc("/api/v1/health", server.HealthCheck)
	mux.HandleFunc("/api/v1/ready", server.ReadyCheck)

	mux.HandleFunc(
		"/api/v1/auth/wallet/challenge",
		server.HandleWalletChallenge,
	)
	mux.HandleFunc(
		"/api/v1/auth/wallet/verify",
		server.HandleWalletVerify,
	)

	mux.HandleFunc("/api/v1/accounts", server.HandleAccounts)
	mux.HandleFunc("/api/v1/transactions", server.HandleTransactions)

	mux.Handle(
		"/api/v1/kyc/submit",
		auth.Middleware(http.HandlerFunc(server.HandleKYCSubmit)),
	)
	mux.Handle(
		"/api/v1/kyc/status",
		auth.Middleware(http.HandlerFunc(server.HandleKYCStatus)),
	)

	mux.Handle(
		"/api/v1/admin/kyc/pending",
		auth.AdminOnly(http.HandlerFunc(server.HandleAdminPendingKYC)),
	)
	mux.Handle(
		"/api/v1/admin/kyc/review",
		auth.AdminOnly(http.HandlerFunc(server.HandleAdminReviewKYC)),
	)

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

func openDatabase() *sql.DB {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Println("WARNING: DATABASE_URL is not configured")
		return nil
	}

	database, err := dbpkg.Open(databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	return database
}
