package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"bricspay/internal/api"
	"bricspay/internal/db"
	_ "github.com/lib/pq"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	connStr := os.Getenv("DATABASE_URL")

	var database *sql.DB
	var err error

	if connStr != "" {
		database, err = db.InitDB(connStr)
		if err != nil {
			log.Printf("Database connection warning: %v", err)
		} else {
			defer database.Close()
		}
	}

	// ایجاد نمونه سرور با دیتابیس
	srv := api.NewServer(database)

	mux := http.NewServeMux()

	// ۱. سرو کردن فایل‌های CSS و JS از مسیر واقعی پروژه
	// نکته کلیدی: مسیر روی سرور ./internal/api/static است
	staticDir := "./internal/api/static"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		// حالت فال‌بک در صورتی که دایرکتوری در روت کپی شده باشد
		staticDir = "./static"
	}
	fs := http.FileServer(http.Dir(staticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// ۲. روت‌های API و احراز هویت
	mux.HandleFunc("/api/v1/accounts", srv.HandleAccounts)
	mux.HandleFunc("/api/v1/transactions", srv.HandleTransactions)
	mux.HandleFunc("/api/v1/kyc", srv.HandleKYCSubmission)
	mux.HandleFunc("/api/v1/admin/kyc", srv.HandleKYCList)
	mux.HandleFunc("/api/v1/auth/login", srv.HandleLogin)
	mux.HandleFunc("/api/v1/auth/register", srv.HandleRegister)

	// ۳. روت سلامت سرویس
	mux.HandleFunc("/health", srv.HealthCheck)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// ۴. ریشه سایت (داشبورد اصلی)
	mux.HandleFunc("/", srv.HandleRoot)

	log.Printf("BRICS Pay Settlement Server running on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}
