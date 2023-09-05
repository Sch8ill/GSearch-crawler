package elasticClient

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/rs/zerolog/log"
)

type ElasticClient struct {
	url    string
	client *elasticsearch.Client
}

func New(url string) *ElasticClient {
	return &ElasticClient{url: url}
}

func (ec *ElasticClient) Connect() error {
	cfg := elasticsearch.Config{
		Addresses: []string{ec.url},
	}
	var err error
	ec.client, err = elasticsearch.NewClient(cfg)
	if err != nil {
		return err
	}

	log.Debug().Msg("Connected to Elasticsearch")
	return nil
}
