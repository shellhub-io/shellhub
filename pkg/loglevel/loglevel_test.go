package loglevel

import (
	envMocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestSetLevel(t *testing.T) {
	mocks := new(envMocks.Backend)
	cases := []struct {
		description   string
		level         string
		expected      logrus.Level
		requiredMocks func()
	}{
		{
			requiredMocks: func() {
				mocks.On("Get", "SHELLHUB_ENV").Return().Once()
			},
			description: {},
		},
		{},
	}

}
