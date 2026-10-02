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
		
		log.Println("✅ JWT_SECRET سفارشی با موفقیت فعال شد.")
	}

	var database *sql.DB
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("هشدار: DATABASE_URL تنظیم نشده است؛ در حال اجرا در حالت بدون دیتابیس.")
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
		log.Println("✅ اسکیمای پایگاه داده و جداول احراز هویت کیف پول با موفقیت بررسی و آماده شدند.")
	}

	srv := api.NewServer(database)
	mux := http.NewServeMux()

	// Static frontend assets
	fs := http.FileServer(http.Dir("./internal/api/static"))
	mux.Handle("/", fs)

	// Health
	mux.HandleFunc("/api/v1/health", srv.HealthCheck)

	// Standard Auth (Email/Password)
	mux.HandleFunc("/api/v1/auth/register", srv.HandleRegister)
	mux.HandleFunc("/auth/login", srv.HandleLogin)

	// MetaMask EIP-191 Personal Sign Auth
	mux.HandleFunc("/api/v1/auth/wallet/challenge", srv.HandleWalletChallenge)
	mux.HandleFunc("/api/v1/auth/wallet/verify", srv.HandleWalletVerify)

	// KYC
	mux.HandleFunc("/api/v1/kyc/upload", srv.HandleKYCUpload)
	mux.HandleFunc("/api/v1/admin/kyc/profiles", srv.HandleAdminProfiles)
	mux.HandleFunc("/api/v1/admin/kyc/decision", srv.HandleAdminDecision)
	mux.HandleFunc("/api/v1/admin/kyc/audit", srv.HandleAdminAudit)

	// Accounts & Transactions
	mux.HandleFunc("/api/v1/accounts", srv.HandleAccounts)
	mux.HandleFunc("/api/v1/transactions", srv.HandleTransactions)

	log.Printf("🚀 سرور BRICS Pay روی پورت %s در حال اجرا است", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("خطا در راه‌اندازی سرور: %v", err)
	}
}
