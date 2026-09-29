package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"bricspay/internal/api"

	_ "github.com/lib/pq"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	var db *sql.DB
	var err error

	if dbURL != "" {
		db, err = sql.Open("postgres", dbURL)
		if err != nil {
			log.Printf("هشدار: اتصال به پایگاه داده با خطا مواجه شد: %v", err)
		} else {
			if err = db.Ping(); err != nil {
				log.Printf("هشدار: ارتباط مستقیم با پایگاه داده برقرار نشد: %v", err)
			} else {
				log.Println("ارتباط با دیتابیس Postgres با موفقیت برقرار شد.")
			}
		}
	} else {
		log.Println("DATABASE_URL تنظیم نشده است؛ در حال اجرا در حالت بدون دیتابیس.")
	}

	server := api.NewServer(db)

	mux := http.NewServeMux()

	// روت‌های اصلی و سلامت سیستم
	mux.HandleFunc("/", server.HandleRoot)
	mux.HandleFunc("/health", server.HealthCheck)
	mux.HandleFunc("/healthz", server.HealthCheck)

	// روت‌های لجر و حساب‌ها
	mux.HandleFunc("/accounts", server.HandleAccounts)
	mux.HandleFunc("/transactions", server.HandleTransactions)
	mux.HandleFunc("/api/v1/accounts", server.HandleAccounts)
	mux.HandleFunc("/api/v1/transactions", server.HandleTransactions)

	// روت‌های احراز هویت (Auth)
	mux.HandleFunc("/api/v1/auth/register", server.HandleRegister)
	mux.HandleFunc("/api/v1/auth/login", server.HandleLogin)

	// روت بارگذاری مدارک KYC
	mux.HandleFunc("/api/v1/kyc/upload", server.HandleKYCUpload)

	// روت‌های ادمین و نظارت (منطبق با متدهای پیاده‌شده در kyc_handlers.go)
	mux.HandleFunc("/api/v1/admin/kyc/list", server.HandleAdminProfiles)
	mux.HandleFunc("/api/v1/admin/kyc/decision", server.HandleAdminDecision)
	mux.HandleFunc("/api/v1/admin/audit", server.HandleAdminAudit)

	// فایل‌های استاتیک فرانت‌اند
	fs := http.FileServer(http.Dir("internal/api/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	addr := fmt.Sprintf(":%s", port)
	log.Printf("سرور BRICS Pay با موفقیت روی پورت %s آماده دریافت درخواست‌ها است...", port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("خطا در اجرای سرور: %v", err)
	}
}
