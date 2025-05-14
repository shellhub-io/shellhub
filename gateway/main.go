package main

import (
	"log"
)

func main() {
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

		if config.Tunnels {
			domain := config.Domain

			if config.TunnelsDomain != "" {
				domain = config.TunnelsDomain
			}

			certBot.tunnels = &tunnels{
				domain: domain,
				token:  config.TunnelsDNSProviderToken,
			}
		}

		certBot.ensureCertificates()
		certBot.executeRenewCertificates()

		go certBot.renewCertificates()
	}

	if config.Env == "development" {
		go nginxController.watchConfigTemplates()
	}

	nginxController.generateConfigs()
	nginxController.start()
}
