module github.com/shellhub-io/shellhub/ssh

go 1.14

require (
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/gliderlabs/ssh v0.3.2
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.8 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/parnurzeal/gorequest v0.2.16
	github.com/pires/go-proxyproto v0.4.2
	github.com/shellhub-io/shellhub v0.4.2
	github.com/sirupsen/logrus v1.7.0
	golang.org/x/crypto v0.0.0-20201117144127-c1f2f97bffc9
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
	golang.org/x/sys v0.0.0-20201119102817-f84b799fce68 // indirect
)

replace github.com/shellhub-io/shellhub => ../
