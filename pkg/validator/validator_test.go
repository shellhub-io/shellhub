package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserName(t *testing.T) {
	tests := []struct {
		description string
		value       string
		want        bool
	}{
		{
			description: "failed when the name is empty",
			value:       "",
			want:        false,
		},
		{
			description: "failed when the name is too long",
			value:       "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaax",
			want:        false,
		},
		{
			description: "success when the name contains spaces",
			value:       "test test",
			want:        true,
		},
		{
			description: "success when the name is valid",
			value:       "test",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			data := struct {
				Name string `validate:"required,name"`
			}{
				Name: tt.value,
			}

			ok, _ := New().Struct(data)

			assert.Equal(t, tt.want, ok)
		})
	}
}

func TestUserUsername(t *testing.T) {
	tests := []struct {
		description string
		value       string
		want        bool
	}{
		{
			description: "failed when the username is empty",
			value:       "",
			want:        false,
		},
		{
			description: "failed when the username is too short",
			value:       "a",
			want:        false,
		},
		{
			description: "failed when the username is too long",
			value:       "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaax",
			want:        false,
		},
		{
			description: "failed when the username contains invalid characters",
			value:       "test$",
			want:        false,
		},
		{
			description: "failed when the username contains spaces",
			value:       "test test",
			want:        false,
		},
		{
			description: "success when the username is valid",
			value:       "test",
			want:        true,
		},
		{
			description: "success when the username is valid with @",
			value:       "test@",
			want:        true,
		},
		{
			description: "success when the username is valid with -",
			value:       "test-",
			want:        true,
		},
		{
			description: "success when the username is valid with _",
			value:       "test_",
			want:        true,
		},
		{
			description: "success when the username is valid with .",
			value:       "test.",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			data := struct {
				Username string `validate:"required,username"`
			}{
				Username: tt.value,
			}

			ok, _ := New().Struct(data)

			assert.Equal(t, tt.want, ok)
		})
	}
}

func TestUserEmail(t *testing.T) {
	tests := []struct {
		description string
		value       string
		want        bool
	}{
		{
			description: "failed when the email is empty",
			value:       "",
			want:        false,
		},
		{
			description: "failed when the email is invalid",
			value:       "test",
			want:        false,
		},
		{
			description: "success when the email is valid",
			value:       "test@shellhub.io",
			want:        true,
		},
		{
			description: "success when the email is valid with +",
			value:       "test+go@shellhub.io",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			data := struct {
				Email string `validate:"required,email"`
			}{
				Email: tt.value,
			}

			ok, _ := New().Struct(data)

			assert.Equal(t, tt.want, ok)
		})
	}
}

func TestUserPassword(t *testing.T) {
	tests := []struct {
		description string
		value       string
		want        bool
	}{
		{
			description: "failed when the password is empty",
			value:       "",
			want:        false,
		},
		{
			description: "failed when the password is too short",
			value:       "a",
			want:        false,
		},
		{
			description: "failed when the password is too long",
			value:       "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaax",
			want:        false,
		},
		{
			description: "success when the password is valid",
			value:       "password",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			data := struct {
				Password string `validate:"required,password"`
			}{
				Password: tt.value,
			}

			ok, _ := New().Struct(data)

			assert.Equal(t, tt.want, ok)
		})
	}
}

func TestDeviceName(t *testing.T) {
	tests := []struct {
		description string
		value       string
		want        bool
	}{
		{
			description: "failed when the device name is empty",
			value:       "",
			want:        false,
		},
		{
			description: "failed when the device name is too long",
			value:       "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaax",
			want:        false,
		},
		{
			description: "failed when the device name is uppercase",
			value:       "TEST",
			want:        false,
		},
		{
			description: "failed when the device name contains invalid characters",
			value:       "test$",
			want:        false,
		},
		{
			description: "success when the device name is valid",
			value:       "test",
			want:        true,
		},
		{
			description: "success when the device name is valid with -",
			value:       "test-",
			want:        true,
		},
		{
			description: "success when the device name is valid with _",
			value:       "test_",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			data := struct {
				DeviceName string `validate:"required,device_name"`
			}{
				DeviceName: tt.value,
			}

			ok, _ := New().Struct(data)

			assert.Equal(t, tt.want, ok)
		})
	}
}
