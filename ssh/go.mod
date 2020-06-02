module github.com/shellhub-io/shellhub/ssh

go 1.14

require (
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/gliderlabs/ssh v0.3.0
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/parnurzeal/gorequest v0.2.16
	github.com/pires/go-proxyproto v0.1.3
	github.com/pkg/errors v0.9.1 // indirect
	github.com/shellhub-io/shellhub v0.3.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/objx v0.1.1 // indirect
	golang.org/x/crypto v0.0.0-20200602180216-279210d13fed
	golang.org/x/net v0.0.0-20200602114024-627f9648deb9
	golang.org/x/sys v0.0.0-20200602100848-8d3cce7afc34 // indirect
	moul.io/http2curl v1.0.0 // indirect
)

replace github.com/shellhub-io/shellhub => ../
