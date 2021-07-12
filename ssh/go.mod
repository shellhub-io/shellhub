module github.com/shellhub-io/shellhub/ssh

go 1.14

require (
	github.com/gliderlabs/ssh v0.3.3
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/parnurzeal/gorequest v0.2.16
	github.com/pires/go-proxyproto v0.6.0
	github.com/shellhub-io/shellhub v0.7.1
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110
)

replace github.com/shellhub-io/shellhub => ../
