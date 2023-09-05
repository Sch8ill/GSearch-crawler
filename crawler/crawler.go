package crawler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/sch8ill/gscrawler/clients/httpClient"
	"github.com/sch8ill/gscrawler/control"
	"github.com/sch8ill/gscrawler/crawler/parser/htmlParser"
	"github.com/sch8ill/gscrawler/crawler/parser/parseUtils"
	"github.com/sch8ill/gscrawler/crawler/parser/textParser"
	"github.com/sch8ill/gscrawler/types"
)

type Crawler struct {
	conn       control.ControllerConnection
	httpClient *httpClient.HttpClient
	waitGroup  *sync.WaitGroup
}

const timestampFormat string = "2006-01-02T15:04:05"

func New(conn control.ControllerConnection, httpClient *httpClient.HttpClient, waitGroup *sync.WaitGroup) *Crawler {
	return &Crawler{
		conn:      conn,
		httpClient: httpClient,
		waitGroup: waitGroup,
	}
}

// start the crawler goroutine
func (c *Crawler) Start() {
	go c.Run()
}

// the main function of the crawler goroutine
func (c *Crawler) Run() {
	for {
		job := c.conn.GetJob()
		switch job.Cmd {
		// a command for the crawler to wait
		case control.WaitCommand:
			time.Sleep(control.EmptyQueueWaitDuration)
			continue

		// a command for the crawler to return
		case control.StopCommand:
			c.waitGroup.Done()
			return

		case control.ScrapeCommand:
			if len(job.Site.Url) < 10 {
				fmt.Println("job.Site.Url is shorter than 10! (crawler.go ~ 58)") ///////////////////////////////////////////////////////////////
				fmt.Println(job.Site.Url)
			}

			parsedSite, err := c.scrape(&job.Site)
			if err != nil {
				continue
			}
			c.conn.SubmitResult(parsedSite)
		}
	}
}

// scrape a site by requesting it from the server and parsing it
func (c *Crawler) scrape(site *types.Site) (types.Site, error) {
	parseUrl(site)

	res, err := c.httpClient.Get(site.Url)

	if err != nil {
		log.Warn().Err(err).Msg("")
		return types.Site{}, err
	}
	defer res.Body.Close()

	site.Timestamp = time.Now().Format(timestampFormat)

	if err := ParseSite(res, site); err != nil {
		log.Warn().Err(err).Msg("")
	}

	return *site, nil
}

// parses a sites content by finding the right parser
func ParseSite(res *http.Response, site *types.Site) (err error) {
	bodyStream := res.Body
	defer bodyStream.Close()

	// determine the content type and parse it with the right parser
	// ToDo: implement a proper parser for the http content header
	contentType := res.Header.Get("Content-Type")

	if strings.Contains(contentType, "text/html") {
		parser := htmlParser.New(site, bodyStream)
		parser.Parse()
		site.Type = "text/html"

	} else if strings.Contains(contentType, "text/plain") {
		parser := textParser.New(site, bodyStream)
		parser.Parse()
		site.Type = "text/plain"

	} else if strings.Contains(contentType, "text/markdown") {
		parser := textParser.New(site, bodyStream)
		parser.Parse()
		site.Type = "text/markdown"

	} else if strings.Contains(contentType, "text/csv") {
		parser := textParser.New(site, bodyStream)
		parser.Parse()
		site.Type = "text/csv"

	} else if strings.Contains(contentType, "application/json") {
		parser := textParser.New(site, bodyStream)
		parser.Parse()
		site.Type = "application/json"

	} else if strings.Contains(contentType, "application/xml") {
		parser := textParser.New(site, bodyStream)
		parser.Parse()
		site.Type = "application/xml"

	} else {
		// no parser found for the content type
		site.Type = contentType
		site.Err = errors.New("no parser for content type: " + site.Type)
		return site.Err
	}

	site.Links = parseUtils.RemoveDuplicateItems(site.Links)
	site.Links = removeUnparsableUrls(site.Links)
	for index, link := range site.Links {
		site.Links[index], _ = extendRelativeUrl(link, site.Url)
	}

	return err
}

// parses a url and adds additional data to the site struct
func parseUrl(site *types.Site) error {
	parsedUrl, err := url.Parse(site.Url)

	if err != nil {
		log.Warn().Err(err).Msg("")
		return err
	}

	site.Host = parsedUrl.Host
	site.Scheme = parsedUrl.Scheme
	return nil
}

// removes urls that are not parsable identified by suffixes
func removeUnparsableUrls(rawUrls []string) []string {
	parsableUrls := []string{}

outerLoop:
	for _, url := range rawUrls {
		if url == "" {
			continue
		}

		cleanUrl := strings.ToLower(url)
		if strings.Contains(cleanUrl, "?") {
			cleanUrl = strings.Split(cleanUrl, "?")[0]
		}

		for _, unparseableSuffix := range UnparseableFileSuffixes {
			if strings.HasSuffix(cleanUrl, unparseableSuffix) {
				continue outerLoop
			}
		}

		parsableUrls = append(parsableUrls, url)
	}

	return parsableUrls
}

// extends a relative url to a full url
func extendRelativeUrl(relativeUrl string, parentUrl string) (string, error) {
	parsedParentUrl, err := url.Parse(parentUrl)
	if err != nil {
		return "", err
	}

	parsedUrl, err := url.Parse(relativeUrl)
	if err != nil {
		return "", err
	}

	if parsedUrl.Host == "" {
		parsedUrl.Host = parsedParentUrl.Host
	}
	if parsedUrl.Scheme == "" {
		parsedUrl.Scheme = parsedParentUrl.Scheme
	}
	return parsedUrl.String(), nil
}
