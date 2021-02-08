module github.com/shellhub-io/shellhub/agent

go 1.14

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/GehirnInc/crypt v0.0.0-20200316065508-bb7000b8a962
	github.com/Masterminds/semver v1.5.0
	github.com/Microsoft/go-winio v0.4.15 // indirect
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/creack/pty v1.1.11
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v0.0.0-20190404075923-dbe4a30928d4
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gliderlabs/ssh v0.3.2
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattn/go-shellwords v1.0.11
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/shellhub-io/shellhub v0.4.2
	github.com/sirupsen/logrus v1.7.0
	golang.org/x/crypto v0.0.0-20201117144127-c1f2f97bffc9
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b // indirect
	golang.org/x/sys v0.0.0-20201119102817-f84b799fce68
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	google.golang.org/genproto v0.0.0-20201119123407-9b1e624d6bc4 // indirect
	google.golang.org/grpc v1.33.2 // indirect
	gotest.tools v2.2.0+incompatible // indirect
)

replace github.com/shellhub-io/shellhub => ../
