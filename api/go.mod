module github.com/shellhub-io/shellhub/api

go 1.14

require (
	github.com/aws/aws-sdk-go v1.35.33 // indirect
	github.com/cnf/structhash v0.0.0-20201013183111-a92e111048cd
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/golang/snappy v0.0.2 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/klauspost/compress v1.11.3 // indirect
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.3.0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mitchellh/mapstructure v1.4.1
	github.com/satori/go.uuid v1.2.0
	github.com/shellhub-io/shellhub v0.4.2
	github.com/stretchr/testify v1.7.0
	github.com/undefinedlabs/go-mpatch v1.0.6
	github.com/valyala/fasttemplate v1.2.1 // indirect
	github.com/xakep666/mongo-migrate v0.2.1
	go.mongodb.org/mongo-driver v1.4.4
	golang.org/x/crypto v0.0.0-20201117144127-c1f2f97bffc9
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9 // indirect
	golang.org/x/sys v0.0.0-20201119102817-f84b799fce68 // indirect
	golang.org/x/text v0.3.4 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637
)

replace github.com/shellhub-io/shellhub => ../
