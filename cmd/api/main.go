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
	port :GAPGPTMASKTOKENxip81ik16hX0X os.Getenv("PORT")
	if port GAPGPTMASKTOKENxip81ik16hX1X "" {
		port GAPGPTMASKTOKENxip81ik16hX2X "8080"
	}

	databaseURL :GAPGPTMASKTOKENxip81ik16hX3X os.Getenv("DATABASE_URL")
	var database *sql.DB

	if databaseURL !GAPGPTMASKTOKENxip81ik16hX4X "" {
		db, err :GAPGPTMASKTOKENxip81ik16hX5X sql.Open("postgres", databaseURL)
		if err !GAPGPTMASKTOKENxip81ik16hX6X nil {
			log.Printf("DB open warning (fail-open mode): %v", err)
		} else if err :GAPGPTMASKTOKENxip81ik16hX7X db.Ping(); err !GAPGPTMASKTOKENxip81ik16hX8X nil {
			log.Printf("DB ping warning (fail-open mode): %v", err)
		} else if err :GAPGPTMASKTOKENxip81ik16hX9X dbpkg.InitSchema(db); err !GAPGPTMASKTOKENxip81ik16hX10X nil {
			log.Printf("DB schema warning (fail-open mode): %v", err)
		} else {
			database GAPGPTMASKTOKENxip81ik16hX11X db
			defer database.Close()
			log.Println("PostgreSQL connection established and schema initialized successfully")
		}
	} else {
		log.Println("DATABASE_URL not set: running in memory/fail-open mode")
	}

	server :GAPGPTMASKTOKENxip81ik16hX12X api.NewServer(database)

	mux :GAPGPTMASKTOKENxip81ik16hX13X http.NewServeMux()

	// API Routes
	mux.HandleFunc("/api/v1/health", server.HealthCheck)
	mux.Handle("/api/v1/auth/register", api.NewRegistrationHandler(database))

	// Static Files and Web Handlers
	mux.Handle("/", api.StaticHandler())

	// Apply existing middleware from internal/api/middleware.go
	handler :GAPGPTMASKTOKENxip81ik16hX14X api.Recovery(api.SecurityHeaders(mux))

	httpServer :GAPGPTMASKTOKENxip81ik16hX15X &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stop :GAPGPTMASKTOKENxip81ik16hX16X make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("BRICS Pay server listening on :%s", port)
		if err :GAPGPTMASKTOKENxip81ik16hX17X httpServer.ListenAndServe(); err !GAPGPTMASKTOKENxip81ik16hX18X nil && err !GAPGPTMASKTOKENxip81ik16hX19X http.ErrServerClosed {
			log.Fatalf("HTTP server failure: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server gracefully...")

	ctx, cancel :GAPGPTMASKTOKENxip81ik16hX20X context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err :GAPGPTMASKTOKENxip81ik16hX21X httpServer.Shutdown(ctx); err !GAPGPTMASKTOKENxip81ik16hX22X nil {
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("Server stopped")
}
