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
	"bricspay/internal/ledger"
)

func main() {
	// ۱. دریافت متغیرهای محیطی با مقادیر امن پیش‌فرض
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Println("⚠️ هشدار: متغیر DATABASE_URL یافت نشد؛ بررسی دیتابیس محلی...")
		databaseURL = "postgres://postgres:postgres@localhost:5432/bricspay?sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "bricspay-production-secure-secret-key-v11"
	}

	// ۲. اتصال مستقیم و استاندارد به دیتابیس PostgreSQL
	dbConn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("❌ خطا در ایجاد کانکشن دیتابیس: %v", err)
	}
	defer dbConn.Close()

	dbConn.SetMaxOpenConns(25)
	dbConn.SetMaxIdleConns(10)
	dbConn.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := dbConn.PingContext(ctx); err != nil {
		log.Printf("⚠️ هشدار اتصال دیتابیس: %v (سیستم در حال ادامه راه‌اندازی است)", err)
	} else {
		log.Println("✅ اتصال به پایگاه داده با موفقیت برقرار شد.")
	}

	// ۳. آماده‌سازی سرویس‌ها (تطابق کامل با v11: بدون توابع منسوخ شده)
	authService := auth.NewService(jwtSecret)
	ledgerService := ledger.New(dbConn)

	// ۴. راه‌اندازی روت‌ها و سرور API
	server := api.NewServer(dbConn, authService, ledgerService)
	router := server.Routes()

	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ۵. مدیریت Graceful Shutdown برای Railway
	go func() {
		log.Printf("🚀 سرویس BricsPay v11.0.0 بر روی پورت %s آماده دریافت ترافیک است...", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ خطای اجرای سرور: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 دریافت سیگنال خروج، در حال متوقف‌سازی ایمن سرور...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("❌ خطا در متوقف‌سازی ایمن: %v", err)
	}

	log.Println("✅ سرور با موفقیت متوقف شد.")
}
