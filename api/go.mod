module github.com/shellhub-io/shellhub/api

go 1.14

require (
	github.com/aws/aws-sdk-go v1.37.19 // indirect
	github.com/cnf/structhash v0.0.0-20201127153200-e1b16c1ebc08
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/golang/snappy v0.0.3 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/klauspost/compress v1.11.8 // indirect
	github.com/labstack/echo/v4 v4.2.2
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/magefile/mage v1.11.0 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mitchellh/mapstructure v1.4.1
	github.com/satori/go.uuid v1.2.0
	github.com/shellhub-io/shellhub v0.5.2
	github.com/sirupsen/logrus v1.8.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/undefinedlabs/go-mpatch v1.0.6
	github.com/xakep666/mongo-migrate v0.2.1
	go.mongodb.org/mongo-driver v1.5.1
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/net v0.0.0-20210224082022-3d97a244fca7 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210225134936-a50acf3fe073 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637
	gopkg.in/yaml.v2 v2.3.0 // indirect
)

replace github.com/shellhub-io/shellhub => ../
