package main

import (
	"log"
	"net/http"
	"os"

	"bricspay/internal/api"
	"bricspay/internal/db"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Println("WARNING: DATABASE_URL is not set. Running with nil DB...")
	}

	var database *db.DB
	var err error
	if connStr != "" {
		database, err = db.InitDB(connStr)
		if err != nil {
			log.Fatalf("Database initialization failed: %v", err)
		}
		defer database.Close()
	}

	// ایجاد سرور با اتصال پایگاه‌داده
	srv := api.NewServer(database)

	mux := http.NewServeMux()

	// روت‌های اصلی تراکنش و سیستم
	mux.HandleFunc("/api/v1/transactions", srv.HandleTransactions)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// روت‌های احراز هویت و KYC
	mux.HandleFunc("/api/v1/auth/register", srv.HandleRegister)
	mux.HandleFunc("/api/v1/auth/login", srv.HandleLogin)
	mux.HandleFunc("/api/v1/kyc/upload", srv.HandleKYCUpload)
	mux.HandleFunc("/api/v1/admin/profiles", srv.HandleAdminProfiles)
	mux.HandleFunc("/api/v1/admin/decision", srv.HandleAdminDecision)
	mux.HandleFunc("/api/v1/admin/audit", srv.HandleAdminAudit)

	// سرویس‌دهی فرانت‌اند و فایل‌های استاتیک
	fs := http.FileServer(http.Dir("./internal/api/static"))
	mux.Handle("/", fs)

	log.Printf("BRICS Pay Core Ledger started on port %s...", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}
