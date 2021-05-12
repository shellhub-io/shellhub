package main

import (
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	MongoUri   string `envconfig:"mongo_uri" default:"mongodb://mongo:27017"`
	StoreCache bool   `envconfig:"store_cache" default:"false"`
}

func main() {
	var cfg config
	if err := envconfig.Process("api", &cfg); err != nil {
		panic(err.Error())
	}
	InitializeAPI(cfg)
}
