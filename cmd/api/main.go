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
	
	// تعریف متغیر پایگاه داده به صورت *sql.DB
	var database *sql.DB
	var err error
	
	// استفاده از تابع InitDB به جای db.DB
	if connStr != "" {
		database, err = db.InitDB(connStr) 
		if err != nil {
			log.Printf("Database connection failed: %v", err)
		} else {
			defer database.Close()
		}
	}

	// ایجاد سرور
	srv := api.NewServer(database)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/transactions", srv.HandleTransactions)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// سرویس‌دهی فایل‌های استاتیک
	fs := http.FileServer(http.Dir("./internal/api/static"))
	mux.Handle("/", fs)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
