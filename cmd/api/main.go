package main

import (
	"log"
	"net/http"
	"os"

	"bricspay/internal/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// سرو فایل‌های استاتیک داشبورد
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.Handle("/", fs)

	// روت‌های API
	mux.HandleFunc("/api/v1/transactions", api.TransactionHandler)

	// Healthcheck برای اطمینان Railway از سلامت سرویس
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Server listening on port %s...", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}
