package web

import (
	"fmt"
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/johnstcn/freshcomics/common/log"
	"github.com/johnstcn/freshcomics/frontend/models"
)

var tpl *template.Template

func indexHandler(resp http.ResponseWriter, req *http.Request) {
	dao := models.GetDAO()
	comics, _ := dao.GetComics()
	enc := json.NewEncoder(resp)
	enc.SetIndent("", "\t")
	enc.Encode(comics)
}

func ServeFrontend(host, port string) {
	listenAddress := fmt.Sprintf("%s:%s", host, port)
	log.Info("Listening on", listenAddress)
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(listenAddress, http.DefaultServeMux)
}

func humanDuration(t time.Time) string {
	suffix := "ago"
	now := time.Now()
	delta := now.Sub(t)
	if t.After(now) {
		// never know when time will get weird
		suffix = "from now"
	}
	return fmt.Sprintf("%v %s", delta, suffix)
}

func initTemplates() {
	fm := template.FuncMap{
		"humanDuration": humanDuration,
	}
	tpl = template.New("").Funcs(fm)

}