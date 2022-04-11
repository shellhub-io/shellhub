module github.com/shellhub-io/shellhub/cli

go 1.14

require (
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/shellhub-io/shellhub v0.5.2
	github.com/shellhub-io/shellhub/api v0.7.4
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.4.0
	github.com/stretchr/testify v1.7.1
	go.mongodb.org/mongo-driver v1.9.0
)

replace github.com/shellhub-io/shellhub => ../

replace github.com/shellhub-io/shellhub/api => ../api
