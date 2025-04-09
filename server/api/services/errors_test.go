package services

import (
	"reflect"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// TestNewErrPublicKeyDataInvalid_data_field checks if the models.PublicKey structure contains the "Data" field required
// by NewErrPublicKeyDataInvalid function.
func TestNewErrPublicKeyDataInvalid_data_field(t *testing.T) {
	key := models.PublicKey{}

	value := reflect.TypeOf(key)
	_, ok := value.FieldByName("Data")
	if !ok {
		t.Fatal("public key model does not contains Data field")
	}
}
