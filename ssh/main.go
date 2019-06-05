package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/Sirupsen/logrus"
)

func main() {
	opts := &Options{
		Addr:           ":2222",
		Broker:         "tcp://emq:1883",
		ConnectTimeout: 30 * time.Second,
	}

	portRange, _ := ioutil.ReadFile("/proc/sys/net/ipv4/ip_local_port_range")
	fmt.Sscanf(string(portRange), "%d %d", &opts.MinPort, &opts.MaxPort)

	server := NewServer(opts)

	logrus.Fatal(server.ListenAndServe())
}
