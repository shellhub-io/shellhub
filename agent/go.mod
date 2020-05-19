module github.com/shellhub-io/shellhub/agent

go 1.14

require (
	github.com/Masterminds/semver v1.5.0
	github.com/anmitsu/go-shlex v0.0.0-20161002113705-648efa622239 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v0.0.0-20190404075923-dbe4a30928d4
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gliderlabs/ssh v0.2.3-0.20200214030106-f5cb472d2a7a
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/websocket v1.4.2
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kr/pty v1.1.8
	github.com/mattn/go-shellwords v1.0.10
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/parnurzeal/gorequest v0.2.16
	github.com/pkg/errors v0.9.1
	github.com/shellhub-io/shellhub v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.4.2
	golang.org/x/crypto v0.0.0-20200317142112-1b76d66859c6 // indirect
	google.golang.org/grpc v1.29.1 // indirect
	moul.io/http2curl v1.0.0 // indirect
)

replace github.com/shellhub-io/shellhub => ../
