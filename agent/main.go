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

var AgentVersion string

type ConfigOptions struct {
	ServerAddress     string `envconfig:"server_address"`
	PrivateKey        string `envconfig:"private_key"`
	TenantID          string `envconfig:"tenant_id"`
	KeepAliveInterval int    `envconfig:"keepalive_interval" default:"30"`
	PreferredHostname string `envconfig:"preferred_hostname"`
}

type Information struct {
	SSHID string `json:"sshid"`
}

func main() {
	opts := ConfigOptions{}

	// Process unprefixed env vars for backward compatibility
	if err := envconfig.Process("", &opts); err != nil {
		logrus.Panic(err)
	}

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

	sshserver := sshd.NewServer(agent.cli, agent.authData, opts.PrivateKey, opts.KeepAliveInterval)

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

			logrus.WithFields(logrus.Fields{
				"namespace":      agent.authData.Namespace,
				"hostname":       agent.authData.Name,
				"server_address": opts.ServerAddress,
				"ssh_server":     agent.serverInfo.Endpoints.SSH,
				"sshid":          agent.authData.Namespace + "." + agent.authData.Name + "@" + strings.Split(agent.serverInfo.Endpoints.SSH, ":")[0],
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
