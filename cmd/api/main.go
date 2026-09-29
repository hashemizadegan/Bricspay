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
// فایل‌های استاتیک
fs := http.FileServer(http.Dir("./internal/api/static"))
mux.Handle("/static/", http.StripPrefix("/static/", fs))

// روت‌های اصلی و سلامت سرویس
mux.HandleFunc("/", srv.HandleRoot)
mux.HandleFunc("/healthz", srv.HealthCheck)

// مدیریت حساب‌ها و تراکنش‌ها
mux.HandleFunc("/api/accounts", srv.HandleAccounts)
mux.HandleFunc("/api/transactions", srv.HandleTransactions)

// مسیرهای احراز هویت و KYC (اصلاح نام متدها)
mux.HandleFunc("/api/auth/register", srv.HandleRegister)
mux.HandleFunc("/api/auth/login", srv.HandleLogin)
mux.HandleFunc("/api/kyc/upload", srv.HandleKYCUpload)           // جایگزین HandleKYCSubmission
mux.HandleFunc("/api/admin/profiles", srv.HandleAdminProfiles)  // جایگزین HandleKYCList
mux.HandleFunc("/api/admin/decision", srv.HandleAdminDecision)
mux.HandleFunc("/api/admin/audit", srv.HandleAdminAudit)

	log.Printf("BRICS Pay Settlement Server running on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}
