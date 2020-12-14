module github.com/shellhub-io/shellhub/ssh

go 1.14

require (
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/elazarl/goproxy v0.0.0-20201021153353-00ad82a08272 // indirect
	github.com/gliderlabs/ssh v0.3.1
	github.com/gorilla/mux v1.8.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/parnurzeal/gorequest v0.2.16
	github.com/pires/go-proxyproto v0.3.3
	github.com/pkg/errors v0.9.1 // indirect
	github.com/shellhub-io/shellhub v0.4.2 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/smartystreets/goconvey v1.6.4 // indirect
	golang.org/x/crypto v0.0.0-20201117144127-c1f2f97bffc9
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
	golang.org/x/sys v0.0.0-20201119102817-f84b799fce68 // indirect
	moul.io/http2curl v1.0.0 // indirect
)

replace github.com/shellhub-io/shellhub => ../
