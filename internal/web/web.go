package web

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"golang.org/x/exp/slog"

	"github.com/johnstcn/freshcomics/internal/store"
)

type frontend struct {
	*http.ServeMux
	store store.Store
	log   *slog.Logger
	tpl   *template.Template
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

	f.initTemplates()
	f.initRoutes()

	return f
}

func (f *frontend) indexHandler(w http.ResponseWriter, r *http.Request) {
	comics, err := f.store.GetComics()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("get comics: %s", err.Error())))
		return
	}
	data := struct {
		Comics []store.Comic
	}{
		Comics: comics,
	}
	err = f.tpl.ExecuteTemplate(w, "frontend_index", &data)
	if err != nil {
		f.log.Error("execute template frontend_index", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("execute template frontend_index: %s", err.Error())))
		return
	}
}

func (f *frontend) redirectHandler(w http.ResponseWriter, r *http.Request) {
	idVal, ok := r.URL.Query()["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("parameter id is required"))
		return
	}
	if len(idVal) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("parameter id must have a value"))
		return
	}
	updateId, err := strconv.Atoi(idVal[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("parameter id must be an integer"))
		return
	}
	url, err := f.store.Redirect(store.SiteUpdateID(updateId))
	if err != nil {
		http.NotFound(w, r)
		w.Write([]byte("no update found with id " + idVal[0]))
		return
	}
	http.Redirect(w, r, url, http.StatusMovedPermanently)
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
