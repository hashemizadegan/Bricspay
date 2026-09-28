package api

import (
	"embed"
	"io/fs"
	"net/http"
	// سایر ایمپورت‌ها...
)

//go:embed static/*
var staticFS embed.FS

// StaticFS دسترسی به ساب‌فولدر static را برای FileServer فراهم می‌کند
func StaticFS() http.FileSystem {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}
