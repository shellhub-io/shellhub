module github.com/shellhub-io/shellhub/ssh

go 1.14

require (
	github.com/anmitsu/go-shlex v0.0.0-20161002113705-648efa622239 // indirect
	github.com/gliderlabs/ssh v0.3.0
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/parnurzeal/gorequest v0.2.16
	github.com/pires/go-proxyproto v0.0.0-20200315140437-f0371d3cede2
	github.com/pkg/errors v0.9.1 // indirect
	github.com/shellhub-io/shellhub v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.4.2
	golang.org/x/crypto v0.0.0-20200317142112-1b76d66859c6
	golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3
	moul.io/http2curl v1.0.0 // indirect
)

replace github.com/shellhub-io/shellhub => ../
