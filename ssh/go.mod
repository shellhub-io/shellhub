module github.com/shellhub-io/shellhub/ssh

go 1.14

require (
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/gliderlabs/ssh v0.3.5
	github.com/go-resty/resty/v2 v2.7.0
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/pires/go-proxyproto v0.6.2
	github.com/shellhub-io/shellhub v0.8.1
	github.com/sirupsen/logrus v1.9.0
	github.com/stretchr/testify v1.8.0
	golang.org/x/crypto v0.0.0-20220826181053-bd7e27e6170d
	golang.org/x/net v0.0.0-20220826154423-83b083e8dc8b
	google.golang.org/protobuf v1.27.1 // indirect
)

replace github.com/shellhub-io/shellhub => ../
