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
	dbpkg "bricspay/internal/db"

	_ "github.com/lib/pq"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	var database *sql.DB

	if databaseURL != "" {
		db, err := sql.Open("postgres", databaseURL)
		if err != nil {
			log.Printf("DB open warning (fail-open mode): %v", err)
		} else if err := db.Ping(); err != nil {
			log.Printf("DB ping warning (fail-open mode): %v", err)
		} else if err := dbpkg.InitSchema(db); err != nil {
			log.Printf("DB schema warning (fail-open mode): %v", err)
		} else {
			database = db
			defer database.Close()
			log.Println("PostgreSQL connection established and schema initialized successfully")
		}
	} else {
		log.Println("DATABASE_URL not set: running in memory/fail-open mode")
	}

	server := api.NewServer(database)

	mux := http.NewServeMux()

	// هندلرهای API
	mux.HandleFunc("/api/v1/health", server.HealthCheck)
	mux.Handle("/api/v1/auth/register", api.NewRegistrationHandler(database))

	// هندلر فایل‌های استاتیک و صفحات فرانت‌اند (باید آخرین روت باشد)
	mux.Handle("/", api.StaticHandler())

	handler := api.RecoveryMiddleware(api.AuditMiddleware(api.CORSMiddleware(mux)))

	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("BRICS Pay server listening on :%s", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failure: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("Server stopped")
}
