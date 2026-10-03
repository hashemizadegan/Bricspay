package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bricspay/internal/api"
	"bricspay/internal/db"
	"bricspay/internal/ledger"

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
		database, err = db.New(dbURL)
		if err != nil {
			log.Printf("[WARN] DB init failed: %v", err)
		} else {
			defer database.Close()
		}
	} else {
		log.Println("[WARN] DATABASE_URL is not set. Running in partial mode.")
	}

	ldg := ledger.NewService(database)
	srv := api.NewServer(database, ldg)

	mux := http.NewServeMux()

	// 1. Core Health & Metrics
	mux.HandleFunc("/health", srv.HealthCheck)
	mux.HandleFunc("/ready", srv.ReadyCheck)

	// 2. Ledger & Accounts
	mux.HandleFunc("/accounts", srv.HandleAccounts)
	mux.HandleFunc("/transactions", srv.HandleTransactions)

	// 3. Auth & Corporate KYC Routes
	mux.HandleFunc("/api/v1/auth/register", srv.HandleRegister)
	mux.HandleFunc("/api/v1/auth/login", srv.HandleLogin)
	mux.HandleFunc("/api/v1/kyc/upload", srv.HandleKYCUpload)

	// 4. Admin KYC & Audit Routes
	mux.HandleFunc("/api/v1/admin/kyc/list", srv.HandleAdminProfiles)
	mux.HandleFunc("/api/v1/admin/kyc/decide", srv.HandleAdminDecision)
	mux.HandleFunc("/api/v1/admin/audit/list", srv.HandleAdminAudit)

	// 5. Web3 / MetaMask Wallet Challenge-Response
	mux.HandleFunc("/api/v1/wallet/challenge", srv.HandleWalletChallenge)
	mux.HandleFunc("/api/v1/wallet/verify", srv.HandleWalletVerify)

	// 6. Static Assets (Fixed with StripPrefix)
	staticFS := http.FileServer(http.Dir("internal/api/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", staticFS))

	// Root Index Serve
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "internal/api/static/index.html")
	})

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[INFO] BRICS Pay Core Settlement running on port :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] Listen error: %v", err)
		}
	}()

	<-stop
	log.Println("[INFO] Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] Forced shutdown: %v", err)
	}
	log.Println("[INFO] Server stopped cleanly.")
}
