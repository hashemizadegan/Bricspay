package api

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"bricspay/internal/auth"
	"bricspay/internal/ledger"
)

//go:embed web/*
var webFiles embed.FS

type Server struct {
	DB     *sql.DB
	Auth   *auth.Service
	Ledger *ledger.Ledger
}

func NewServer(db *sql.DB, authService *auth.Service, ledgerService *ledger.Ledger) *Server {
	return &Server{
		DB:     db,
		Auth:   authService,
		Ledger: ledgerService,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	// ۱. روت بررسی سلامت جهت تست Railway Healthcheck
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","version":"11.0.0","brics_node":"active"}`))
	})

	// ۲. مسیرهای اپلیکیشن
	mux.HandleFunc("/app/login", s.handleAppLogin)
	mux.HandleFunc("/app/register", s.handleAppRegister)

	// ۳. ساب‌سیستم فایل‌های وب و لندینگ‌پیج
	subFS, err := fs.Sub(webFiles, "web")
	if err != nil {
		panic(fmt.Sprintf("خطا در بارگذاری قالب‌ها: %v", err))
	}
	fileServer := http.FileServer(http.FS(subFS))

	// ۴. ریدایرکت خودکار یا تحویل صفحه اصلی بر اساس زبان
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			// تشخیص ترجیح اولیه مرورگر یا پیش‌فرض فارسی/انگلیسی
			acceptLang := r.Header.Get("Accept-Language")
			if strings.Contains(strings.ToLower(acceptLang), "fa") {
				http.Redirect(w, r, "/fa/", http.StatusFound)
				return
			}
			http.Redirect(w, r, "/en/", http.StatusFound)
			return
		}
		fileServer.ServeHTTP(w, r)
	})

	return s.corsMiddleware(mux)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleAppLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(renderAuthPage("ورود به درگاه BRICS Pay", "Login to BRICS Pay", "/app/login")))
}

func (s *Server) handleAppRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(renderAuthPage("افتتاح حساب و ثبت‌نام درگاه", "Register BricsPay Account", "/app/register")))
}

func renderAuthPage(titleFa, titleEn, action string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="fa" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s | BRICS Pay</title>
    <style>
        :root {
            --bg-dark: #070B19;
            --card-bg: rgba(16, 24, 48, 0.95);
            --primary-blue: #0066FF;
            --accent-gold: #D4AF37;
            --text-light: #F4F7FC;
        }
        body {
            margin: 0;
            background: radial-gradient(circle at 50%% 20%%, #121E42 0%%, var(--bg-dark) 100%%);
            font-family: system-ui, -apple-system, sans-serif;
            color: var(--text-light);
            display: flex;
            align-items: center;
            justify-content: center;
            min-height: 100vh;
        }
        .auth-card {
            background: var(--card-bg);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 16px;
            padding: 40px;
            width: 100%%;
            max-width: 420px;
            box-shadow: 0 20px 40px rgba(0, 0, 0, 0.6);
            backdrop-filter: blur(10px);
        }
        .logo { font-size: 26px; font-weight: 800; text-align: center; margin-bottom: 24px; color: #FFF; }
        .logo span { color: var(--accent-gold); }
        .title { font-size: 18px; margin-bottom: 24px; text-align: center; color: #94A3B8; }
        .form-group { margin-bottom: 18px; }
        label { display: block; font-size: 13px; margin-bottom: 8px; color: #CBD5E1; }
        input {
            width: 100%%;
            padding: 12px;
            border-radius: 8px;
            border: 1px solid #334155;
            background: #0B132B;
            color: #FFF;
            box-sizing: border-box;
            outline: none;
        }
        input:focus { border-color: var(--primary-blue); }
        button {
            width: 100%%;
            padding: 13px;
            background: linear-gradient(135deg, #0066FF 0%%, #0044AA 100%%);
            border: none;
            color: white;
            font-weight: 600;
            border-radius: 8px;
            cursor: pointer;
            font-size: 15px;
            margin-top: 10px;
        }
        .footer-links { margin-top: 20px; text-align: center; font-size: 13px; }
        .footer-links a { color: var(--accent-gold); text-decoration: none; }
    </style>
</head>
<body>
    <div class="auth-card">
        <div class="logo">BRICS<span>PAY</span></div>
        <div class="title">%s / %s</div>
        <form action="%s" method="POST">
            <div class="form-group">
                <label>شناسه کاربری / ایمیل / Username</label>
                <input type="text" name="username" required>
            </div>
            <div class="form-group">
                <label>رمز عبور / Password</label>
                <input type="password" name="password" required>
            </div>
            <button type="submit">تایید و ادامه</button>
        </form>
        <div class="footer-links">
            <a href="/">بازگشت به صفحه اصلی / Back to Home</a>
        </div>
    </div>
</body>
</html>`, titleFa, titleFa, titleEn, action)
}
