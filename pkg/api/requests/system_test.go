package requests

import (
	"testing"

	"github.com/gustavosbarreto/structsnapshot"
	"github.com/stretchr/testify/assert"
)

func TestSystemSnapShot(t *testing.T) {
	systemStructList := []interface{}{
		SystemGetInfo{},
		SystemInstallScript{},
	}

	for _, a := range systemStructList {
		snapShot, err := structsnapshot.TakeSnapshot(a)
		assert.NoError(t, err)

		loadedSnapshot, err := structsnapshot.LoadSnapshot(a)
		assert.NoError(t, err)

		assert.Equal(t, loadedSnapshot, snapShot)
	}
}
