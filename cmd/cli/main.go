package main

import (
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/sch8ill/gscrawler/clients/httpClient"
	"github.com/sch8ill/gscrawler/config"
	"github.com/sch8ill/gscrawler/control"
	"github.com/sch8ill/gscrawler/crawler"
	"github.com/sch8ill/gscrawler/db"
)

const (
	stringSliceSeperator string = ","
	timestampFormat      string = "2006-01-02T15:04:05.000"

	logLevel zerolog.Level = zerolog.DebugLevel
)

func main() {
	createLogger()

	app := createApp()
	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("")
	}
}

func createApp() *cli.App {
	return &cli.App{
		Name:      "GSearch-crawler",
		Usage:     "recursively crawl websites and write there contents to a database",
		Copyright: "Copyright (c) 2023 Sch8ill",
		Version:   config.Version,
		Action:    run,
		Flags:     declareFlags(),
	}
}

// declareFlags declares the cli flags
func declareFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "urls",
			Aliases:  []string{"url"},
			Usage:    "the url(s) you want the crawler to recursively crawl",
			Required: true,
		},
		&cli.IntFlag{
			Name:    "crawlers",
			Aliases: []string{"c"},
			Value:   1,
			Usage:   "the number of crawlers that should run simultaneously",
		},
		&cli.DurationFlag{
			Name:  "timeout",
			Usage: "the timeout in seconds that the crawlers should use for http requests",
		},
		&cli.StringFlag{
			Name:  "whitelisted-hosts",
			Usage: "a list of whitelisted hosts that are allowd to be crawled",
		},
		&cli.IntFlag{
			Name:  "max-depth",
			Usage: "the max allowed depth the crawler is allowed to crawl",
		},
		&cli.StringFlag{
			Name:  "proxy",
			Usage: "the url to a proxy that should be used for crawling",
		},
		&cli.BoolFlag{
			Name:  "mock-db",
			Usage: "if the the crawling results should be send to a mock database",
		},
	}
}

// run initiates and starts all goroutines and clients
func run(ctx *cli.Context) error {
	startUrls := parseSliceParamater("urls", ctx)
	crawlerCount := ctx.Int("crawlers")
	controllerConfig := createControllerConfig(ctx)

	// the wait group for the controller
	controllerWaitGroup := &sync.WaitGroup{}
	controllerWaitGroup.Add(1)

	// the wait group for the crawlers
	crawlerWaitGroup := &sync.WaitGroup{}
	crawlerWaitGroup.Add(crawlerCount)

	log.Info().Msg("GSearch-crawler starting...")
	log.Info().Str("Go version", runtime.Version()).Msg("")
	log.Info().Str("GSearch version", "GSearch-crawler@" + config.Version).Msg("")
	log.Info().Str("MongoDB URI", removePasswordFromURL(config.MongoDBURI)).Msg("")
	log.Info().Str("Elastic URL", removePasswordFromURL(config.ElasticURL)).Msg("")
	log.Info().Str("Start URL(s)", strings.Join(startUrls, ", ")).Msg("")

	if ctx.IsSet("proxy") {
		log.Info().Str("HTTP proxy", ctx.String("proxy")).Msg("")
	}

	db := createDBConnection(ctx)

	controller := control.NewController(controllerConfig, db, crawlerWaitGroup, controllerWaitGroup)
	channelBundle := controller.GetChannelBundle()
	controller.SeedCrawlingQueue(startUrls)
	if err := controller.Run(); err != nil {
		log.Fatal().Err(err).Msg("")
	}

	log.Debug().Msgf("Starting %d crawler(s)...", crawlerCount)
	startCrawlers(crawlerCount, channelBundle, crawlerWaitGroup, ctx)

	endChannel := make(chan os.Signal, 1)
	signal.Notify(endChannel, os.Interrupt)
	// block the execution until a signal is send
	<-endChannel

	// terminate all goroutines
	controller.Stop()

	// wait until the controller has stopped
	controllerWaitGroup.Wait()

	db.Close()
	return nil
}

// createControllerConfig creates a controllerConfig struct from the provided cli context
func createControllerConfig(ctx *cli.Context) control.ControllerConfig {
	controllerConfig := control.ControllerConfig{}

	if ctx.IsSet("whitelisted-hosts") {
		controllerConfig.WhitelistedHosts = parseSliceParamater("whitelisted-hosts", ctx)
		controllerConfig.RequireWhitelisting = true
	}

	if ctx.IsSet("max-depth") {
		controllerConfig.MaxDepth = ctx.Int("max-depth")
	} else {
		controllerConfig.MaxDepth = -1
	}
	return controllerConfig
}

// createDBConnection creates a db client or a mock db client based on the cli flags
func createDBConnection(ctx *cli.Context) db.DBI {
	if ctx.Bool("mock-db") {
		return db.NewMockDB()
	} else {
		return db.New(config.MongoDBURI, config.ElasticURL)
	}
}

// startCrawlers starts the crawlers
func startCrawlers(crawlerCount int, channelBundle control.ChannelBundle, crawlerWaitGroup *sync.WaitGroup, ctx *cli.Context) {
	for i := 0; i < crawlerCount; i++ {
		conn := control.NewConnection(channelBundle)
		httpClient := createHttpClient(ctx)
		crawler := crawler.New(*conn, httpClient, crawlerWaitGroup)
		crawler.Start()
	}
}

// createHttpClient creates an http client
func createHttpClient(ctx *cli.Context) *httpClient.HttpClient {
	var client *httpClient.HttpClient
	if ctx.IsSet("timeout") {
		client = httpClient.New(ctx.Duration("timeout"))
	} else {
		client = httpClient.New(httpClient.DefualtTimeout)
	}

	if ctx.IsSet("proxy") {
		transport := httpClient.NewTransportProxy(ctx.String("proxy"))
		client.SetTransport(transport)
	}
	return client
}

// parseSliceParamater parses a cli string parameter to a string slice by seperating it with the stringSliceSeperator
func parseSliceParamater(name string, ctx *cli.Context) []string {
	rawSlice := ctx.String(name)
	var slice []string
	for _, item := range strings.Split(rawSlice, stringSliceSeperator) {
		if item != "" {
			slice = append(slice, item)
		}
	}
	return slice
}

// createLogger creates and configurates the zerolog logger
func createLogger() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: timestampFormat,
	}
	log.Logger = log.Output(consoleWriter).Level(logLevel).With().Timestamp().Logger()
}

// removePasswordFromURL removes the password from a url
func removePasswordFromURL(rawUrl string) string {
	if strings.Contains(rawUrl, "@") {
		parsedURL, _ := url.Parse(rawUrl)
		parsedURL.User = url.User(parsedURL.User.Username())
		return parsedURL.String()
	} else {
		return rawUrl
	}
}
