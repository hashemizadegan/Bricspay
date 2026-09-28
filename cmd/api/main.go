package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bricspay/internal/api"
	"bricspay/internal/db"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	database, err := db.InitDB(databaseURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to database: %v. Running in in-memory mode.", err)
	}

	server := api.NewServer(database)

	mux := http.NewServeMux()

	// سرو کردن فایل‌های CSS و JS استاتیک
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(api.StaticFS())))

	// روت‌های اصلی برنامه
	mux.HandleFunc("/", server.HandleRoot)
	mux.HandleFunc("/health", server.HealthCheck)
	mux.HandleFunc("/api/v1/accounts", server.HandleAccounts)
	mux.HandleFunc("/api/v1/transactions", server.HandleTransactions)
	mux.HandleFunc("/api/v1/banks", server.HandleBanks)

	// روت‌های احراز هویت و KYC
	mux.HandleFunc("/api/v1/auth/register", server.HandleRegister)
	mux.HandleFunc("/api/v1/auth/login", server.HandleLogin)
	mux.HandleFunc("/api/v1/kyc/upload", server.HandleKYCUpload)
	mux.HandleFunc("/api/v1/admin/profiles", server.HandleAdminProfiles)
	mux.HandleFunc("/api/v1/admin/decision", server.HandleAdminDecision)
	mux.HandleFunc("/api/v1/admin/audit", server.HandleAdminAudit)

	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("BRICS Pay Server starting on port %s...", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server stopped successfully.")
}
