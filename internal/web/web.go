package web

import (
	"encoding/json"
	"net/http"

	"golang.org/x/exp/slog"

	"github.com/johnstcn/freshcomics/internal/store"
)

type frontend struct {
	*http.ServeMux
	store store.Store
	log   *slog.Logger
}

type Deps struct {
	Mux    *http.ServeMux
	Store  store.Store
	Logger *slog.Logger
}

func New(deps Deps) *frontend {
	f := &frontend{
		ServeMux: deps.Mux,
		store:    deps.Store,
		log:      deps.Logger,
	}

	f.HandleFunc("/api/comics/", f.listComics)

	return f
}

type ListComicsResponse struct {
	Data  []store.Comic `json:"data"`
	Error string        `json:"error"`
}

func (f *frontend) listComics(w http.ResponseWriter, r *http.Request) {
	resp := ListComicsResponse{
		Data:  []store.Comic{},
		Error: "",
	}
	code := http.StatusOK
	data, err := f.store.GetComics()
	if err != nil {
		f.log.Error("get data from store", "err", err, "handler", "listComics")
		code = http.StatusInternalServerError
		resp.Error = err.Error()
	} else {
		resp.Data = data
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		f.log.Error("write response", "err", err, "handler", "listComics")
	}
}
