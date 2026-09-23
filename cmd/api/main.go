package main

import (
	"log"
	"net/http"
	"os"

	"bricspayir/internal/api"
	"bricspayir/internal/db"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	database, err := db.InitDB(dbURL)
	if err != nil {
		log.Fatalf("db init failed: %v", err)
	}
	defer database.Close()

	srv := api.NewServer(database)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.HealthCheck)
	mux.HandleFunc("/api/v1/accounts", srv.HandleAccounts)
	mux.HandleFunc("/api/v1/transactions", srv.HandleTransactions)

	log.Printf("bricspayir core ledger listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
