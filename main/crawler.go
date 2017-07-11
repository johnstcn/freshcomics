package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/johnstcn/freshcomics/crawler"
	"fmt"
)

var tpl *template.Template

type IndexHandler struct {
	db *gorm.DB
}

func (h *IndexHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	defs := make([]crawler.SiteDef, 0, 0)
	h.db.Find(&defs)
	err := tpl.ExecuteTemplate(resp, "sitedef_index.gohtml", &defs)
	if err != nil {
		log.Println("could not execute template:", err)
	}
}

type DetailsResponse struct {
	SiteDef *crawler.SiteDef
	Success bool
	Message string
}

type DetailsHandler struct {
	db *gorm.DB
}

func (h *DetailsHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get("id")
	def := &crawler.SiteDef{}
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
	err := tpl.ExecuteTemplate(resp, "sitedef_details.gohtml", r)
	if err != nil {
		log.Println("could not execute template:", err)
	}
}

type NewSiteDefHandler struct {
	db *gorm.DB
}

func (h *NewSiteDefHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var err error
	var def crawler.SiteDef
	r := &DetailsResponse{
		SiteDef: &def,
		Success: true,
		Message: "",
	}
	if req.Method == http.MethodPost {
		def, err := crawler.NewSiteDef(
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
	err = tpl.ExecuteTemplate(resp, "sitedef_new.gohtml", r)
	if err != nil {
		log.Println("could not execute template:", err)
	}
}

func TestHandler (resp http.ResponseWriter, req *http.Request) {
	sd, err := crawler.NewSiteDef(
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
			Success bool
			Error   string
		}{
			false,
			err.Error(),
		})
	}
	res := crawler.TestCrawl(sd)
	enc := json.NewEncoder(resp)
	enc.Encode(res)
}

func serveAdmin() {
	db := GetDB()
	defer db.Close()
	listenAddress := os.Getenv("HOSTPORT")
	log.Printf("Listening on %s", listenAddress)
	http.Handle("/", &IndexHandler{db: db})
	http.Handle("/sitedef", &DetailsHandler{db: db})
	http.Handle("/newsitedef", &NewSiteDefHandler{db: db})
	http.HandleFunc("/test", TestHandler)
	log.Fatal(http.ListenAndServe(listenAddress, http.DefaultServeMux))
}

func GetDB() (*gorm.DB) {
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open("postgres", dsn)
	if err != nil {
		log.Fatalln("Error connecting to database:", err)
	}
	if os.Getenv("VERBOSE") == "1" {
		db.LogMode(true)
	}
	log.Println("Connected to database")

	db.AutoMigrate(crawler.SiteDef{}, crawler.SiteUpdate{})

	return db
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
	tpl = template.Must(template.New("").Funcs(fm).ParseGlob("templates/sitedef_*.gohtml"))
}

func init() {
	initTemplates()
}

func main() {
	db := GetDB()
	defer db.Close()
	go serveAdmin()
	for {
		tick := 1 * time.Second
		next := crawler.GetLastCheckedSiteDef(db)
		if next != nil {
			go crawler.Crawl(db, next)
		}
		time.Sleep(tick)
	}
}
