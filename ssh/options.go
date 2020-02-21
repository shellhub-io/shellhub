package main

import "time"

type Options struct {
	Addr           string
	Broker         string
	ConnectTimeout time.Duration
}
