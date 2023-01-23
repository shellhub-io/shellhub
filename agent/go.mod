module github.com/shellhub-io/shellhub/agent

go 1.14

require (
	github.com/GehirnInc/crypt v0.0.0-20200316065508-bb7000b8a962
	github.com/Masterminds/semver v1.5.0
	github.com/Microsoft/go-winio v0.4.16 // indirect
	github.com/creack/pty v1.1.18
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v20.10.23+incompatible
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gliderlabs/ssh v0.3.5
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattn/go-shellwords v1.0.12
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pkg/sftp v1.13.5
	github.com/shellhub-io/shellhub v0.5.2
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cobra v1.6.1
	github.com/stretchr/testify v1.8.1
	golang.org/x/crypto v0.0.0-20220826181053-bd7e27e6170d
	golang.org/x/sys v0.0.0-20220825204002-c680a09ffe64
)

replace github.com/shellhub-io/shellhub => ../
