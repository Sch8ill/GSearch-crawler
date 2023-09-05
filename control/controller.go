package control

import (
	"math/rand"
	"net/url"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/sch8ill/gscrawler/crawler/parser/parseUtils"
	"github.com/sch8ill/gscrawler/db"
	"github.com/sch8ill/gscrawler/types"
)

type Command int

type Controller struct {
	config              ControllerConfig
	channels            ChannelBundle
	crawlerWaitGroup    *sync.WaitGroup
	controllerWaitGroup *sync.WaitGroup
	terminationChannel  chan bool
	scrapeQueue         []types.Site
	scrapedSites        []string
	scrapedSitesCount   int
	db                  db.DBI
	stop                bool
}

type ControllerConfig struct {
	MaxDepth            int
	RequireWhitelisting bool
	WhitelistedHosts    []string
}

// defines a job that is send to the crawlers
type Job struct {
	Cmd  Command
	Site types.Site
}

// bundles multiple channels for communication between controller and crawler
type ChannelBundle struct {
	JobRequestChannel chan bool
	JobChannel        chan Job
	ResultChannel     chan types.Site
}

const (
	ScrapeCommand Command = iota
	StopCommand
	WaitCommand
)

const (
	EmptyQueueWaitDuration time.Duration = time.Second * 1

	// the threshold for randomizing the crawling queue index
	randomIndexThreshold int = 300

	statusLogFrequency int = 30
)

func NewControllerConfig(whitelistedHosts []string, requireWhitelisting bool) ControllerConfig {
	return ControllerConfig{
		WhitelistedHosts:    whitelistedHosts,
		RequireWhitelisting: requireWhitelisting,
	}
}

func NewController(config ControllerConfig, db db.DBI, crawlerWaitGroup *sync.WaitGroup, controllerWaitGroup *sync.WaitGroup) *Controller {
	channelBundle := ChannelBundle{
		JobRequestChannel: make(chan bool),
		JobChannel:        make(chan Job),
		ResultChannel:     make(chan types.Site),
	}
	return &Controller{
		config:              config,
		channels:            channelBundle,
		crawlerWaitGroup:    crawlerWaitGroup,
		controllerWaitGroup: controllerWaitGroup,
		terminationChannel:  make(chan bool),
		scrapeQueue:         []types.Site{},
		scrapedSites:        []string{},
		scrapedSitesCount:   0,
		db:            db,
		stop:                false,
	}
}

// starts the the controller
func (c *Controller) Run() error {
	log.Debug().Msg("Starting Controller...")
	if err := c.db.Connect(); err != nil {
		return err
	}
	
	go c.terminationNotifier()
	go c.serve()
	return nil
}

// tells the controller to stop the crawling process
func (c *Controller) Stop() {
	log.Debug().Msg("Stopping crawlers and controller...")
	c.stop = true
}

// the main loop of the controller
func (c *Controller) serve() {
	for {
		select {
		case <-c.terminationChannel:
			c.controllerWaitGroup.Done()
			return

		case <-c.channels.JobRequestChannel:
			if c.stop {
				c.channels.JobChannel <- Job{Cmd: StopCommand}

			} else {
				c.channels.JobChannel <- c.getScrapingJob()
			}

		case site := <-c.channels.ResultChannel:
			c.submitResult(site)
		}
	}
}

// returns a job from the crawling queue
func (c *Controller) getScrapingJob() Job {
	var index int
	scrapeQueueLength := len(c.scrapeQueue)

	for {
		if len(c.scrapeQueue) > 0 {
			// if possible, randomize the index to prevent accidentally dossing hosts
			if scrapeQueueLength > randomIndexThreshold {
				index = rand.Intn(randomIndexThreshold)
			} else {
				index = 0
			}

			job := Job{
				Cmd:  ScrapeCommand,
				Site: c.scrapeQueue[index],
			}

			c.scrapeQueue = append(c.scrapeQueue[:index], c.scrapeQueue[index+1:]...)

			if checkIfItemInList(job.Site.Url, c.scrapedSites) { // fixes multiple scrapes of on site //////////////////////////////////////////////////////////////////////
				//log.Warn().Msg("alredy scraped site in queue")
				continue
			}
			return job

		} else {
			log.Warn().Msg("the crawling queue is empty")
			return Job{Cmd: WaitCommand}

		}
	}
}

// adds the site to the database and logs the event
func (c *Controller) submitResult(site types.Site) {
	c.scrapedSites = append(c.scrapedSites, site.Url)

	if site.Err == nil {
		log.Info().Int("depth", site.Depth).Int("links", len(site.Links)).Int("text", len(site.Text)).Str("type", site.Type).Str("url", site.Url).Msg("")
		c.db.InsertSite(site)
		c.scrapedSitesCount++

	} else {
		log.Warn().Str("url", site.Url).Err(site.Err).Msg("")
	}

	c.addLinksToQueue(site)
	c.logStatus()
}

// addLinksToQueue adds links to the scrape queue that have been found through scraping
func (c *Controller) addLinksToQueue(site types.Site) {
	depth := site.Depth + 1
	if c.config.MaxDepth > -1 { // check if the max depth is set
		// check if the depth is within the allowed depth
		if depth > c.config.MaxDepth {
			return
		}
	}

	for _, link := range site.Links {
		if checkIfItemInList(link, c.scrapedSites) {
			continue
		}
		parsedUrl, _ := url.Parse(link)

		if c.config.RequireWhitelisting {
			if !parseUtils.Contains(c.config.WhitelistedHosts, parsedUrl.Host) {
				continue
			}
		}

		c.AddToQueue(types.Site{
			Url:          link,
			Depth:        depth,
			FoundThrough: site.Url,
		})
	}
}

// logs the current status of the crawling operation every statusLogFrequency sites scraped
func (c *Controller) logStatus() {
	// modulus
	if c.scrapedSitesCount%statusLogFrequency == 0 {
		log.Info().Int("scrapedSites", c.scrapedSitesCount).Int("queueSize", len(c.scrapeQueue)).Int("dbBuffer", 0).Msg("")
	}
}

// goroutine that notifies the controller when the wait group is done
func (c *Controller) terminationNotifier() {
	c.crawlerWaitGroup.Wait()
	c.terminationChannel <- true
}

// adds the given site to the crawling queue
func (c *Controller) AddToQueue(site types.Site) {
	c.scrapeQueue = append(c.scrapeQueue, site)
}

// adds the given urls to the crawling queue
func (c *Controller) SeedCrawlingQueue(seedUrls []string) {
	for _, url := range seedUrls {
		site := types.Site{
			Url:          url,
			Depth:        0,
			FoundThrough: "",
		}
		c.AddToQueue(site)
	}
}

// returns the channel bundle of the controller
func (c *Controller) GetChannelBundle() ChannelBundle {
	return c.channels
}

// checks if the given item is in the given slice
func checkIfItemInList[T string | int](item T, list []T) bool {
	for _, listItem := range list {
		if listItem == item {
			return true
		}
	}
	return false
}
