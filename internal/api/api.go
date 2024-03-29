package api

import (
	"encoding/json"
	"net/http"

	"golang.org/x/exp/slog"

	"github.com/johnstcn/freshcomics/internal/store"
)

type handler struct {
	*http.ServeMux
	store store.Store
	log   *slog.Logger
}

type Deps struct {
	Mux    *http.ServeMux
	Store  store.Store
	Logger *slog.Logger
}

func New(deps Deps) {
	f := &handler{
		ServeMux: deps.Mux,
		store:    deps.Store,
		log:      deps.Logger,
	}

	f.HandleFunc("/api/comics/", f.listComics)
}

type ListComicsResponse struct {
	Data  []store.Comic `json:"data"`
	Error string        `json:"error"`
}

func (h *handler) listComics(w http.ResponseWriter, r *http.Request) {
	resp := ListComicsResponse{
		Data:  []store.Comic{},
		Error: "",
	}
	code := http.StatusOK
	data, err := h.store.GetComics()
	if err != nil {
		h.log.Error("get data from store", "err", err, "handler", "listComics")
		code = http.StatusInternalServerError
		resp.Error = err.Error()
	} else {
		resp.Data = data
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.log.Error("write response", "err", err, "handler", "listComics")
	}
}
