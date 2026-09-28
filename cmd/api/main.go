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

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("DATABASE_URL not set; running with nil db (in-memory/limited mode)")
	}

	var databaseConn *db.Database = nil
	var sqlDB = (*struct{ *http.Server })(nil) // placeholder check

	database, err := db.InitDB(dbURL)
	if err != nil {
		log.Printf("Warning: Database connection failed: %v. Running in mock/fallback mode.\n", err)
	} else {
		log.Println("Successfully connected to PostgreSQL database.")
		defer database.Close()
	}

	server := api.NewServer(database)

	mux := http.NewServeMux()
	mux.HandleFunc("/", server.HandleRoot)
	mux.HandleFunc("/health", server.HealthCheck)
	mux.HandleFunc("/api/v1/accounts", server.HandleAccounts)
	mux.HandleFunc("/api/v1/transactions", server.HandleTransactions)
	mux.HandleFunc("/api/v1/banks", server.HandleBanks)

	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("BRICS Pay Gateway starting on port %s\n", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped.")
}
