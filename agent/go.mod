module github.com/shellhub-io/shellhub/agent

go 1.14

require (
	github.com/anmitsu/go-shlex v0.0.0-20161002113705-648efa622239 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/gliderlabs/ssh v0.2.3-0.20200214030106-f5cb472d2a7a
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/websocket v1.4.2
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kr/pty v1.1.8
	github.com/mattn/go-shellwords v1.0.10
	github.com/parnurzeal/gorequest v0.2.16
	github.com/pkg/errors v0.9.1
	github.com/shellhub-io/shellhub v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.4.2
	golang.org/x/crypto v0.0.0-20200317142112-1b76d66859c6 // indirect
	moul.io/http2curl v1.0.0 // indirect
)

replace github.com/shellhub-io/shellhub => ../
