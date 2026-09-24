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
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	htmlData, err := content.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(htmlData)
}

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) HandleAccounts(w http.ResponseWriter, r *http.Request) {
    // Insert your existing account retrieval logic here
}

func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
    // Insert your existing transaction processing logic here
}
