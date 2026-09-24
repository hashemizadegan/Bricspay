package api

import (
	"embed"
	"net/http"
	"bricspayir/internal/ledger"
)

//go:embed static/index.html
var content embed.FS

type Server struct {
	Ledger *ledger.Ledger
}

func (s *Server) HandleRoot(w http.ResponseWriter, r *http.Request) {
	file, err := content.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(file)
}
