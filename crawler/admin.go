package crawler

import (
	"os"
	"github.com/jinzhu/gorm"
	"fmt"
	"html/template"
	"net/http"
	"log"
	"encoding/json"
	"time"
)

var tpl *template.Template

type IndexHandler struct {
	db *gorm.DB
}

func (h *IndexHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	defs := make([]SiteDef, 0, 0)
	h.db.Find(&defs)
	err := tpl.ExecuteTemplate(resp, "admin_index.gohtml", &defs)
	if err != nil {
		log.Println("could not execute template:", err)
	}
}

type DetailsResponse struct {
	SiteDef *SiteDef
	Success bool
	Message string
}

type DetailsHandler struct {
	db *gorm.DB
}

func (h *DetailsHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get("id")
	def := &SiteDef{}
	h.db.First(&def, id)
	r := &DetailsResponse{
		SiteDef: def,
		Success: true,
		Message: "",
	}
	if req.Method == http.MethodPost {
		r.Message = "Updated successfully."
		def.Name = req.PostFormValue("name")
		def.StartURL = req.PostFormValue("starturl")
		def.RefRegexp = req.PostFormValue("refregexp")
		def.PagTemplate = req.PostFormValue("pagtemplate")
		def.TitleXpath = req.PostFormValue("titlexpath")
		def.TitleRegexp = req.PostFormValue("titleregexp")
		def.DateXpath = req.PostFormValue("datexpath")
		def.DateRegexp = req.PostFormValue("dateregexp")
		def.DateFormat = req.PostFormValue("dateformat")
		def.SetLastChecked(req.PostFormValue("lastchecked"))
		err := h.db.Save(def).Error
		if err != nil {
			r.Success = false
			r.Message = err.Error()
		}
	}
	err := tpl.ExecuteTemplate(resp, "admin_details.gohtml", r)
	if err != nil {
		log.Println("could not execute template:", err)
	}
}

type NewSiteDefHandler struct {
	db *gorm.DB
}

func (h *NewSiteDefHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var err error
	var def SiteDef
	r := &DetailsResponse{
		SiteDef: &def,
		Success: true,
		Message: "",
	}
	if req.Method == http.MethodPost {
		def, err := NewSiteDef(
			req.PostFormValue("name"),
			req.PostFormValue("starturl"),
			req.PostFormValue("pagtemplate"),
			req.PostFormValue("refxpath"),
			req.PostFormValue("refregexp"),
			req.PostFormValue("titlexpath"),
			req.PostFormValue("titleregexp"),
			req.PostFormValue("datexpath"),
			req.PostFormValue("dateregexp"),
			req.PostFormValue("dateformat"),
		)
		r.SiteDef = def
		if err != nil {
			r.Message = err.Error()
		} else {
			h.db.Save(def)
			r.Success = true
			r.Message = "Created successfully."
			sdUrl := fmt.Sprintf("/sitedef?id=%d", def.ID)
			http.Redirect(resp, req, sdUrl, 302)
		}
	}
	err = tpl.ExecuteTemplate(resp, "admin_new.gohtml", r)
	if err != nil {
		log.Println("could not execute template:", err)
	}
}

func TestHandler (resp http.ResponseWriter, req *http.Request) {
	sd, err := NewSiteDef(
		req.PostFormValue("name"),
		req.PostFormValue("starturl"),
		req.PostFormValue("pagtemplate"),
		req.PostFormValue("refxpath"),
		req.PostFormValue("refregexp"),
		req.PostFormValue("titlexpath"),
		req.PostFormValue("titleregexp"),
		req.PostFormValue("datexpath"),
		req.PostFormValue("dateregexp"),
		req.PostFormValue("dateformat"),
	)
	if err != nil {
		enc := json.NewEncoder(resp)
		enc.Encode(struct {
			Error   string
			NextURL string
			Result struct{}
		}{
			err.Error(),
			 "",
			struct{}{},
		})
	} else {
		res := TestCrawl(sd)
		enc := json.NewEncoder(resp)
		enc.SetIndent("", "\t")
		enc.Encode(res)
	}
}

func ServeAdmin(db *gorm.DB) {
	defer db.Close()
	listenAddress := os.Getenv("HOSTPORT")
	log.Printf("Listening on %s", listenAddress)
	http.Handle("/", &IndexHandler{db: db})
	http.Handle("/sitedef", &DetailsHandler{db: db})
	http.Handle("/newsitedef", &NewSiteDefHandler{db: db})
	http.HandleFunc("/test", TestHandler)
	initTemplates()
	log.Fatal(http.ListenAndServe(listenAddress, http.DefaultServeMux))
}

func initTemplates() {
	fm := template.FuncMap{
		"datetime": func(ts int64) string {
			t := time.Unix(ts, 0).UTC()
			return t.Format("2006-01-02T15:04:05")
		},
		"duration": func(ts int64) string {
			mins := ts / 60
			secs := ts % 60
			return fmt.Sprintf("%dm %ds", mins, secs)
		},
	}
	tpl = template.Must(template.New("").Funcs(fm).ParseGlob("templates/admin_*.gohtml"))
}