package web

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/johnstcn/freshcomics/common/log"
	"github.com/johnstcn/freshcomics/frontend/models"
	"github.com/dustin/go-humanize"
	"time"
)

var tpl *template.Template
var comics *[]models.Comic

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
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(listenAddress, http.DefaultServeMux)
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

func init() {
	initTemplates()
	go updateComics()
}