package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/miekg/dns"
)

type dnsHandler struct {
	records map[string]string
}

func (d *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		d.parseQuery(m)
	}

	w.WriteMsg(m)
}

func (d *dnsHandler) parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			if ip, ok := d.records[q.Name]; ok {
				rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
			}
		}
	}
}

func waitForConnection(proto, addr string) bool {
	for i := 0; i < 10; i++ {
		conn, err := net.Dial(proto, addr)
		if err == nil {
			conn.Close()
			return true
		}

		time.Sleep(time.Second)
	}

	return false
}

func TestMain(m *testing.M) {
	// Create a virtual network adapter
	if _, err := exec.Command("ifconfig", "eth0:0", "127.0.0.11").Output(); err != nil {
		panic(err)
	}

	server := &dns.Server{
		Addr: "127.0.0.11:53",
		Net:  "udp",
		Handler: &dnsHandler{
			records: map[string]string{
				"api.": "127.0.0.1",
				"ssh.": "127.0.0.1",
				"ui.":  "127.0.0.1",
			},
		},
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	// Wait for DNS test server to be started
	if !waitForConnection("udp", "127.0.0.11:53") {
		panic("Failed to connect to DNS test server")
	}

	// Start OpenResty daemon
	cmd := exec.Command("/entrypoint.sh", "/usr/local/openresty/bin/openresty", "-g", "daemon off;")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	// Wait for OpenResty to be started
	if !waitForConnection("tcp", "127.0.0.1:80") {
		panic("Failed to connect to OpenResty")
	}

	// Run unit test
	code := m.Run()

	server.Shutdown()

	if err := cmd.Process.Kill(); err != nil {
		panic(err)
	}

	os.Exit(code)
}
