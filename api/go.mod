module github.com/shellhub-io/shellhub/api

go 1.14

require (
	github.com/cnf/structhash v0.0.0-20180104161610-62a607eb0224
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.3.0 // indirect
	github.com/mitchellh/mapstructure v1.2.1
	github.com/shellhub-io/shellhub v0.0.0-00010101000000-000000000000
	github.com/valyala/fasttemplate v1.1.0 // indirect
	github.com/xakep666/mongo-migrate v0.2.1
	go.mongodb.org/mongo-driver v1.3.1
	golang.org/x/crypto v0.0.0-20200317142112-1b76d66859c6
	golang.org/x/net v0.0.0-20200319234117-63522dbf7eec
)

replace github.com/shellhub-io/shellhub => ../
