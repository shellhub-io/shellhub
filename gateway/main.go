package main

import (
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	log "github.com/sirupsen/logrus"
)

const (
	// defaultNginxRootDir is the default base directory for Nginx configuration files.
	defaultNginxRootDir = "/etc/nginx"
	// defaultNginxTemplateDir is the default directory where Nginx template files are stored.
	defaultNginxTemplateDir = "/templates"
	// defaultCertBotRootDir is the default directory where Certbot keeps
	// generated certificates, keys, and related assets.
	defaultCertBotRootDir = "/etc/letsencrypt"
)

func main() {
	loglevel.UseEnvs()

	config, err := loadGatewayConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	log.WithField("config", config).Info("configuration loaded")

	nginxController := &NginxController{
		gatewayConfig: config,
		rootDir:       defaultNginxRootDir,
		templatesDir:  defaultNginxTemplateDir,
	}

	if config.Env != "development" && config.EnableAutoSSL {
		log.Info("auto ssl enabled")

		certBot := newCertBot(&Config{
			Domain:          config.Domain,
			RootDir:         defaultCertBotRootDir,
			RenewedCallback: nginxController.reload,
		})

		if config.Tunnels {
			log.Info("tunnels enabled")

			domain := config.Domain

			if config.TunnelsDomain != "" {
				domain = config.TunnelsDomain
			}

			log.WithFields(log.Fields{
				"domain": domain,
				"token":  halfString(config.TunnelsDNSProviderToken),
			}).Info("tunnels info")

			certBot.Config.Tunnels = &Tunnels{
				Domain:   domain,
				Provider: DigitalOceanDNSProvider,
				Token:    config.TunnelsDNSProviderToken,
			}
		}

		certBot.ensureCertificates()
		log.Info("certificates ensured")

		certBot.executeRenewCertificates()
		log.Info("renew executed")

		go certBot.renewCertificates()
	}

	if config.Env == "development" {
		log.Info("shellhub environment is developer")

		go nginxController.watchConfigTemplates()
	}

	log.Info("generating configurations")

	nginxController.generateConfigs()
	log.Info("configuration generated")

	log.Info("nginx controller running")
	nginxController.start()
}
