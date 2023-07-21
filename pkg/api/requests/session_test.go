package requests

import (
	"testing"

	"github.com/gustavosbarreto/structsnapshot"
	"github.com/stretchr/testify/assert"
)

func TestSessionSnapShot(t *testing.T) {
	sessionStructList := []interface{}{
		SessionKeepAlive{},
		SessionFinish{},
		SessionCreate{},
		SessionAuthenticatedSet{},
		SessionGet{},
		SessionIDParam{},
	}

	for _, a := range sessionStructList {
		snapShot, err := structsnapshot.TakeSnapshot(a)
		assert.NoError(t, err)

		loadedSnapshot, err := structsnapshot.LoadSnapshot(a)
		assert.NoError(t, err)

		assert.Equal(t, loadedSnapshot, snapShot)
	}
}
