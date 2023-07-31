package requests

import (
	"testing"

	"github.com/gustavosbarreto/structsnapshot"
	"github.com/stretchr/testify/assert"
)

func TestTagsSnapShot(t *testing.T) {
	tagStructList := []interface{}{
		TagRename{},
		TagDelete{},
		TagBody{},
		TagParam{},
	}

	for _, a := range tagStructList {
		snapShot, err := structsnapshot.TakeSnapshot(a)
		assert.NoError(t, err)

		loadedSnapshot, err := structsnapshot.LoadSnapshot(a)
		assert.NoError(t, err)

		assert.Equal(t, loadedSnapshot, snapShot)
	}
}
