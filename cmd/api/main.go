package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"bricspay/internal/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	var db *sql.DB
	var err error

	if dbURL != "" {
		db, err = sql.Open("postgres", dbURL)
		if err != nil {
			log.Printf("Warning: Database connection failed: %v", err)
		} else {
			defer db.Close()
			if err = db.Ping(); err != nil {
				log.Printf("Warning: Database ping failed: %v", err)
			} else {
				log.Println("Connected to PostgreSQL successfully.")
			}
		}
	} else {
		log.Println("DATABASE_URL not set; running without database.")
	}

	server := api.NewServer(db)
	mux := http.NewServeMux()

	// روت‌های اصلی
	mux.HandleFunc("/", server.HandleRoot)
	mux.HandleFunc("/health", server.HealthCheck)
	mux.HandleFunc("/accounts", server.HandleAccounts)
	mux.HandleFunc("/transactions", server.HandleTransactions)

	// سرو کردن فایل‌های استاتیک
	fs := http.FileServer(http.Dir("./internal/api/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Printf("Server listening on port %s...\n", port)
	
	// اینجا فقط از = استفاده شده تا خطای no new variables رخ ندهد
	if err = http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
