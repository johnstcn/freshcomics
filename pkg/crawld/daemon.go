package crawld

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/johnstcn/freshcomics/internal/store"
	"github.com/johnstcn/gocrawl/pkg/crawl"

	"github.com/pkg/errors"
)

var errNoPendingWork = errors.New("no pending work")

func New(cfg Config) (*CrawlDaemon, error) {
	opts := crawl.CrawlerOpts{
		Timeout: time.Duration(cfg.FetchTimeoutSecs) * time.Second,
	}
	pgstore, err := store.NewPGStore(cfg.DSN)
	if err != nil {
		return nil, err
	}
	stopScheduler := make(chan bool)
	stopWorker := make(chan bool)
	return &CrawlDaemon{
		stopScheduler: stopScheduler,
		stopWorker:    stopWorker,
		now:           time.Now,
		exit:          os.Exit,
		crawler:       crawl.New(opts),
		config:        cfg,
		siteDefs:      pgstore,
		siteUpdates:   pgstore,
		crawlInfos:    pgstore,
	}, nil
}

type CrawlDaemon struct {
	stopScheduler chan bool
	stopWorker    chan bool
	now           func() time.Time
	exit          func(status int)
	crawler       crawl.Crawler
	config        Config
	siteDefs      store.SiteDefStore
	siteUpdates   store.SiteUpdateStore
	crawlInfos    store.CrawlInfoStore
}

func (d *CrawlDaemon) Run() error {
	ch := make(chan os.Signal)
	signal.Notify(ch)
	go d.handleSignals(ch)
	go d.scheduleWorkForever()
	go d.doWorkForever()
	select {}
	return nil
}

func (d *CrawlDaemon) handleSignals(ch <-chan os.Signal) {
	for s := range ch {
		if s == syscall.SIGINT || s == syscall.SIGTERM {
			log.Printf("caught signal %s, exiting\n", s)
			d.stopWorker <- true
			d.stopScheduler <- true
			d.exit(1)
		} else {
			log.Printf("caught signal %s, ignoring\n", s)
		}
	}
}

func (d *CrawlDaemon) scheduleWorkForever() {
	for {
		select {
		case <-d.stopScheduler:
			log.Println("stopping scheduler")
			return
		case <-time.After(time.Duration(d.config.ScheduleIntervalSecs) * time.Second):
			if err := d.scheduleWorkOnce(); err != nil {
				log.Println(err)
			}
		}
	}
}

// TODO(cian): make this not terrible
func (d *CrawlDaemon) scheduleWorkOnce() error {
	pending, err := d.crawlInfos.GetPendingCrawlInfos()
	if err != nil {
		return errors.Wrap(err, "fetching pending work")
	}

	pendingIDs := make(map[store.SiteDefID]bool)
	for _, item := range pending {
		pendingIDs[item.SiteDefID] = true
	}

	defs, err := d.siteDefs.GetSiteDefs(false)
	if err != nil {
		return errors.Wrap(err, "fetching active sitedefs")
	}

	for _, def := range defs {
		if _, found := pendingIDs[def.ID]; found {
			log.Println("pending work already exists for site def: ", def.ID)
			continue
		}

		crawls, err := d.crawlInfos.GetCrawlInfo(def.ID)
		if err != nil {
			log.Println("fetching previous crawls for site def: ", def.ID)
			continue
		}

		nextScheduleTime := crawls[0].EndedAt.Add(time.Duration(d.config.CheckIntervalSecs) * time.Second)
		shouldSchedule := nextScheduleTime.After(d.now())

		if !shouldSchedule {
			continue
		}

		lastURL, err := d.siteDefs.GetLastURL(def.ID)
		if err != nil {
			log.Println("fetching last URL for site def: ", def.ID)
			continue
		}
		if _, err := d.crawlInfos.CreateCrawlInfo(def.ID, lastURL); err != nil {
			log.Println("scheduling work for site def: ", def.ID)
		}
	}

	return nil
}

func (d *CrawlDaemon) doWorkForever() {
	for {
		select {
		case <-d.stopWorker:
			log.Println("stopping worker")
			return
		case <-time.After(time.Duration(d.config.WorkPollIntervalSecs) * time.Second):
			item, err := d.getWorkOnce()
			if err == errNoPendingWork {
				log.Println(err)
				continue
			}
			if err != nil {
				log.Println("error fetching pending work:", err)
				continue
			}

			log.Println("got work:", item)
			d.doWorkOnce(item)
		}
	}
}

func (d *CrawlDaemon) getWorkOnce() (*store.CrawlInfo, error) {
	pending, err := d.crawlInfos.GetPendingCrawlInfos()
	if err != nil {
		return nil, err
	}

	if len(pending) == 0 {
		return nil, errNoPendingWork
	}

	return &pending[0], nil
}

func (d *CrawlDaemon) doWorkOnce(ci *store.CrawlInfo) {
	// TODO(cian): implement me
	var seen int
	var crawlErr error
	var currPage = ci.URL

	if err := d.crawlInfos.StartCrawlInfo(ci.ID); err != nil {
		log.Printf("marking crawl %d started: %v\n", ci.ID, err)
	} else {
		log.Printf("starting crawl %d page %s", ci.ID, currPage)
	}

	defer func() {
		if crawlErr != nil {
			log.Printf("crawl %d error on page: %s: %v\n", ci.ID, currPage, crawlErr)
		}

		if err := d.crawlInfos.EndCrawlInfo(ci.ID, crawlErr, seen); err != nil {
			log.Printf("marking crawl %d completed: %v\n", ci.ID, err)
		}
	}()

	def, err := d.siteDefs.GetSiteDef(ci.SiteDefID)
	if err != nil {
		log.Println("fetching sitedef", ci.SiteDefID, err)
		return
	}

	crawlJob := crawl.Job{
		Request: crawl.Request{
			URL:     currPage,
			Method:  http.MethodGet,
			Headers: map[string]string{"User-Agent": d.config.UserAgent},
			Body:    "",
		},
		Rules: []crawl.Rule{
			{
				Name:  "ref",
				XPath: def.RefXpath,
				Filters: []crawl.Filter{
					{
						Find:    def.RefXpath,
						Replace: "$1",
					},
				},
			},
			{
				Name:  "title",
				XPath: def.TitleXpath,
				Filters: []crawl.Filter{
					{
						Find:    def.TitleRegexp,
						Replace: "$1",
					},
				},
			},
		},
	}

	result, crawlErr := d.crawler.Crawl(crawlJob)
	if crawlErr != nil {
		return
	}

	refResult, found := result["ref"]
	if !found {
		crawlErr = errors.New("no output for ref rule")
		return
	}

	if refResult.Error != "" {
		crawlErr = errors.New(refResult.Error)
		return
	}

	if len(refResult.Values) < 1 {
		crawlErr = errors.New("no matches for ref Xpath")
		return
	}

	newRef := refResult.Values[0]
	currPage = fmt.Sprintf(def.URLTemplate, newRef)

	crawlJob.Request.URL = currPage

	for {
		result, crawlErr = d.crawler.Crawl(crawlJob)
		if crawlErr != nil {
			return
		}

		titleResult, found := result["title"]
		if !found {
			crawlErr = errors.New("no output for title rule")
			return
		}

		if titleResult.Error != "" {
			crawlErr = errors.New(titleResult.Error)
			return
		}

		if len(titleResult.Values) < 1 {
			crawlErr = errors.New("no matches for title Xpath")
			return
		}

		newTitle := titleResult.Values[0]

		newUpdate := store.SiteUpdate{
			SiteDefID: ci.SiteDefID,
			URL:       currPage,
			Ref:       newRef,
			Title:     newTitle,
			SeenAt:    d.now(),
		}

		if _, err := d.siteUpdates.CreateSiteUpdate(newUpdate); err != nil {
			log.Printf("failed to persist site update for crawl %d: %+v\n", ci.ID, newUpdate)
			return
		} else {
			seen += 1
		}

		refResult, found := result["ref"]
		if !found {
			crawlErr = errors.New("no output for ref rule")
			return
		}

		if refResult.Error != "" {
			crawlErr = errors.New(refResult.Error)
			return
		}

		if len(refResult.Values) < 1 {
			crawlErr = errors.New("no matches for ref Xpath")
			return
		}

		newRef = refResult.Values[0]
		currPage = fmt.Sprintf(def.URLTemplate, newRef)
	}
}
