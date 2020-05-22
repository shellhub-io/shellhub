package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
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
	"github.com/parnurzeal/gorequest"
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

func buildAPIUrl(protocol string, e *models.Endpoints, uri string) string {
	return fmt.Sprintf("%s://%s/api/%s", protocol, e.API, uri)
}

func sendAuthRequest(endpoints *models.Endpoints, protocol string, identity *models.DeviceIdentity, info *models.DeviceInfo, pubKey *rsa.PublicKey, tenantID string, sessions []string) (*models.DeviceAuthResponse, error) {
	var auth models.DeviceAuthResponse

	_, _, errs := gorequest.New().Post(buildAPIUrl(protocol, endpoints, "/devices/auth")).Send(&models.DeviceAuthRequest{
		Info:     info,
		Sessions: sessions,
		DeviceAuth: &models.DeviceAuth{
			Identity: identity,
			TenantID: tenantID,
			PublicKey: string(pem.EncodeToMemory(&pem.Block{
				Type:  "RSA PUBLIC KEY",
				Bytes: x509.MarshalPKCS1PublicKey(pubKey),
			})),
		},
	}).EndStruct(&auth)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return &auth, nil
}

func Revdial(ctx context.Context, protocol, address, path string) (*websocket.Conn, *http.Response, error) {
	return websocket.DefaultDialer.DialContext(ctx, strings.Join([]string{fmt.Sprintf("%s://%s", protocol, address), path}, ""), nil)
}

type Information struct {
	SSHID string `json:"sshid"`
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

func main() {
	opts := ConfigOptions{}

	err := envconfig.Process("", &opts)
	if err != nil {
		logrus.Panic(err)
	}

	updater, err := NewUpdater()
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

	info := models.Info{}

	_, _, errs := gorequest.New().Get(fmt.Sprintf("%s/info", opts.ServerAddress)).EndStruct(&info)
	if len(errs) > 0 {
		logrus.WithFields(logrus.Fields{"err": errs[0]}).Fatal("Failed to get endpoints")
	}

	identity, err := GetDeviceIdentity()
	if err != nil {
		logrus.Fatal(err)
	}

	devinfo, err := GetDeviceInfo()
	if err != nil {
		logrus.Fatal(err)
	}
	devinfo.Version = AgentVersion

	if _, err := os.Stat(opts.PrivateKey); os.IsNotExist(err) {
		logrus.Info("Private key not found. Generating...")
		err := generatePrivateKey(opts.PrivateKey)
		if err != nil {
			logrus.Fatal(err)
		}
	}

	pubKey, err := readPublicKey(opts.PrivateKey)
	if err != nil {
		logrus.Fatal(err)
	}

	serverURL, _ := url.Parse(opts.ServerAddress)

	auth, err := sendAuthRequest(&info.Endpoints, serverURL.Scheme, identity, devinfo, pubKey, opts.TenantID, []string{})
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Panic("Failed authenticate device")
	}
	if l := len(os.Args); l > 1 && os.Args[1] == "info" {
		fmt.Println(getInfo(auth.Namespace + "." + auth.Name + "@" + strings.Split(info.Endpoints.SSH, ":")[0]))
		return
	}

	server := NewSSHServer(opts.PrivateKey)

	router := mux.NewRouter()
	router.HandleFunc("/ssh/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		conn := r.Context().Value("http-conn").(net.Conn)
		server.sessions[vars["id"]] = conn
		server.sshd.HandleConn(conn)
	})
	router.HandleFunc("/ssh/close/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		server.closeSession(vars["id"])
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
				nextVersion, err := CheckUpdate(opts.ServerAddress)
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
		sessions := make([]string, 0, len(server.sessions))
		for key := range server.sessions {
			sessions = append(sessions, key)
		}

		auth, err = sendAuthRequest(&info.Endpoints, serverURL.Scheme, identity, devinfo, pubKey, opts.TenantID, sessions)
		if err == nil {
			server.SetDeviceName(auth.Name)
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
