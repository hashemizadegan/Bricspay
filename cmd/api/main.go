package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"bricspay/internal/api"
	"bricspay/internal/auth"
	"bricspay/internal/db"
	"bricspay/internal/ledger"
)

func main() {
	// 1. بارگذاری متغیرهای محیطی
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Println("هشدار: DATABASE_URL تنظیم نشده است؛ در حال استفاده از حالت پیش‌فرض محلی...")
		databaseURL = "postgres://postgres:postgres@localhost:5432/bricspay?sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "bricspay-production-secure-jwt-key-default"
	}
	auth.SetJWTSecret(jwtSecret)

	// 2. اتصال به دیتابیس
	database, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("خطا در اتصال اولیه به دیتابیس: %v", err)
	}
	defer database.Close()

	database.SetMaxOpenConns(25)
	database.SetMaxIdleConns(5)
	database.SetConnMaxLifetime(5 * time.Minute)

	if err := database.Ping(); err != nil {
		log.Printf("هشدار: امکان برقراری ارتباط با دیتابیس وجود ندارد: %v", err)
	} else {
		log.Println("اتصال به پایگاه‌داده با موفقیت برقرار شد.")
	}

	// 3. اعمال مایگریشن‌ها و مقداردهی اولیه‌ی جداول
	if err := db.InitDB(database); err != nil {
		log.Printf("هشدار در اجرای مایگریشن‌های پایگاه‌داده: %v", err)
	}

	// 4. راه‌اندازی لایه‌های سرویس و هندلرها
	ledgerService := ledger.NewService(database)
	apiHandler := api.NewHandler(database, ledgerService)

	mux := http.NewServeMux()

	// --- [الف] روت‌های فرانت‌اند و صفحات استاتیک ---

	// ۱. لندینگ پیج مستقل در روت اصلی
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "internal/api/static/index.html")
	})

	// ۲. نرم‌افزار و پنل اصلی (SPA) در مسیر /app
	mux.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "internal/api/static/app.html")
	})
	mux.HandleFunc("/app/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "internal/api/static/app.html")
	})

	// ۳. سرو کردن استاتیک فایل‌ها (CSS, JS, Images, ...)
	fileServer := http.FileServer(http.Dir("internal/api/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// --- [ب] روت‌های سیستم و Health Check ---
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","version":"11.0.0"}`))
	})

	// --- [ج] روت‌های عمومی احراز هویت و کیف پول (API) ---
	mux.HandleFunc("/api/auth/login", apiHandler.HandleLogin)
	mux.HandleFunc("/api/auth/register", apiHandler.HandleRegister)
	mux.HandleFunc("/api/wallet/nonce", apiHandler.HandleWalletNonce)
	mux.HandleFunc("/api/wallet/verify", apiHandler.HandleWalletVerify)

	// --- [د] روت‌های نیازمند احراز هویت (Protected API) ---
	mux.Handle("/api/cards", api.JWTMiddleware(http.HandlerFunc(apiHandler.HandleCards)))
	mux.Handle("/api/cards/", api.JWTMiddleware(http.HandlerFunc(apiHandler.HandleCardAction)))
	mux.Handle("/api/kyc/submit", api.JWTMiddleware(http.HandlerFunc(apiHandler.HandleKYCSubmit)))
	mux.Handle("/api/kyc/status", api.JWTMiddleware(http.HandlerFunc(apiHandler.HandleKYCStatus)))
	mux.Handle("/api/transfers", api.JWTMiddleware(http.HandlerFunc(apiHandler.HandleTransfer)))
	mux.Handle("/api/accounts/balance", api.JWTMiddleware(http.HandlerFunc(apiHandler.HandleGetBalance)))
	mux.Handle("/api/transactions", api.JWTMiddleware(http.HandlerFunc(apiHandler.HandleGetTransactions)))

	// --- [هـ] روت‌های پنل مدیریت (Admin Protected) ---
	mux.Handle("/api/admin/kyc/pending", api.AdminMiddleware(http.HandlerFunc(apiHandler.HandleAdminPendingKYC)))
	mux.Handle("/api/admin/kyc/review", api.AdminMiddleware(http.HandlerFunc(apiHandler.HandleAdminReviewKYC)))
	mux.Handle("/api/admin/stats", api.AdminMiddleware(http.HandlerFunc(apiHandler.HandleAdminStats)))

	// 5. اعمال Middleware های امنیتی و عمومی (CORS و Logging)
	handler := api.LoggingMiddleware(api.CORSMiddleware(mux))

	// 6. پیکربندی و اجرای سرور
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 7. مدیریت خاموش شدن تمیز (Graceful Shutdown)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("سرور BRICS Pay (نسخه 11.0.0) روی پورت :%s در حال اجرا است...", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("خطا در اجرای سرور: %v", err)
		}
	}()

	<-stop
	log.Println("دریافت سیگنال خروج؛ در حال متوقف کردن سرور...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("خطا در خاموش کردن سرور: %v", err)
	}

	log.Println("سرور با موفقیت متوقف شد.")
}
