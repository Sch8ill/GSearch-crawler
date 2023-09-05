package db

import (
	"github.com/sch8ill/gscrawler/clients/elasticClient"
	"github.com/sch8ill/gscrawler/clients/mongoClient"
	"github.com/sch8ill/gscrawler/types"
)

type DBI interface{
	Connect() error
	InsertSite(types.Site) error
	Close() error
}

type DB struct {
	mongo *mongoClient.MongoClient
	elastic *elasticClient.ElasticClient
}

func New(mongoURI string, elasticURL string) *DB {
	return &DB{
		mongo: mongoClient.New(mongoURI),
		elastic: elasticClient.New(elasticURL),
	}
}

func (db *DB) Connect() error {
	if err := db.mongo.Connect(); err != nil {
		return err
	}
	if err := db.elastic.Connect(); err != nil {
		return err
	}
	return nil
}

func (db *DB) InsertSite(site types.Site) error {
	return db.mongo.InsertSite(site)
}

func (db *DB) Close() error {
	return db.mongo.Close()
}