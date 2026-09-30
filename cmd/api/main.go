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

	dbURL := os.Getenv("DATABASE_URL")
	var database *sql.DB
	var err error

	if dbURL != "" {
		// فراخوانی InitDB که جداول accounts، users و kyc_profiles را خودکار می‌سازد
		database, err = dbpkg.InitDB(dbURL)
		if err != nil {
			log.Printf("خطا در راه‌اندازی و مایگریشن دیتابیس: %v", err)
		} else {
			log.Println("ارتباط با دیتابیس Postgres برقرار و تمام جداول (users, kyc) ساخته شدند.")
			defer database.Close()
		}
	} else {
		log.Println("هشدار: DATABASE_URL تنظیم نشده است؛ در حال اجرا در حالت بدون دیتابیس.")
	}

	server := api.NewServer(database)

	mux := http.NewServeMux()

	// سرو فایل‌های استاتیک UI
	fs := http.FileServer(http.Dir("internal/api/static"))
	mux.Handle("/", fs)

	// روت‌های احراز هویت و KYC
	mux.HandleFunc("/api/v1/auth/register", server.HandleRegister)
	mux.HandleFunc("/api/v1/auth/login", server.HandleLogin)
	mux.HandleFunc("/api/v1/kyc/profile", server.HandleKYCProfile)
	mux.HandleFunc("/api/v1/kyc/documents", server.HandleKYCDocuments)
	mux.HandleFunc("/api/v1/admin/kyc/list", server.HandleAdminKYCList)
	mux.HandleFunc("/api/v1/admin/kyc/evaluate", server.HandleAdminKYCEvaluate)

	// روت‌های پایه Ledger
	mux.HandleFunc("/api/v1/accounts", server.HandleAccounts)
	mux.HandleFunc("/api/v1/transactions", server.HandleTransactions)
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"online","service":"BRICS Pay Settlement API Gateway"}`))
	})

	log.Printf("سرور BRICS Pay روی پورت %s آغاز به کار کرد...", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("خطای اجرای سرور: %v", err)
	}
}
