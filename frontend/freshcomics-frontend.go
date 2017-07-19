package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/johnstcn/freshcomics/common/log"
	"github.com/johnstcn/freshcomics/frontend/models"
)


func comicsHandler(resp http.ResponseWriter, req *http.Request) {
	dao := models.GetDAO()
	comics, _ := dao.GetComics()
	enc := json.NewEncoder(resp)
	enc.SetIndent("", "\t")
	enc.Encode(comics)
}

func main() {
	listenAddress := os.Getenv("HOSTPORT")
	log.Info("Listening on", listenAddress)
	http.HandleFunc("/", comicsHandler)
	http.ListenAndServe(listenAddress, http.DefaultServeMux)
}
