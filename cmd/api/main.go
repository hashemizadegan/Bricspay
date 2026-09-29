package main

import (
	"log"
	"net/http"
	"os"
	"bricspay/internal/api"
	"bricspay/internal/db"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	connStr := os.Getenv("DATABASE_URL")
	// مقداردهی دیتابیس مستقیماً به صورت *sql.DB
	database, err := db.InitDB(connStr)
	if err != nil {
		log.Printf("Database warning: %v", err)
	} else {
		defer database.Close()
	}

	// ایجاد سرور با استفاده از دیتابیس
	srv := api.NewServer(database)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/transactions", srv.HandleTransactions)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
