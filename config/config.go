package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

const Version string = "0.1.11"

var (
	MongoDBURI string = getEnvKey("MONGODBURI")
	ElasticURL string = getEnvKey("ELASTICSEARCHURL")
)

func getEnvKey(key string) string {
	godotenv.Load(".env")
	envVar := os.Getenv(key)
	if envVar == "" {
		log.Logger.Fatal().Msgf("Could not load environment variable: %s", key)
	}
	return envVar
}
