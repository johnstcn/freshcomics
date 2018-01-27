package web

import (
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strings"
	"regexp"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/css"

	"github.com/johnstcn/freshcomics/internal/common/log"
	"github.com/johnstcn/freshcomics/internal/store"
	"github.com/johnstcn/freshcomics/internal/frontend/config"
)

type frontend struct {
	mux.Router
	store store.Store
	comics *[]store.Comic
	tpl *template.Template
}

func NewFrontend(s store.Store) *frontend {
	comics := make([]store.Comic, 0)
	f := &frontend{
		store: s,
		comics: &comics,
	}

	f.initTemplates()
	f.initRoutes()
	go f.UpdateComics()

	return f
}

func (f *frontend) UpdateComics() {
	interval := time.Duration(config.Cfg.RefreshIntervalSecs) * time.Second
	for {
		log.Info("updating comic list")
		newComics, err := f.store.GetComics()
		if err != nil {
			log.Error("could not update comic list:", err)
		} else {
			f.comics = newComics
		}
		time.Sleep(interval)
	}
}

func (f *frontend) indexHandler(resp http.ResponseWriter, req *http.Request) {
	data := struct{
		Comics *[]store.Comic
	}{
		Comics: f.comics,
	}
	err := f.tpl.ExecuteTemplate(resp, "frontend_index", &data)
	if err != nil {
		log.Error("could not execute frontend_index template:", err)
	}
}

func (f *frontend) remoteIP(req *http.Request) (net.IP, error) {
	fwdHdr := req.Header.Get("X-Forwarded-For")
	log.Debug("X-Forwarded-For header:", fwdHdr)
	fwdAddr := regexp.MustCompile(`^\s*([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)`).FindString(fwdHdr)
	addr := net.ParseIP(fwdAddr)
	if fwdAddr == "" || addr == nil {
		return nil, errors.New(fmt.Sprintf("unable to parse X-Forwarded-For: %s", fwdHdr))
	}
	return addr, nil
}

func (f *frontend) redirectHandler(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	updateId := vars["id"]
	addr, err := f.remoteIP(req)
	if err != nil {
		log.Error(err)
	} else {
		err = f.store.RecordClick(updateId, addr)
		if err != nil {
			log.Error("error recording click:", err)
		}
	}
	url, err := f.store.GetRedirectURL(updateId)
	if err != nil {
		http.NotFound(resp, req)
		return
	}
	http.Redirect(resp, req, url, 301)
}

func (f *frontend) cssHandler(resp http.ResponseWriter, req *http.Request) {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	log.Debug(req.RequestURI)
	cssAsset, err := Asset(strings.TrimLeft(req.RequestURI, "/"))
	if err != nil {
		log.Error(err)
	}
	minified, err := m.Bytes("text/css", cssAsset)
	if err != nil {
		log.Error("unable to minify css:", err)
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
			log.Error("unable to minify template:", an)
		}
		tpl.Parse(string(ms))
		log.Debug("template init:", an)
	}
}

func (f *frontend) initRoutes() {
	f.HandleFunc("/", f.indexHandler)
	f.HandleFunc("/view/{id:[0-9]+}", f.redirectHandler)
	f.HandleFunc("/style.css", f.cssHandler)
}
