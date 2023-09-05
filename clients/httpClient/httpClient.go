package httpClient

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/sch8ill/gscrawler/config"
)

var (
	userAgent      string            = fmt.Sprintf("Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/108.0.0.0 gscrawler/%s", config.Version)
	DefualtTimeout time.Duration     = time.Second * 10
	headers        map[string]string = map[string]string{
		"User-Agent":      userAgent,
		"Connection":      "keep-alive",
		"Accept-Encoding": "*",
		"Accept":          "*",
	}
)

type HttpClient struct {
	client  *http.Client
	Headers map[string]string
}

func New(timeout time.Duration) *HttpClient {
	client := &http.Client{
		Timeout: timeout,
	}
	return &HttpClient{client: client, Headers: headers}
}

// Get requests contents from a host using a http GET request
func (c *HttpClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return c.client.Do(req)
}

// SetTransport sets the transport layer of the unerlying http client
func (c *HttpClient) SetTransport(transport *http.Transport) {
	c.client.Transport = transport
}

// NewTransportProxy creates a http.Transport struct using the given proxy url
func NewTransportProxy(rawProxyUrl string) *http.Transport {
	proxyUrl, err := url.Parse(rawProxyUrl)
	if err != nil {
		log.Warn().Err(err)
	}
	return &http.Transport{Proxy: http.ProxyURL(proxyUrl)}

}
