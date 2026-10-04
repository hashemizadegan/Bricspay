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
		log.Println("JWT_SECRET سفارشی فعال شد.")
	}

	var database *sql.DB

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("هشدار: DATABASE_URL تنظیم نشده است؛ سرور بدون دیتابیس اجرا می‌شود.")
	} else {
		var err error
		database, err = sql.Open("postgres", dbURL)
		if err != nil {
			log.Fatalf("خطا در اتصال به پایگاه داده: %v", err)
		}
		defer database.Close()

		if err := dbpkg.InitSchema(database); err != nil {
			log.Fatalf("خطا در راه‌اندازی schema: %v", err)
		}

		log.Println("Schema پایگاه داده آماده است.")
	}

	server := api.NewServer(database)
	mux := http.NewServeMux()

	mux.Handle("/", api.StaticHandler())

// سلامت سرویس
mux.HandleFunc("/api/v1/health", server.HealthCheck)
mux.HandleFunc("/api/v1/ready", server.ReadyCheck)

// احراز هویت کیف‌پولی؛ در v12 جایگزین register/login قدیمی
mux.HandleFunc(
	"/api/v1/auth/wallet/challenge",
	server.HandleWalletChallenge,
)
mux.HandleFunc(
	"/api/v1/auth/wallet/verify",
	server.HandleWalletVerify,
)

// عملیات نیازمند احراز هویت
mux.Handle(
	"/api/v1/kyc/submit",
	auth.AuthMiddleware(http.HandlerFunc(server.HandleKYCSubmit)),
)
mux.Handle(
	"/api/v1/kyc/status",
	auth.AuthMiddleware(http.HandlerFunc(server.HandleKYCStatus)),
)

// در v12 ارسال مجدد با همان handler ثبت KYC انجام می‌شود.
// اگر فرانت‌اند از این مسیر استفاده می‌کند، این route را نگه دارید.
mux.Handle(
	"/api/v1/kyc/resubmit",
	auth.AuthMiddleware(http.HandlerFunc(server.HandleKYCSubmit)),
)

// مدیریت KYC
mux.Handle(
	"/api/v1/admin/kyc/pending",
	auth.AuthMiddleware(http.HandlerFunc(server.HandleAdminPendingKYC)),
)
mux.Handle(
	"/api/v1/admin/kyc/review",
	auth.AuthMiddleware(http.HandlerFunc(server.HandleAdminReviewKYC)),
)

// کارت‌ها، حساب‌ها و تراکنش‌ها
mux.Handle(
	"/api/v1/cards",
	auth.AuthMiddleware(http.HandlerFunc(server.HandleCards)),
)
mux.Handle(
	"/api/v1/cards/",
	auth.AuthMiddleware(http.HandlerFunc(server.HandleCardItem)),
)
mux.Handle(
	"/api/v1/accounts",
	auth.AuthMiddleware(http.HandlerFunc(server.HandleAccounts)),
)
mux.Handle(
	"/api/v1/transactions",
	auth.AuthMiddleware(http.HandlerFunc(server.HandleTransactions)),
)

	log.Printf("سرور BRICS Pay روی پورت %s در حال اجرا است", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("خطا در راه‌اندازی سرور: %v", err)
	}
}
