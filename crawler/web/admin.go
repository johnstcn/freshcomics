package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/johnstcn/freshcomics/crawler/log"
	"github.com/johnstcn/freshcomics/crawler/models"
	"github.com/johnstcn/freshcomics/crawler/util"
)

var tpl *template.Template

// Shows a list of SiteDefs
func siteDefsHandler(resp http.ResponseWriter, req *http.Request) {
	dao := models.GetDAO()
	defs, err := dao.GetAllSiteDefs(true)
	if err != nil {
		log.Error.Println(err)
	}
	err = tpl.ExecuteTemplate(resp, "admin_index.gohtml", &defs)
	if err != nil {
		log.Error.Println("could not execute template:", err)
	}
}

type detailsResponse struct {
	SiteDef *models.SiteDef
	Success bool
	Message string
}

func newSiteDefHandler(resp http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		dao := models.GetDAO()
		name := req.PostFormValue("name")
		def, err := dao.CreateSiteDef()
		if err != nil {
			log.Error.Println(err)
		}
		def.Name = name
		err = dao.SaveSiteDef(def)
		if err != nil {
			log.Error.Println(err)
		}
		rdir := fmt.Sprintf("/sitedef?id=%d", def.ID)
		log.Info.Println("Redirecting to", rdir)
		http.Redirect(resp, req, rdir, 302)
	} else {
		http.Redirect(resp, req, "/", 302)
	}
}

func detailsHandler(resp http.ResponseWriter, req *http.Request) {
	dao := models.GetDAO()
	id, err := strconv.Atoi(req.URL.Query().Get("id"))
	if err != nil {
		log.Error.Println(err)
		http.Redirect(resp, req, "/", 500)
	}
	def, err := dao.GetSiteDefByID(int64(id))
	if err != nil {
		log.Error.Println(err)
		http.Redirect(resp, req, "/", 404)
	}
	r := &detailsResponse{
		SiteDef: def,
		Success: true,
		Message: "",
	}
	if req.Method == http.MethodPost {
		r.Message = "Updated successfully."
		def.Name = req.PostFormValue("name")
		def.Active, _ = strconv.ParseBool(req.PostFormValue("active"))
		def.NSFW, _ = strconv.ParseBool(req.PostFormValue("nsfw"))
		def.StartURL = req.PostFormValue("starturl")
		def.URLTemplate = req.PostFormValue("urltemplate")
		def.RefXpath = req.PostFormValue("refxpath")
		def.RefRegexp = req.PostFormValue("refregexp")
		def.TitleXpath = req.PostFormValue("titlexpath")
		def.TitleRegexp = req.PostFormValue("titleregexp")
		def.DateXpath = req.PostFormValue("datexpath")
		def.DateRegexp = req.PostFormValue("dateregexp")
		def.DateFormat = req.PostFormValue("dateformat")
		def.NextPageXpath = req.PostFormValue("nextpagexpath")
		def.NextPageRegexp = req.PostFormValue("nextpageregexp")
		def.LastChecked, _ = time.Parse("2006-01-02T15:04:05", req.PostFormValue("lastchecked"))
		err := dao.SaveSiteDef(def)
		if err != nil {
			r.Success = false
			r.Message = err.Error()
			log.Error.Println(err)
		}
	}
	err = tpl.ExecuteTemplate(resp, "admin_details.gohtml", r)
	if err != nil {
		log.Error.Println("could not execute template:", err)
	}
}

func testHandler(resp http.ResponseWriter, req *http.Request) {
	sd := models.SiteDef{
		ID:          0,
		Name:        req.PostFormValue("name"),
		Active:      false,
		NSFW:        false,
		StartURL:    req.PostFormValue("starturl"),
		LastChecked: time.Time{},
		URLTemplate: req.PostFormValue("urltemplate"),
		RefXpath:    req.PostFormValue("refxpath"),
		RefRegexp:   req.PostFormValue("refregexp"),
		TitleXpath:  req.PostFormValue("titlexpath"),
		TitleRegexp: req.PostFormValue("titleregexp"),
		DateXpath:   req.PostFormValue("datexpath"),
		DateRegexp:  req.PostFormValue("dateregexp"),
		DateFormat:  req.PostFormValue("dateformat"),
		NextPageXpath: req.PostFormValue("nextpagexpath"),
		NextPageRegexp: req.PostFormValue("nextpageregexp"),
	}
	res := util.TestCrawl(&sd)
	enc := json.NewEncoder(resp)
	enc.SetIndent("", "\t")
	enc.Encode(res)
}

func ServeAdmin() {
	initTemplates()
	listenAddress := os.Getenv("HOSTPORT")
	http.HandleFunc("/", siteDefsHandler)
	http.HandleFunc("/sitedef", detailsHandler)
	http.HandleFunc("/sitedef/new", newSiteDefHandler)
	http.HandleFunc("/sitedef/test", testHandler)
	log.Info.Printf("Listening on %s", listenAddress)
	log.Info.Fatal(http.ListenAndServe(listenAddress, http.DefaultServeMux))
}

func initTemplates() {
	fm := template.FuncMap{
		"datetime": func(t time.Time) string {
			return t.Format("2006-01-02T15:04:05")
		},
	}
	tpl = template.Must(template.New("").Funcs(fm).ParseGlob("templates/admin_*.gohtml"))
}