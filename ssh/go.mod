module github.com/shellhub-io/shellhub/ssh

go 1.14

require (
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/gliderlabs/ssh v0.3.2
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/parnurzeal/gorequest v0.2.16
	github.com/pires/go-proxyproto v0.5.0
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/shellhub-io/shellhub v0.5.2
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/net v0.0.0-20210224082022-3d97a244fca7
	golang.org/x/sys v0.0.0-20210225134936-a50acf3fe073 // indirect
)

replace github.com/shellhub-io/shellhub => ../
