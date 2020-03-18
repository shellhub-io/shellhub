package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/kelseyhightower/envconfig"
	"github.com/parnurzeal/gorequest"
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

type Info struct {
	Version   string    `json:"version"`
	Endpoints Endpoints `json:"endpoints"`
}

type Endpoints struct {
	API string `json:"api"`
	SSH string `json:"ssh"`
}

func (e *Endpoints) buildAPIUrl(uri string) string {
	return fmt.Sprintf("http://%s/api/%s", e.API, uri)
}

func sendAuthRequest(endpoints *Endpoints, identity *DeviceIdentity, info *DeviceInfo, pubKey *rsa.PublicKey, tenantID string, sessions []string) (*AuthResponse, error) {
	var auth AuthResponse

	_, _, errs := gorequest.New().Post(endpoints.buildAPIUrl("/devices/auth")).Send(&AuthRequest{
		Identity: identity,
		Info:     info,
		PublicKey: string(pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(pubKey),
		})),
		TenantID: tenantID,
		Version:  AgentVersion,
		Sessions: sessions,
	}).EndStruct(&auth)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return &auth, nil
}

func Revdial(ctx context.Context, address, path string) (*websocket.Conn, *http.Response, error) {
	return websocket.DefaultDialer.DialContext(ctx, strings.Join([]string{fmt.Sprintf("ws://%s", address), path}, ""), nil)
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

	info := Info{}

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

	auth, err := sendAuthRequest(&info.Endpoints, identity, devinfo, pubKey, opts.TenantID, []string{})
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
			listener, err := NewListener(info.Endpoints.API, auth.Token)
			if err != nil {
				time.Sleep(time.Second * 10)
				continue
			}

			if err := sv.Serve(listener); err != nil {
				continue
			}
		}
	}()

	ticker := time.NewTicker(10 * time.Second)

	for _ = range ticker.C {
		sessions := make([]string, 0, len(server.sessions))
		for key := range server.sessions {
			sessions = append(sessions, key)
		}

		auth, err = sendAuthRequest(&info.Endpoints, identity, devinfo, pubKey, opts.TenantID, sessions)
		if err == nil {
			server.SetDeviceName(auth.Name)
		}
	}
}

func NewListener(host string, token string) (*revdial.Listener, error) {
	req, _ := http.NewRequest("GET", "", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	wsConn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s/ssh/connection", host), req.Header)
	if err != nil {
		return nil, err
	}

	listener := revdial.NewListener(wsconnadapter.New(wsConn), func(ctx context.Context, path string) (*websocket.Conn, *http.Response, error) {
		return Revdial(ctx, host, path)
	})

	return listener, nil
}
