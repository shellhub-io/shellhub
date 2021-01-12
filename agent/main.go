package main

import (
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/shellhub-io/shellhub/agent/selfupdater"
	"github.com/shellhub-io/shellhub/agent/sshd"
	"github.com/sirupsen/logrus"
)

// Agent version to be embed inside the binary. This is injected using `-ldflags`
// build option (e.g: `go build -ldflags "-X main.AgentVersion=1.2.3"`).
//
// If set to `latest`, the auto-updating mechanism is disabled. This is intended
// to be used during development only.
var AgentVersion string

// Provides the configuration for the agent service. The values are load from
// the system environment and control multiple aspects of the service.
type ConfigOptions struct {
	// Set the ShellHub Cloud server address the agent will use to connect.
	ServerAddress string `envconfig:"server_address" required:"true"`

	// Specify the path to the device private key.
	PrivateKey string `envconfig:"private_key" required:"true"`

	// Sets the account tenant id used during communication to associate the
	// device to a specific tenant.
	TenantID string `envconfig:"tenant_id" required:"true"`

	// Determine the interval to send the keep alive message to the server. This
	// has a direct impact of the bandwidth used by the device when in idle
	// state. Default is 30 seconds.
	KeepAliveInterval int `envconfig:"keepalive_interval" default:"30"`

	// Set the device preferred hostname. This provides a hint to the server to
	// use this as hostname if it is available.
	PreferredHostname string `envconfig:"preferred_hostname"`
	SimplePassword    string `envconfig:"simple_password"`
}

type Information struct {
	SSHID string `json:"sshid"`
}

func main() {
	if os.Geteuid() != 0 {
		logrus.Error("ShellHub must be run as root")
		os.Exit(1)
	}

	opts := ConfigOptions{}
	defer envconfig.Usage("shellhub", &opts)

	envconfig.Process("", &opts)
	if err := envconfig.Process("shellhub", &opts); err != nil {
		logrus.Panic(err)
	}

	updater, err := selfupdater.NewUpdater(AgentVersion)
	if err != nil {
		logrus.Panic(err)
	}

	if err := updater.CompleteUpdate(); err != nil {
		logrus.Warning(err)
		os.Exit(0)
	}

	currentVersion := new(semver.Version)

	if AgentVersion != "latest" {
		currentVersion, err = updater.CurrentVersion()
		if err != nil {
			logrus.Panic(err)
		}
	}

	logrus.WithFields(logrus.Fields{
		"version": AgentVersion,
	}).Info("Starting ShellHub")

	agent, err := NewAgent(&opts)
	if err != nil {
		logrus.Fatal(err)
	}

	if err := agent.initialize(); err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Fatal("Failed to initialize agent")
	}

	sshserver := sshd.NewServer(agent.cli, agent.authData, opts.PrivateKey, opts.KeepAliveInterval, opts.SimplePassword)

	tunnel := NewTunnel()
	tunnel.connHandler = func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		conn := r.Context().Value("http-conn").(net.Conn)
		sshserver.Sessions[vars["id"]] = conn
		sshserver.HandleConn(conn)
	}
	tunnel.closeHandler = func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sshserver.CloseSession(vars["id"])
	}

	sshserver.SetDeviceName(agent.authData.Name)

	go func() {
		for {
			listener, err := agent.newReverseListener()
			if err != nil {
				time.Sleep(time.Second * 10)
				continue
			}

			namespace := agent.authData.Namespace
			tenantName := agent.authData.Name
			sshEndpoint := agent.serverInfo.Endpoints.SSH

			sshid := strings.NewReplacer(
				"{namespace}", namespace,
				"{tenantName}", tenantName,
				"{sshEndpoint}", strings.Split(sshEndpoint, ":")[0],
			).Replace("{namespace}.{tenantName}@{sshEndpoint}")

			logrus.WithFields(logrus.Fields{
				"namespace":      namespace,
				"hostname":       tenantName,
				"server_address": opts.ServerAddress,
				"ssh_server":     sshEndpoint,
				"sshid":          sshid,
			}).Info("Server connection established")

			if err := tunnel.Listen(listener); err != nil {
				continue
			}
		}
	}()

	// Disable check update in development mode
	if AgentVersion != "latest" {
		go func() {
			for {
				nextVersion, err := agent.checkUpdate()
				if err != nil {
					logrus.Error(err)
					goto sleep
				}

				if nextVersion.GreaterThan(currentVersion) {
					if err := updater.ApplyUpdate(nextVersion); err != nil {
						logrus.Error(err)
					}
				}

			sleep:
				time.Sleep(time.Hour * 24)
			}
		}()
	}

	ticker := time.NewTicker(time.Duration(opts.KeepAliveInterval) * time.Second)

	for range ticker.C {
		sessions := make([]string, 0, len(sshserver.Sessions))
		for key := range sshserver.Sessions {
			sessions = append(sessions, key)
		}

		agent.sessions = sessions

		if err := agent.authorize(); err != nil {
			sshserver.SetDeviceName(agent.authData.Name)
		}
	}
}
