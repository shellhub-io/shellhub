package main

import "time"

type Options struct {
	Addr           string
	Broker         string
	MinPort        uint32
	MaxPort        uint32
	ConnectTimeout time.Duration
}
