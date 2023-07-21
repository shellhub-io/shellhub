package requests

import (
	"testing"

	"github.com/gustavosbarreto/structsnapshot"
	"github.com/stretchr/testify/assert"
)

func TestNamespaceSnapShot(t *testing.T) {
	namespaceStructList := []interface{}{
		TenantParam{},
		RoleBody{},
		MemberParam{},
		NamespaceCreate{},
		NamespaceGet{},
		NamespaceDelete{},
		NamespaceAddUser{},
		NamespaceRemoveUser{},
		NamespaceEditUser{},
		SessionEditRecordStatus{},
	}

	for _, ns := range namespaceStructList {
		snapShot, err := structsnapshot.TakeSnapshot(ns)
		assert.NoError(t, err)

		loadedSnapshot, err := structsnapshot.LoadSnapshot(ns)
		assert.NoError(t, err)

		assert.Equal(t, loadedSnapshot, snapShot)
	}
}
