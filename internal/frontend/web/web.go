package web

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"golang.org/x/exp/slog"

	"github.com/johnstcn/freshcomics/internal/frontend/config"
	"github.com/johnstcn/freshcomics/internal/store"
)

type frontend struct {
	mux.Router
	store  store.Store
	comics []store.Comic
	tpl    *template.Template
	log    *slog.Logger
}

func NewFrontend(s store.Store) *frontend {
	comics := make([]store.Comic, 0)
	f := &frontend{
		store:  s,
		comics: comics,
	}

	f.initTemplates()
	f.initRoutes()
	go f.UpdateComics()

	return f
}

func (f *frontend) UpdateComics() {
	interval := time.Duration(config.Cfg.RefreshIntervalSecs) * time.Second
	for {
		f.log.Debug("updating comic list")
		newComics, err := f.store.GetComics()
		if err != nil {
			f.log.Error("update comic list", "err", err)
		} else {
			f.comics = newComics
		}
		time.Sleep(interval)
	}
}

func (f *frontend) indexHandler(resp http.ResponseWriter, req *http.Request) {
	data := struct {
		Comics []store.Comic
	}{
		Comics: f.comics,
	}
	err := f.tpl.ExecuteTemplate(resp, "frontend_index", &data)
	if err != nil {
		f.log.Error("execute template frontend_index", "err", err)
	}
}

func (f *frontend) redirectHandler(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	if vars["id"] == "" {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	updateId, err := strconv.Atoi(vars["id"])
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	url, err := f.store.Redirect(store.SiteUpdateID(updateId))
	if err != nil {
		http.NotFound(resp, req)
		return
	}
	http.Redirect(resp, req, url, http.StatusMovedPermanently)
}

func (f *frontend) cssHandler(resp http.ResponseWriter, req *http.Request) {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	cssAsset, err := Asset(strings.TrimLeft(req.RequestURI, "/"))
	if err != nil {
		f.log.Error("css asset", "err", err)
	}
	minified, err := m.Bytes("text/css", cssAsset)
	if err != nil {
		f.log.Error("minify css", "err", err)
	}
	resp.Header().Add("Content-Type", "text/css")
	resp.Write(minified)
}

func (f *frontend) initTemplates() {
	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	fm := template.FuncMap{
		"humanDuration": humanize.Time,
	}
	tpl := template.New("").Funcs(fm)
	for _, an := range AssetNames() {
		// only minify html here
		if !strings.HasSuffix(an, ".gohtml") {
			continue
		}
		s := MustAsset(an)
		ms, err := m.Bytes("text/html", s)
		if err != nil {
			f.log.Error("minify template", "asset", an, "err", err)
		}
		tpl.Parse(string(ms))
	}
}

func (f *frontend) initRoutes() {
	f.HandleFunc("/", f.indexHandler)
	f.HandleFunc("/view/{id:[0-9]+}", f.redirectHandler)
	f.HandleFunc("/style.css", f.cssHandler)
}
