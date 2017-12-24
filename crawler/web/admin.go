package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/johnstcn/freshcomics/common/log"
	"github.com/johnstcn/freshcomics/crawler/util"
	"github.com/johnstcn/freshcomics/common/store"
)

type Admin struct {
	mux.Router
	store store.Store
	tpl *template.Template
}

func NewAdmin(s store.Store) *Admin {
	a := &Admin{
		store: s,
	}
	a.initTemplates()
	a.initRoutes()
	return a
}

// Shows a list of SiteDefs
func (a *Admin) siteDefsHandler(resp http.ResponseWriter, req *http.Request) {
	defs, err := a.store.GetAllSiteDefs(true)
	if err != nil {
		log.Error(err)
	}
	//tpl := getTemplate("admin_index.gohtml")
	err = a.tpl.ExecuteTemplate(resp, "admin_index", &defs)
	if err != nil {
		log.Error("could not execute template:", err)
	}
}

type detailsResponse struct {
	SiteDef *store.SiteDef
	Updates *[]store.SiteUpdate
	Events *[]store.CrawlEvent
	Success bool
	Message string
}

func (a *Admin) newSiteDefHandler(resp http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		name := req.PostFormValue("name")
		def, err := a.store.CreateSiteDef()
		if err != nil {
			log.Error(err)
		}
		def.Name = name
		err = a.store.SaveSiteDef(def)
		if err != nil {
			log.Error(err)
		}
		rdir := fmt.Sprintf("/sitedef/%d", def.ID)
		log.Info("Redirecting to", rdir)
		http.Redirect(resp, req, rdir, 302)
	} else {
		http.Redirect(resp, req, "/", 302)
	}
}

func (a *Admin) detailsHandler(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Error(err)
		http.Redirect(resp, req, "/", 500)
	}
	def, err := a.store.GetSiteDefByID(int64(id))
	if err != nil {
		log.Error(err)
		http.Redirect(resp, req, "/", 404)
	}
	updates, err := a.store.GetSiteUpdatesBySiteDefID(def.ID, 10)
	if err != nil {
		log.Error(err)
	}
	events, err := a.store.GetCrawlEventsBySiteDefID(def.ID, 10)
	if err != nil {
		log.Error(err)
	}
	r := &detailsResponse{
		SiteDef: def,
		Updates: updates,
		Events: events,
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
		err := a.store.SaveSiteDef(def)
		if err != nil {
			r.Success = false
			r.Message = err.Error()
			log.Error(err)
		}
	}
	err = a.tpl.ExecuteTemplate(resp, "admin_details", r)
	if err != nil {
		log.Error("could not execute template:", err)
	}
}

func (a *Admin) testHandler(resp http.ResponseWriter, req *http.Request) {
	sd := store.SiteDef{
		ID:            0,
		Name:          req.PostFormValue("name"),
		Active:        false,
		NSFW:          false,
		StartURL:      req.PostFormValue("starturl"),
		LastCheckedAt: time.Time{},
		URLTemplate:   req.PostFormValue("urltemplate"),
		RefXpath:      req.PostFormValue("refxpath"),
		RefRegexp:     req.PostFormValue("refregexp"),
		TitleXpath:    req.PostFormValue("titlexpath"),
		TitleRegexp:   req.PostFormValue("titleregexp"),
	}
	res := util.TestCrawl(&sd)
	enc := json.NewEncoder(resp)
	enc.SetIndent("", "\t")
	enc.Encode(res)
}

func (a *Admin) forceCrawlHandler(resp http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		req.ParseForm()
		defId, _ := strconv.Atoi(req.Form.Get("id"))
		sd, err := a.store.GetSiteDefByID(int64(defId))
		if err != nil {
			log.Error("unknown sitedef id:", err)
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}
		a.store.SetSiteDefLastChecked(sd, time.Unix(0, 0))
		resp.WriteHeader(http.StatusOK)
		log.Info("manual crawl initiated:", sd.Name)
	} else {
		resp.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *Admin) siteDefEventsHandler(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	defId, _ := strconv.Atoi(vars["id"])
	events, err := a.store.GetCrawlEventsBySiteDefID(int64(defId), -1)
	if err != nil {
		log.Error(err)
	}
	enc := json.NewEncoder(resp)
	enc.SetIndent("", "\t")
	enc.Encode(events)
}

func (a *Admin) siteDefUpdatesHandler(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	defId, _ := strconv.Atoi(vars["id"])
	updates, err := a.store.GetSiteUpdatesBySiteDefID(int64(defId), -1)
	if err != nil {
		log.Error(err)
	}
	enc := json.NewEncoder(resp)
	enc.SetIndent("", "\t")
	enc.Encode(updates)
}

func (a *Admin) eventsHandler(resp http.ResponseWriter, req *http.Request) {
	limit, err := strconv.Atoi(req.FormValue("limit"))
	if err != nil {
		limit = 100
	}
	events, _ := a.store.GetCrawlEvents(limit)
	enc := json.NewEncoder(resp)
	enc.SetIndent("", "\t")
	enc.Encode(events)
}


func (a *Admin) initRoutes() {
	a.HandleFunc("/", a.siteDefsHandler)
	a.HandleFunc("/sitedef/{id:[0-9]+}", a.detailsHandler)
	a.HandleFunc("/sitedef/{id:[0-9]+}/events", a.siteDefEventsHandler)
	a.HandleFunc("/sitedef/{id:[0-9]+}/updates", a.siteDefUpdatesHandler)
	a.HandleFunc("/sitedef/new", a.newSiteDefHandler)
	a.HandleFunc("/sitedef/test", a.testHandler)
	a.HandleFunc("/sitedef/crawl", a.forceCrawlHandler)
	a.HandleFunc("/events", a.eventsHandler)
}

func (a *Admin) initTemplates() {
	fm := template.FuncMap{
		"datetime": func(t time.Time) string {
			return t.Format("2006-01-02T15:04:05")
		},
	}
	a.tpl = template.New("").Funcs(fm)
	for _, an := range AssetNames() {
		a.tpl.Parse(string(MustAsset(an)))
	}
}
