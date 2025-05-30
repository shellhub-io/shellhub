package main

import (
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	log "github.com/sirupsen/logrus"
)

const (
	// DefaultNginxRootDir is the default base directory for Nginx configuration files.
	DefaultNginxRootDir = "/etc/nginx"
	// DefaultNginxTemplateDir is the default directory where Nginx template files are stored.
	DefaultNginxTemplateDir = "/templates"
	// DefaultCertBotRootDir is the default directory where Certbot keeps
	// generated certificates, keys, and related assets.
	DefaultCertBotRootDir = "/etc/letsencrypt"
)

func main() {
	loglevel.UseEnvs()

	config, err := LoadGatewayConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	log.WithField("config", config).Info("configuration loaded")

	nginxController := &NginxController{
		RootDir:       DefaultNginxRootDir,
		TemplatesDir:  DefaultNginxTemplateDir,
		GatewayConfig: config,
	}

	if config.Env != "development" && config.EnableAutoSSL {
		log.Info("auto ssl enabled")

		certBot := NewCertBot(&Config{
			Domain:          config.Domain,
			RootDir:         DefaultCertBotRootDir,
			RenewedCallback: nginxController.Reload,
		})

		if config.Tunnels {
			log.Info("tunnels enabled")

			domain := config.Domain

			if config.TunnelsDomain != "" {
				domain = config.TunnelsDomain
			}

			log.WithFields(log.Fields{
				"domain":   domain,
				"provider": config.TunnelsDNSProvider,
				"token":    half(config.TunnelsDNSProviderToken),
			}).Info("tunnels info")

			certBot.Config.Tunnels = &Tunnels{
				Domain:   domain,
				Provider: DigitalOceanDNSProvider,
				Token:    config.TunnelsDNSProviderToken,
			}
		}

		certBot.EnsureCertificates()
		log.Info("certificates ensured")

		certBot.ExecuteRenewCertificates()
		log.Info("renew executed")

		go certBot.RenewCertificates()
	}

	if config.Env == "development" {
		log.Info("shellhub environment is developer")

		go nginxController.WatchConfigTemplates()
	}

	log.Info("generating configurations")

	nginxController.GenerateConfigs()
	log.Info("configuration generated")

	log.Info("nginx controller running")
	nginxController.Start()
}
