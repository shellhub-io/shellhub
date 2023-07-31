package requests

import (
	"testing"

	"github.com/gustavosbarreto/structsnapshot"
	"github.com/stretchr/testify/assert"
)

func TestPublickeySnapShot(t *testing.T) {
	publicKeyStructList := []interface{}{
		FingerprintParam{},
		PublicKeyGet{},
		PublicKeyFilter{},
		PublicKeyDelete{},
		PublicKeyTagAdd{},
		PublicKeyTagRemove{},
		PublicKeyTagsUpdate{},
		PublicKeyAuth{},
	}

	for _, a := range publicKeyStructList {
		snapShot, err := structsnapshot.TakeSnapshot(a)
		assert.NoError(t, err)

		loadedSnapshot, err := structsnapshot.LoadSnapshot(a)
		assert.NoError(t, err)

		assert.Equal(t, loadedSnapshot, snapShot)
	}
}
