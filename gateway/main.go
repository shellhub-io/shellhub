package main

import (
	"context"
	"log"

	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/pkg/api/query"
)

type GatewayState struct {
	Setuped bool
}

func main() {
	ctx := context.Background()

	config, err := loadGatewayConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	nginxController := &NginxController{
		rootDir:       "/etc/nginx",
		templatesDir:  "/templates",
		gatewayConfig: config,
	}

	if config.Env != "development" && config.EnableAutoSSL {
		certBot := &CertBot{
			domain:          config.Domain,
			rootDir:         "/etc/letsencrypt",
			renewedCallback: nginxController.reload,
		}
		certBot.ensureCertificates()
		go certBot.renewCertificates()
	}

	if config.Env == "development" {
		go nginxController.watchConfigTemplates()
	}

	_, db, err := mongo.Connect(ctx, config.MongoURI)
	if err != nil {
		log.Fatalf("unable to connect to MongoDB: %v", err)
	}

	store, err := mongo.NewStore(ctx, db, nil)
	if err != nil {
		log.Fatalf("failed to create the store: %v", err)
	}

	_, count, err := store.UserList(ctx, query.Paginator{}, query.Filters{})
	if err != nil {
		log.Fatalf("failed to count how many user are in the instance: %v", err)
	}

	nginxController.gatewayState = &GatewayState{
		Setuped: count != 0,
	}

	nginxController.generateConfigs()
	nginxController.start()
}
