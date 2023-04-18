package app

import (
	"embed"
	"io/fs"
	"net/http"
	"sync/atomic"

	"golang.org/x/exp/slog"
)

//go:embed all:dist
var distFS embed.FS

//go:embed dist/index.html
var indexHTML []byte

type Deps struct {
	Mux    *http.ServeMux
	Logger *slog.Logger
}

func New(deps Deps) {
	assetsFS, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	logFS(deps.Logger, assetsFS, "assets")
	deps.Mux.Handle("/assets/", http.FileServer(http.FS(assetsFS)))
	deps.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(indexHTML)
	})
}

func logFS(log *slog.Logger, tgtFS fs.FS, prefix string) {
	var embedCount atomic.Int64
	fs.WalkDir(tgtFS, prefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		embedCount.Add(1)
		switch d.Type() {
		case fs.ModeDir:
			log.Info("app embed", "type", "d", "path", path)
		default:
			log.Info("app embed", "type", "f", "path", path)
		}
		return nil
	})
	log.Info("embed count", "total", embedCount.Load())
}
