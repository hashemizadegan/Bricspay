package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"bricspay/internal/api"
	"bricspay/internal/auth"
	"bricspay/internal/db"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set") // هرگز مقدار secret را لاگ نکن
	}
	if len(os.Getenv("JWT_SECRET")) < 32 {
		log.Fatal("JWT_SECRET must be set and at least 32 bytes") // حذف fallback پیش‌فرض
	}

	database, err := db.New(dsn) // از constructor دارای pool + Ping استفاده کن
	if err != nil {
		log.Fatalf("database init failed: %v", err) // تضمین کن db.New خود DSN را در error نمی‌گذارد
	}
	defer database.Close()

	srv := api.NewServer(database)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.HealthCheck) // liveness
	mux.HandleFunc("/readyz", srv.ReadyCheck)  // readiness (DB ping)

	// همه مسیرهای دارای داده باید از middleware عبور کنند
	mux.Handle("/accounts", auth.Middleware(srv.HandleAccounts))
	mux.Handle("/transactions", auth.Middleware(srv.HandleTransactions))

	// مسیرهای احراز هویت wallet (عمومی، اما rate-limited)
	// NOTE: نام دقیق این handlerها در wallet_handlers.go تأیید نشده — تطبیق بده
	mux.HandleFunc("/auth/wallet/challenge", srv.HandleWalletChallenge)
	mux.HandleFunc("/auth/wallet/verify", srv.HandleWalletVerify)

	// استاتیک از داخل باینری سرو شود، نه از دایرکتوری اجرا
	mux.Handle("/", api.StaticHandler())

	var handler http.Handler = mux
	handler = api.SecurityHeaders(handler)
	handler = api.RateLimit(handler)
	handler = api.Recovery(handler) // بیرونی‌ترین لایه

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	httpSrv := &http.Server{
		Addr:              ":" + strings.TrimPrefix(port, ":"),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("server listening on %s", httpSrv.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
