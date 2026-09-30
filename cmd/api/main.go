package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"bricspay/internal/api"
	dbpkg "bricspay/internal/db"

	_ "github.com/lib/pq"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
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
	}

	srv := api.NewServer(database)
	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir("./internal/api/static"))
	mux.Handle("/", fs)

	// Health
	mux.HandleFunc("/api/v1/health", srv.HealthCheck)

	// Auth
	mux.HandleFunc("/api/v1/auth/register", api.HandleRegister(database))
	mux.HandleFunc("/auth/login", api.HandleLogin(database))

	// KYC
	mux.HandleFunc("/api/v1/kyc/upload", api.HandleKYCUpload(database))
	mux.HandleFunc("/api/v1/admin/kyc/profiles", api.HandleAdminProfiles(database))
	mux.HandleFunc("/api/v1/admin/kyc/decision", api.HandleAdminDecision(database))
	mux.HandleFunc("/api/v1/admin/kyc/audit", api.HandleAdminAudit(database))

	// Accounts & Transactions
	mux.HandleFunc("/api/v1/accounts", srv.HandleAccounts)
	mux.HandleFunc("/api/v1/transactions", srv.HandleTransactions)

	log.Printf("سرور روی پورت %s در حال اجرا است", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("خطا در راه‌اندازی سرور: %v", err)
	}
}
