package web

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/johnstcn/freshcomics/common/log"
	"github.com/johnstcn/freshcomics/frontend/models"
	"github.com/dustin/go-humanize"
	"time"
	"github.com/gorilla/mux"
)

var tpl *template.Template
var comics *[]models.Comic
var rtr *mux.Router

func updateComics() {
	for {
		log.Info("updating comic list")
		dao := models.GetDAO()
		newComics, err := dao.GetComics()
		if err != nil {
			log.Error("could not update comic list:", err)
		} else {
			comics = newComics
		}
		time.Sleep(10 * time.Second)
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

func ServeFrontend(host, port string) {
	listenAddress := fmt.Sprintf("%s:%s", host, port)
	log.Info("Listening on", listenAddress)
	http.ListenAndServe(listenAddress, rtr)
}

func initTemplates() {
	fm := template.FuncMap{
		"humanDuration": humanize.Time,
	}
	tpl = template.New("").Funcs(fm)
	for _, an := range AssetNames() {
		tpl.Parse(string(MustAsset(an)))
		log.Debug("template init:", an)
	}
}

func initRoutes() {
	rtr = mux.NewRouter()
	rtr.HandleFunc("/", indexHandler)
}

func init() {
	go updateComics()
	initTemplates()
	initRoutes()
}