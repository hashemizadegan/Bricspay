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

	fs := mux.Handle("/", api.StaticHandler())
	mux.Handle("/", fs)

	mux.HandleFunc("/api/v1/health", srv.HealthCheck)
	mux.HandleFunc("/api/v1/auth/register", srv.HandleRegister)
	mux.HandleFunc("/auth/login", srv.HandleLogin)
	mux.HandleFunc("/api/v1/auth/wallet/challenge", srv.HandleWalletChallenge)
	mux.HandleFunc("/api/v1/auth/wallet/verify", srv.HandleWalletVerify)
	mux.HandleFunc("/api/v1/kyc/upload", srv.HandleKYCUpload)
	mux.HandleFunc("/api/v1/admin/kyc/profiles", srv.HandleAdminProfiles)
	mux.HandleFunc("/api/v1/admin/kyc/decision", srv.HandleAdminDecision)
	mux.HandleFunc("/api/v1/admin/kyc/audit", srv.HandleAdminAudit)
	mux.HandleFunc("/api/v1/accounts", srv.HandleAccounts)
	mux.HandleFunc("/api/v1/transactions", srv.HandleTransactions)
    mux.Handle("/api/v1/kyc/status", api.authWrap(srv.HandleKYCStatus))
    mux.Handle("/api/v1/kyc/resubmit", api.authWrap(srv.HandleKYCResubmit))
    mux.Handle("/api/v1/cards", api.authWrap(srv.HandleCards))
    mux.Handle("/api/v1/cards/", api.authWrap(srv.HandleCardItem))

	log.Printf("🚀 سرور BRICS Pay روی پورت %s در حال اجرا است", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("خطا در راه‌اندازی سرور: %v", err)
	}
}
