package api

import (
	"database/sql"
	"embed"
	"encoding/json"
	"net/http"
)

//go:embed static/index.html
var content embed.FS

type Server struct {
	DB *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{DB: db}
}

func (s *Server) HandleRoot(w http.ResponseWriter, r *http.Request) {
	file, err := content.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(file)
}

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) HandleAccounts(w http.ResponseWriter, r *http.Request) {
    // پیاده‌سازی حساب‌ها (اگر منطق خاصی دارید اینجا اضافه کنید)
    w.Write([]byte("Accounts endpoint"))
}

func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
    // پیاده‌سازی تراکنش‌ها (اگر منطق خاصی دارید اینجا اضافه کنید)
    w.Write([]byte("Transactions endpoint"))
}
