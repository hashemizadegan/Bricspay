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

	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		auth.SetJWTSecret(jwtSecret)
		log.Println("JWT_SECRET is set")
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
			log.Printf("warning: DB open failed (%v); running without database", err)
		} else if err := db.Ping(); err != nil {
			log.Printf("warning: DB ping failed (%v); running without database", err)
			db.Close()
		} else if err := dbpkg.InitSchema(db); err != nil {
			log.Printf("warning: schema init failed (%v); running without database", err)
			db.Close()
		} else {
			database = db
			defer database.Close()
			log.Println("database connected, schema ready")
		}
	}

	server := api.NewServer(database)
	mux := http.NewServeMux()

	mux.Handle("/", api.StaticHandler())
	mux.HandleFunc("/api/v1/health", server.HealthCheck)
	// ... بقیه مسیرها دقیقاً مثل قبل بدون تغییر باقی بمانند ...

	log.Printf("BRICS Pay server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Printf("server error: %v", err)
		os.Exit(1)
	}
}
