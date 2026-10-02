package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"bricspay/internal/api"
	_ "github.com/lib/pq"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is requireGAPGPTMASKTOKEN051tukh7evbnX0X	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer sqlDB.Close()

	srv := api.NewServer(sqlDB)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.HealthCheck)
	mux.HandleFunc("/accounts", srv.HandleAccounts)
	mux.HandleFunc("/transactions", srv.HandleTransactions)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
