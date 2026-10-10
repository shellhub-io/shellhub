package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"
)

type seen struct {
	Name string `json:"name"`
	Host string `json:"host"`
	URI  string `json:"uri"`
	TLS  bool   `json:"tls"`
}

func selfSigned(domain string) (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: domain},
		DNSNames:     []string{domain},
		NotBefore:    time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
		NotAfter:     time.Date(2100, time.January, 1, 0, 0, 0, 0, time.UTC),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, nil
}

func main() {
	name := os.Getenv("NAME")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(seen{Name: name, Host: r.Host, URI: r.RequestURI, TLS: r.TLS != nil}); err != nil {
			log.Printf("answering a request: %v", err)
		}
	})

	certificate, err := selfSigned(os.Getenv("TLS_DOMAIN"))
	if err != nil {
		log.Fatal(err)
	}

	var listen net.ListenConfig

	plainListener, err := listen.Listen(context.Background(), "tcp", os.Getenv("HTTP_ADDRESS"))
	if err != nil {
		log.Fatal(err)
	}

	tcpListener, err := listen.Listen(context.Background(), "tcp", os.Getenv("TLS_ADDRESS"))
	if err != nil {
		log.Fatal(err)
	}

	secureListener := tls.NewListener(tcpListener, &tls.Config{
		Certificates: []tls.Certificate{certificate},
		MinVersion:   tls.VersionTLS13,
	})

	plain := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	secure := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second}

	go func() { log.Fatal(secure.Serve(secureListener)) }()

	log.Print("listening")
	log.Fatal(plain.Serve(plainListener))
}
