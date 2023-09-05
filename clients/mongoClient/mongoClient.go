package mongoClient

import (
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/sch8ill/gscrawler/types"
)


type MongoClient struct {
	uri    string
	client *mongo.Client
}

const (
	DBName        string        = "gsearch"
	sitesCollName string        = "sites"
	DBTimeout     time.Duration = 10 * time.Second
)

func New(uri string) *MongoClient {
	return &MongoClient{
		uri: uri,
	}
}

// connects the underlying client and tests the connection
func (dbc *MongoClient) Connect() error {
	ctx, _ := context.WithTimeout(context.Background(), DBTimeout)

	var err error
	dbc.client, err = mongo.Connect(ctx, options.Client().ApplyURI(dbc.uri))
	if err != nil {
		return err
	}

	// Perform ping to test the connection
	if err := dbc.client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return err
	}
	log.Debug().Msg("Connected to MongoDB")
	return nil
}

// closes the underlying client
func (dbc *MongoClient) Close() error {
	return dbc.client.Disconnect(context.TODO())
}

// returns a database collection based on database and collection name
func (dbc *MongoClient) GetColl(dbName string, collName string) *mongo.Collection {
	return dbc.client.Database(dbName).Collection(collName)
}

// inserts a site into the database
func (dbc *MongoClient) InsertSite(scrapedSite types.Site) error {
	cleanUrl := strings.Split(scrapedSite.Url, "://")[1]

	newDocument := bson.M{
		"_id":          cleanUrl,
		"url":          scrapedSite.Url,
		"host":         scrapedSite.Host,
		"scheme":       scrapedSite.Scheme,
		"timestamp":    scrapedSite.Timestamp,
		"text":         scrapedSite.Text,
		"depth":        scrapedSite.Depth,
		"foundThrough": scrapedSite.FoundThrough,
		"type":         scrapedSite.Type,
	}

	coll := dbc.GetColl(DBName, sitesCollName)
	_, err := coll.InsertOne(context.Background(), newDocument)

	if err != nil {
		log.Warn().Err(err).Msg("")
		return err
	}
	return nil
}
