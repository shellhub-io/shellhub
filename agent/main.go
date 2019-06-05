package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ConfigOptions struct {
	ServerAddress string `envconfig:"server_address"`
	PrivateKey    string `envconfig:"private_key"`
}

type Endpoints struct {
	API  string `json:"api"`
	SSH  string `json:"ssh"`
	MQTT string `json:"mqtt"`
}

func (e *Endpoints) buildAPIUrl(uri string) string {
	return fmt.Sprintf("http://%s/api/%s", e.API, uri)
}

func main() {
	opts := ConfigOptions{}

	err := envconfig.Process("", &opts)
	if err != nil {
		logrus.Panic(err)
	}

	endpoints := Endpoints{}

	_, _, errs := gorequest.New().Get(fmt.Sprintf("%s/endpoints", opts.ServerAddress)).EndStruct(&endpoints)
	if len(errs) > 0 {
		logrus.WithFields(logrus.Fields{"err": errs[0]}).Fatal("Failed to get endpoints")
	}

	identity, err := GetDeviceIdentity()
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

	var auth AuthResponse

	_, _, errs = gorequest.New().Post(endpoints.buildAPIUrl("/devices/auth")).Send(&AuthRequest{
		Identity: identity,
		PublicKey: string(pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(pubKey),
		})),
	}).EndStruct(&auth)
	if len(errs) > 0 {
		logrus.WithFields(logrus.Fields{"errs": errs}).Panic("Failed authenticate device")
	}

	freePort, err := getFreePort()
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "failed to get free port"))
	}

	server := NewSSHServer(freePort)
	client := NewSSHClient(opts.PrivateKey, endpoints.SSH, freePort)

	go func() {
		logrus.Fatal(server.ListenAndServe())
	}()

	b := NewBroker(endpoints.MQTT, auth.UID, auth.Token)

	b.Subscribe(fmt.Sprintf("connect/%s", auth.UID), client.connect)
	b.Connect()

	select {}
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
