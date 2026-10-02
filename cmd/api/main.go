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
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer sqlDB.Close()

	dbConn := &db.DB{DB: sqlDB}

	// استفاده از mux استاندارد به جای NewRouter
	router := http.NewServeMux()

	// ثبت هندلرها (در صورت وجود توابع ثبت در پکیج api)
	api.RegisterRoutes(router, dbConn) // اگر این تابع وجود نداشت، بعداً اصلاح می‌کنیم

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
