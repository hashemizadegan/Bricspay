package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"bricspay/internal/api"
	"bricspay/internal/auth"
	dbpkg "bricspay/internal/db"

	_ "github.com/lib/pq"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret != "" {
		auth.SetJWTSecret(jwtSecret)
		log.Println("JWT_SECRET loaded")
	} else {
		log.Println("warning: JWT_SECRET not set")
	}

	var database *sql.DB

	dbURL := os.Getenv("DATABASE_URL")
	switch {
	case dbURL == "":
		log.Println("warning: DATABASE_URL not set; running without database")

	default:
		db, err := sql.Open("postgres", dbURL)
		if err != nil {
			log.Printf("warning: database open failed: %v", err)
		} else if err := db.Ping(); err != nil {
			log.Printf("warning: database ping failed: %v", err)
			db.Close()
		} else if err := dbpkg.InitSchema(db); err != nil {
			log.Printf("warning: database schema initialization failed: %v", err)
			db.Close()
		} else {
			database = db
			defer database.Close()
			log.Println("database connected and schema initialized")
		}
	}

	server := api.NewServer(database)
	mux := http.NewServeMux()

	// Static website
	mux.Handle("/", api.StaticHandler())

	// Health
	mux.HandleFunc("/api/v1/health", server.HealthCheck)

	// Authentication
	mux.Handle(
		"/api/v1/auth/register",
		api.NewRegistrationHandler(database),
	)

	mux.HandleFunc(
		"/api/v1/auth/login",
		api.HandleLogin(database),
	)

	// KYC
	mux.HandleFunc(
		"/api/v1/kyc/submit",
		server.HandleKYCSubmit,
	)

	mux.HandleFunc(
		"/api/v1/kyc/status",
		server.HandleKYCStatus,
	)

	// KYC Upload protected by auth.Middleware
	mux.Handle(
		"/api/v1/kyc/upload",
		auth.Middleware(http.HandlerFunc(server.HandleKYCUpload)),
	)

	// Admin KYC
	mux.HandleFunc(
		"/api/v1/admin/kyc/pending",
		server.HandleAdminPendingKYC,
	)

	mux.HandleFunc(
		"/api/v1/admin/kyc/review",
		server.HandleAdminReviewKYC,
	)

	log.Printf("BRICS Pay server listening on :%s", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Printf("server error: %v", err)
		os.Exit(1)
	}
}
