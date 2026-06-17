package services

import (
	"errors"
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

// TestErrDeviceLicenseLimit asserts that ErrDeviceLicenseLimit is a distinct sentinel
// from ErrDeviceLimit: they share the same layer and code but differ by message,
// so errors.Is must return false when comparing one against the other.
func TestErrDeviceLicenseLimit(t *testing.T) {
	if errors.Is(ErrDeviceLicenseLimit, ErrDeviceLimit) {
		t.Error("ErrDeviceLicenseLimit must NOT match ErrDeviceLimit via errors.Is")
	}
}
