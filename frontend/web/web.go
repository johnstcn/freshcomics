package web

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/css"

	"github.com/johnstcn/freshcomics/common/log"
	"github.com/johnstcn/freshcomics/frontend/config"
	"github.com/johnstcn/freshcomics/frontend/models"
)

var tpl *template.Template
var comics *[]models.Comic
var rtr *mux.Router

func updateComics() {
	interval := time.Duration(config.Cfg.RefreshIntervalSecs) * time.Second
	for {
		log.Info("updating comic list")
		dao := models.GetDAO()
		newComics, err := dao.GetComics()
		if err != nil {
			log.Error("could not update comic list:", err)
		} else {
			comics = newComics
		}
		time.Sleep(interval)
	}
}

func indexHandler(resp http.ResponseWriter, req *http.Request) {
	data := struct{
		Comics *[]models.Comic
	}{
		Comics: comics,
	}
	err := tpl.ExecuteTemplate(resp, "frontend_index", &data)
	if err != nil {
		log.Error("could not execute frontend_index template:", err)
	}
}

func redirectHandler(resp http.ResponseWriter, req *http.Request) {
	dao := models.GetDAO()
	vars := mux.Vars(req)
	updateId := vars["id"]
	url, err := dao.GetRedirectURL(updateId)
	if err != nil {
		http.NotFound(resp, req)
		return
	}
	http.Redirect(resp, req, url, 301)
}

func cssHandler(resp http.ResponseWriter, req *http.Request) {
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

func ServeFrontend(host string, port int) {
	listenAddress := fmt.Sprintf("%s:%d", host, port)
	log.Info("Listening on", listenAddress)
	http.ListenAndServe(listenAddress, rtr)
}

func initTemplates() {
	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	fm := template.FuncMap{
		"humanDuration": humanize.Time,
	}
	tpl = template.New("").Funcs(fm)
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

func initRoutes() {
	rtr = mux.NewRouter()
	rtr.HandleFunc("/", indexHandler)
	rtr.HandleFunc("/view/{id:[0-9]+}", redirectHandler)
	rtr.HandleFunc("/style.css", cssHandler)
}

func init() {
	go updateComics()
	initTemplates()
	initRoutes()
}