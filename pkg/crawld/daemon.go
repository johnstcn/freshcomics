package crawld

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/johnstcn/freshcomics/internal/store"
	"github.com/johnstcn/gocrawl/pkg/crawl"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var errNoPendingWork = errors.New("no pending work")

func New(cfg Config, store store.Store) (*CrawlDaemon, error) {
	opts := crawl.CrawlerOpts{
		Timeout: time.Duration(cfg.FetchTimeoutSecs) * time.Second,
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
		siteDefs:      store,
		siteUpdates:   store,
		crawlInfos:    store,
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
			log.WithField("signal", s).Error("exiting on signal")
			d.stopWorker <- true
			d.stopScheduler <- true
			d.exit(1)
		} else {
			log.WithField("signal", s).Info("ignoring signal")
		}
	}
}

func (d *CrawlDaemon) scheduleWorkForever() {
	for {
		select {
		case <-d.stopScheduler:
			log.Error("stopping scheduler")
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
		return errors.Wrap(err, "fetching active site_defs")
	}

	for _, def := range defs {
		logWithID := log.WithField("site_def_id", def.ID)
		if _, found := pendingIDs[def.ID]; found {
			logWithID.Debug("pending work already exists")
			continue
		}

		crawls, err := d.crawlInfos.GetCrawlInfo(def.ID)
		if err != nil {
			logWithID.Error("fetching previous crawls")
			continue
		}

		if !d.shouldSchedule(def, crawls) {
			continue
		}

		lastURL, err := d.getLastURL(def)
		if err != nil {
			logWithID.Debug("skipping scheduling")
			logWithID.WithError(err).Error("fetch last URL")
		}

		if _, err := d.crawlInfos.CreateCrawlInfo(def.ID, lastURL); err != nil {
			logWithID.Error("scheduling work for site def")
		}
	}

	return nil
}

func (d *CrawlDaemon) getLastURL(def store.SiteDef) (string, error) {
	lastURL, err := d.siteDefs.GetLastURL(def.ID)
	if err == sql.ErrNoRows {
		return def.StartURL, nil
	}

	if err != nil {
		return "", err
	}

	return lastURL, nil
}

func (d *CrawlDaemon) shouldSchedule(def store.SiteDef, crawls []store.CrawlInfo) bool {
	if len(crawls) == 0 {
		return true
	}

	lastCrawl := crawls[0]
	if !lastCrawl.EndedAt.Valid {
		return false
	}

	lastCrawlTime := lastCrawl.EndedAt.Time
	nextScheduleTime := lastCrawlTime.Add(time.Duration(d.config.CheckIntervalSecs) * time.Second)
	return nextScheduleTime.After(d.now())
}

func (d *CrawlDaemon) doWorkForever() {
	for {
		select {
		case <-d.stopWorker:
			log.Error("stopping worker")
			return
		case <-time.After(time.Duration(d.config.WorkPollIntervalSecs) * time.Second):
			item, err := d.getWorkOnce()
			if err == errNoPendingWork {
				log.Debug(err)
				continue
			}
			if err != nil {
				log.WithError(err).Error("fetching pending work")
				continue
			}

			log.WithField("work", item).Debug("got work")
			if err := d.doWorkOnce(item); err != nil {
				log.WithError(err).WithField("work", item).Error("doing work")
			}
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

func (d *CrawlDaemon) makeCrawlJob(def store.SiteDef, url string) crawl.Job {
	return crawl.Job{
		Request: crawl.Request{
			URL:     url,
			Method:  http.MethodGet,
			Headers: map[string]string{"User-Agent": d.config.UserAgent},
			Body:    "",
		},
		Rules: []crawl.Rule{
			{
				Name:  "next_page",
				XPath: def.NextPageXPath,
				Filters: []crawl.Filter{
					{
						Find:    def.RefRegexp,
						Replace: "$1",
					},
				},
			},
			{
				Name:  "title",
				XPath: def.TitleXPath,
				Filters: []crawl.Filter{
					{
						Find:    def.TitleRegexp,
						Replace: "$1",
					},
				},
			},
		},
	}
}

func (d *CrawlDaemon) doWorkOnce(ci *store.CrawlInfo) error {
	// fetch last URL
	// loop
	// 	 parse page
	// 	 check if persisted
	// 	 persist
	// 	 check for next page result
	//   break if no result
	var seen int
	var crawlErr error
	var currentURL = ci.URL

	logWithID := log.WithField("crawl_id", ci.ID)

	if err := d.crawlInfos.StartCrawlInfo(ci.ID); err != nil {
		logWithID.WithError(err).Error("marking crawl started")
	} else {
		logWithID.WithField("current_page", currentURL).Info("starting crawl")
	}

	defer func() {
		if crawlErr != nil {
			logWithID.WithField("current_page", currentURL).WithError(crawlErr).Info("crawl error")
		}

		if err := d.crawlInfos.EndCrawlInfo(ci.ID, crawlErr, seen); err != nil {
			logWithID.WithError(err).Error("marking crawl completed)")
		}
	}()

	def, err := d.siteDefs.GetSiteDef(ci.SiteDefID)
	if err != nil {
		return errors.Wrap(err, "fetching site def")
	}

	refExpr, err := regexp.Compile(def.RefRegexp)
	if err != nil {
		crawlErr = errors.Wrapf(err, "invalid ref regexp %q", def.RefRegexp)
		return nil
	}

	for {
		refResults := refExpr.FindStringSubmatch(currentURL)
		if len(refResults) == 0 {
			crawlErr = fmt.Errorf("no match for ref regexp")
			return nil
		}
		newRef := refResults[1]

		crawlJob := d.makeCrawlJob(def, currentURL)
		result, crawlErr := d.crawler.Crawl(crawlJob)
		if crawlErr != nil {
			return errors.Wrapf(crawlErr, "fetching page %q", crawlJob.Request.URL)
		}

		titleResult, found := result["title"]
		if !found {
			crawlErr = errors.New("no output for title rule")
			return nil
		}

		if titleResult.Error != "" {
			crawlErr = errors.New(titleResult.Error)
			return nil
		}

		if len(titleResult.Values) == 0 {
			crawlErr = errors.New("no matches for title Xpath")
			return nil
		}

		newTitle := titleResult.Values[0]

		newUpdate := store.SiteUpdate{
			SiteDefID: ci.SiteDefID,
			URL:       currentURL,
			Ref:       newRef,
			Title:     newTitle,
			SeenAt:    d.now(),
		}

		if _, found, err := d.siteUpdates.GetSiteUpdate(ci.SiteDefID, newRef); found {
			logWithID.WithField("ref", newRef).Info("already persisted")
		} else if err != nil {
			logWithID.WithError(err).Error("checking if site update already persisted")
		} else if _, err := d.siteUpdates.CreateSiteUpdate(newUpdate); err != nil {
			logWithID.WithError(err).Error("persisting site update")
			return err
		} else {
			logWithID.WithField("update", newUpdate).Info("persisted new update")
			seen += 1
		}

		nextpageResult, found := result["next_page"]
		if !found {
			crawlErr = errors.New("no output for next page rule")
			return nil
		}

		if nextpageResult.Error != "" {
			crawlErr = errors.New(nextpageResult.Error)
			return nil
		}

		if len(nextpageResult.Values) == 0 {
			crawlErr = errors.New("no matches for ref Xpath")
			return nil
		}

		newRef = nextpageResult.Values[0]
		currentURL = fmt.Sprintf(def.URLTemplate, newRef)
	}
}
