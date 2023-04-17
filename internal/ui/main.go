package ui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist/*
var staticFS embed.FS

type Deps struct {
	Mux *http.ServeMux
}

func New(deps Deps) http.Handler {
	h := http.FileServer(http.FS(fs.FS(staticFS)))
	deps.Mux.Handle("/static", h)
	return h
}
