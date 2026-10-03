package main

import (
	"database/sql"
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
		log.Println("JWT_SECRET سفارشی فعال شد.")
	}

	var database *sql.DB

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("هشدار: DATABASE_URL تنظیم نشده است؛ سرور بدون دیتابیس اجرا می‌شود.")
	} else {
		var err error
		database, err = sql.Open("postgres", dbURL)
		if err != nil {
			log.Fatalf("خطا در اتصال به پایگاه داده: %v", err)
		}
		defer database.Close()

		if err := dbpkg.InitSchema(database); err != nil {
			log.Fatalf("خطا در راه‌اندازی schema: %v", err)
		}

		log.Println("Schema پایگاه داده آماده است.")
	}

	server := api.NewServer(database)
	mux := http.NewServeMux()

	mux.Handle("/", api.StaticHandler())

	mux.HandleFunc("/api/v1/health", server.HealthCheck)
	mux.HandleFunc("/api/v1/auth/register", server.HandleRegister)
	mux.HandleFunc("/auth/login", server.HandleLogin)

	// نام صحیح handler در wallet_handlers.go این است:
	mux.HandleFunc(
		"/api/v1/auth/wallet/challenge",
		server.HandleWalletChallenge,
	)
	mux.HandleFunc(
		"/api/v1/auth/wallet/verify",
		server.HandleWalletVerify,
	)

	mux.HandleFunc("/api/v1/kyc/upload", server.HandleKYCUpload)
	mux.HandleFunc("/api/v1/admin/kyc/profiles", server.HandleAdminProfiles)
	mux.HandleFunc("/api/v1/admin/kyc/decision", server.HandleAdminDecision)
	mux.HandleFunc("/api/v1/admin/kyc/audit", server.HandleAdminAudit)
	mux.HandleFunc("/api/v1/accounts", server.HandleAccounts)
	mux.HandleFunc("/api/v1/transactions", server.HandleTransactions)

	mux.Handle(
		"/api/v1/kyc/status",
		api.AuthWrap(server.HandleKYCStatus),
	)
	mux.Handle(
		"/api/v1/kyc/resubmit",
		api.AuthWrap(server.HandleKYCResubmit),
	)
	mux.Handle(
		"/api/v1/cards",
		api.AuthWrap(server.HandleCards),
	)
	mux.Handle(
		"/api/v1/cards/",
		api.AuthWrap(server.HandleCardItem),
	)

	log.Printf("سرور BRICS Pay روی پورت %s در حال اجرا است", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("خطا در راه‌اندازی سرور: %v", err)
	}
}
