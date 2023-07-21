package requests

import (
	"testing"

	"github.com/gustavosbarreto/structsnapshot"
	"github.com/stretchr/testify/assert"
)

func TestSetupSnapShot(t *testing.T) {
	setupStructList := []interface{}{
		Setup{},
	}

	for _, a := range setupStructList {
		snapShot, err := structsnapshot.TakeSnapshot(a)
		assert.NoError(t, err)

		loadedSnapshot, err := structsnapshot.LoadSnapshot(a)
		assert.NoError(t, err)

		assert.Equal(t, loadedSnapshot, snapShot)
	}
}
