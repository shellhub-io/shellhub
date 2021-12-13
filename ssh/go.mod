module github.com/shellhub-io/shellhub/ssh

go 1.14

require (
	github.com/elazarl/goproxy v0.0.0-20210110162100-a92cc753f88e // indirect
	github.com/gliderlabs/ssh v0.3.3
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/parnurzeal/gorequest v0.2.16
	github.com/pires/go-proxyproto v0.6.1
	github.com/shellhub-io/shellhub v0.8.1
	github.com/sirupsen/logrus v1.8.1
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/crypto v0.0.0-20211209193657-4570a0811e8b
	golang.org/x/net v0.0.0-20211209124913-491a49abca63
	golang.org/x/sys v0.0.0-20211210111614-af8b64212486 // indirect
)

replace github.com/shellhub-io/shellhub => ../
