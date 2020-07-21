package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/kelseyhightower/envconfig"
	"github.com/shellhub-io/shellhub/agent/pkg/keygen"
	"github.com/shellhub-io/shellhub/agent/selfupdater"
	"github.com/shellhub-io/shellhub/agent/sshd"
	"github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/revdial"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter"
	"github.com/sirupsen/logrus"
)

var AgentVersion string

type ConfigOptions struct {
	ServerAddress string `envconfig:"server_address"`
	PrivateKey    string `envconfig:"private_key"`
	TenantID      string `envconfig:"tenant_id"`
}

type Information struct {
	SSHID string `json:"sshid"`
}

func main() {
	opts := ConfigOptions{}

	err := envconfig.Process("", &opts)
	if err != nil {
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

	serverAddress, err := url.Parse(opts.ServerAddress)
	if err != nil {
		logrus.Fatal(err)
	}

	cli := client.NewClient(client.WithURL(serverAddress))

	info, err := cli.GetInfo()
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Fatal("Failed to get endpoints")
	}

	agent, err := NewAgent()
	if err != nil {
		logrus.Fatal(err)
	}

	agent.opts = &opts
	agent.Info.Version = AgentVersion

	if err := agent.generatePrivateKey(); err != nil {
		logrus.Fatal(err)
	}

	if err := agent.readPublicKey(); err != nil {
		logrus.Fatal(err)
	}

	serverURL, _ := url.Parse(opts.ServerAddress)

	auth, err := cli.AuthDevice(&models.DeviceAuthRequest{
		Info:     agent.Info,
		Sessions: []string{},
		DeviceAuth: &models.DeviceAuth{
			Identity:  agent.Identity,
			TenantID:  opts.TenantID,
			PublicKey: string(keygen.EncodePublicKeyToPem(agent.pubKey)),
		},
	})

	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Panic("Failed authenticate device")
	}
	if l := len(os.Args); l > 1 && os.Args[1] == "info" {
		fmt.Println(getInfo(auth.Namespace + "." + auth.Name + "@" + strings.Split(info.Endpoints.SSH, ":")[0]))
		return
	}

	server := sshd.NewSSHServer(opts.PrivateKey)

	router := mux.NewRouter()
	router.HandleFunc("/ssh/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		conn := r.Context().Value("http-conn").(net.Conn)
		server.Sessions[vars["id"]] = conn
		server.HandleConn(conn)
	})
	router.HandleFunc("/ssh/close/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		server.CloseSession(vars["id"])
	}).Methods("DELETE")

	sv := http.Server{
		Handler: router,
		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			return context.WithValue(ctx, "http-conn", c)
		},
	}

	server.SetDeviceName(auth.Name)

	go func() {
		for {
			listener, err := NewListener(info.Endpoints.API, serverURL.Scheme, auth.Token)
			if err != nil {
				time.Sleep(time.Second * 10)
				continue
			}

			if err := sv.Serve(listener); err != nil {
				continue
			}
		}
	}()

	// Disable check update in development mode
	if AgentVersion != "latest" {
		go func() {
			for {
				nextVersion, err := CheckUpdate(cli)
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

	ticker := time.NewTicker(10 * time.Second)

	for range ticker.C {
		sessions := make([]string, 0, len(server.Sessions))
		for key := range server.Sessions {
			sessions = append(sessions, key)
		}

		auth, err := cli.AuthDevice(&models.DeviceAuthRequest{
			Info:     agent.Info,
			Sessions: sessions,
			DeviceAuth: &models.DeviceAuth{
				Identity:  agent.Identity,
				TenantID:  opts.TenantID,
				PublicKey: string(keygen.EncodePublicKeyToPem(agent.pubKey)),
			},
		})
		if err == nil && auth.Status == "accepted" {
			server.SetDeviceName(auth.Name)
		} else {
			if auth.Status == "removed" {
				updater.CompleteStopAgent()
			}
		}
	}
}

func NewListener(host, protocol, token string) (*revdial.Listener, error) {
	req, _ := http.NewRequest("GET", "", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	protocol = strings.Replace(protocol, "http", "ws", 1)
	wsConn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("%s://%s/ssh/connection", protocol, host), req.Header)
	if err != nil {
		return nil, err
	}

	listener := revdial.NewListener(wsconnadapter.New(wsConn), func(ctx context.Context, path string) (*websocket.Conn, *http.Response, error) {
		return Revdial(ctx, protocol, host, path)
	})

	return listener, nil
}

func Revdial(ctx context.Context, protocol, address, path string) (*websocket.Conn, *http.Response, error) {
	return websocket.DefaultDialer.DialContext(ctx, strings.Join([]string{fmt.Sprintf("%s://%s", protocol, address), path}, ""), nil)
}

func CheckUpdate(cli client.Client) (*semver.Version, error) {
	info, err := cli.GetInfo()
	if err != nil {
		return nil, err
	}

	return semver.NewVersion(info.Version)
}

func getInfo(input string) string {
	info := Information{
		SSHID: input,
	}
	prettyJSON, err := json.MarshalIndent(info, "", "    ")
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Fatal("Failed to generate json")
	}
	return string(prettyJSON)
}
