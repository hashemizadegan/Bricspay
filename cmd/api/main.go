package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
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

	// JWT_SECRET: اگر تنظیم نشده بود، یک کلید تصادفی می‌سازیم تا سرویس بالا بیاید.
	// (در این حالت توکن‌ها بعد از هر ری‌استارت نامعتبر می‌شوند؛ بعداً در Railway مقدار دائمی ست کنید.)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err != nil {
			log.Fatalf("failed to generate ephemeral JWT secret: %v", err)
		}
		jwtSecret = hex.EncodeToString(buf)
		log.Println("WARNING: JWT_SECRET is not set; using an ephemeral secret for this boot")
	}
	auth.SetJWTSecret(jwtSecret)

	// DATABASE_URL: اگر نبود یا اتصال/مایگریشن شکست خورد، سرویس همچنان بالا می‌آید
	// و endpointهای وابسته به دیتابیس با خطای 500 (توسط Recovery) پاسخ می‌دهند.
	var database *sql.DB

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("WARNING: DATABASE_URL is not set; running without database")
	} else {
		db, err := dbpkg.Open(dbURL)
		if err != nil {
			log.Printf("WARNING: database connection failed (%v); running without database", err)
		} else if err := dbpkg.InitSchema(db); err != nil {
			log.Printf("WARNING: schema migration failed (%v); running without database", err)
			db.Close()
		} else {
			database = db
			log.Println("database connected and schema initialized")
		}
	}
	if database != nil {
		defer database.Close()
	}

	server := api.NewServer(database)
	mux := http.NewServeMux()

	// Static
	mux.Handle("/", api.StaticHandler())

	// Health / readiness
	mux.HandleFunc("/api/v1/health", server.HealthCheck)
	mux.HandleFunc("/api/v1/ready", server.ReadyCheck)

	// Wallet auth
	mux.HandleFunc("/api/v1/auth/wallet/challenge", server.HandleWalletChallenge)
	mux.HandleFunc("/api/v1/auth/wallet/verify", server.HandleWalletVerify)

	// Core
	mux.HandleFunc("/api/v1/accounts", server.HandleAccounts)
	mux.HandleFunc("/api/v1/transactions", server.HandleTransactions)

	// KYC (authenticated)
	mux.Handle("/api/v1/kyc/submit", auth.Middleware(http.HandlerFunc(server.HandleKYCSubmit)))
	mux.Handle("/api/v1/kyc/status", auth.Middleware(http.HandlerFunc(server.HandleKYCStatus)))

	// Admin KYC
	mux.Handle("/api/v1/admin/kyc/pending", auth.AdminOnly(http.HandlerFunc(server.HandleAdminPendingKYC)))
	mux.Handle("/api/v1/admin/kyc/review", auth.AdminOnly(http.HandlerFunc(server.HandleAdminReviewKYC)))

	// Cards
	mux.Handle("/api/v1/cards", auth.Middleware(http.HandlerFunc(server.HandleCards)))
	mux.Handle("/api/v1/cards/", auth.Middleware(http.HandlerFunc(server.HandleCardItem)))

	handler := api.RateLimit(api.Recovery(api.SecurityHeaders(mux)))

	log.Printf("BRICS Pay API listening on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
