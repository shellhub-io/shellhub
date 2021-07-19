module github.com/shellhub-io/shellhub/api

go 1.14

require (
	github.com/cnf/structhash v0.0.0-20201127153200-e1b16c1ebc08
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-redis/cache/v8 v8.4.1
	github.com/go-redis/redis/v8 v8.11.0
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/labstack/echo/v4 v4.3.0
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mitchellh/mapstructure v1.4.1
	github.com/shellhub-io/shellhub v0.5.2
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v0.0.7
	github.com/square/mongo-lock v0.0.0-20201208161834-4db518ed7fb2
	github.com/stretchr/testify v1.7.0
	github.com/undefinedlabs/go-mpatch v1.0.6
	github.com/xakep666/mongo-migrate v0.2.1
	go.mongodb.org/mongo-driver v1.6.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637
)

replace github.com/shellhub-io/shellhub => ../
