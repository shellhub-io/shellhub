package services

import (
	"errors"
	"reflect"
	"testing"

	pkgerrors "github.com/shellhub-io/shellhub/pkg/errors"
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

// TestErrUserUnhandledDuplicate verifies that ErrUserUnhandledDuplicate is a sentinel
// with ErrCodeDuplicated so the echo handler maps it to HTTP 409, and that
// NewErrUserUnhandledDuplicate returns an error that wraps that sentinel.
func TestErrUserUnhandledDuplicate(t *testing.T) {
	t.Run("sentinel has ErrCodeDuplicated code", func(t *testing.T) {
		var e pkgerrors.Error
		if !pkgerrors.As(ErrUserUnhandledDuplicate, &e) {
			t.Fatal("ErrUserUnhandledDuplicate is not a pkgerrors.Error")
		}

		if e.Code != ErrCodeDuplicated {
			t.Errorf("expected code %d (ErrCodeDuplicated), got %d", ErrCodeDuplicated, e.Code)
		}

		if e.Layer != ErrLayer {
			t.Errorf("expected layer %q, got %q", ErrLayer, e.Layer)
		}
	})

	t.Run("NewErrUserUnhandledDuplicate wraps ErrUserUnhandledDuplicate", func(t *testing.T) {
		err := NewErrUserUnhandledDuplicate()
		if err == nil {
			t.Fatal("NewErrUserUnhandledDuplicate returned nil")
		}

		if !errors.Is(err, ErrUserUnhandledDuplicate) {
			t.Error("NewErrUserUnhandledDuplicate result does not wrap ErrUserUnhandledDuplicate")
		}
	})
}
