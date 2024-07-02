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

func TestKeyPEM(t *testing.T) {
	tests := []struct {
		description string
		value       string
		want        bool
	}{
		{
			description: "failed when the private key is empty",
			value:       "",
			want:        false,
		},
		{
			description: "failed when the private key does not have the header",
			value: `
MC4CAQAwBQYDK2VwBCIEIA2Ecxi0E2XsKUNRYBv98VRbpsjl/kD7l7XOa/aKYitU
-----END PRIVATE KEY-----`,
			want: false,
		},
		{
			description: "failed when the private key does not have the footer",
			value: `-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIA2Ecxi0E2XsKUNRYBv98VRbpsjl/kD7l7XOa/aKYitU
`,
			want: false,
		},
		{
			description: "failed when the private key does not have header neither footer",
			value: `
MC4CAQAwBQYDK2VwBCIEIA2Ecxi0E2XsKUNRYBv98VRbpsjl/kD7l7XOa/aKYitU
`,
			want: false,
		},
		{
			description: "success when the private key is a valid ED25519",
			value: `-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIA2Ecxi0E2XsKUNRYBv98VRbpsjl/kD7l7XOa/aKYitU
-----END PRIVATE KEY-----`,
			want: true,
		},
		{
			description: "success when the private key is a valid RSA4096",
			value: `-----BEGIN PRIVATE KEY-----
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
rRrVfCKhbJKpjAZq5U9VKt9LcGe9kA==
-----END PRIVATE KEY-----`,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			data := struct {
				Key string `validate:"required,privateKeyPEM"`
			}{
				Key: tt.value,
			}

			ok, _ := New().Struct(data)

			assert.Equal(t, tt.want, ok)
		})
	}
}

func TestCertPEM(t *testing.T) {
	tests := []struct {
		description string
		value       string
		want        bool
	}{
		{
			description: "failed when the cert is empty",
			value:       "",
			want:        false,
		},
		{
			description: "failed when the cert does not have the header",
			value: `
MC4CAQAwBQYDK2VwBCIEIA2Ecxi0E2XsKUNRYBv98VRbpsjl/kD7l7XOa/aKYitU
-----END CERTIFICATE-----`,
			want: false,
		},
		{
			description: "failed when the cert does not have the footer",
			value: `-----BEGIN CERTIFICATE-----
MC4CAQAwBQYDK2VwBCIEIA2Ecxi0E2XsKUNRYBv98VRbpsjl/kD7l7XOa/aKYitU
`,
			want: false,
		},
		{
			description: "failed when the cert does not have header neither footer",
			value: `
MC4CAQAwBQYDK2VwBCIEIA2Ecxi0E2XsKUNRYBv98VRbpsjl/kD7l7XOa/aKYitU
`,
			want: false,
		},
		{
			description: "success when the cert is a valid",
			value: `-----BEGIN CERTIFICATE-----
MIIFUTCCAzmgAwIBAgIUGOBHWPTiCbwt8iLWYNZwKTDbONUwDQYJKoZIhvcNAQEL
BQAwWzELMAkGA1UEBhMCQlIxDjAMBgNVBAgMBUJhaGlhMRQwEgYDVQQHDAtYaXF1
ZS1YaXF1ZTEQMA4GA1UECgwHSGVucnknczEUMBIGA1UEAwwLZGVsbGcxNTU1MTAw
HhcNMjQwNTI5MTk1NzA0WhcNMjUwNTI5MTk1NzA0WjARMQ8wDQYDVQQDDAZjbGll
bnQwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQDt92hrf1PDvCAwNaEv
1xjfL2QCyEsA7zxBGPIIA5ETsB41LW3yS98oy8F/L72BDEepmsw49DaQLbIZJrjX
yT4dtYKN9oPgv5uwwmwPrWexsiDiVA968DOgSWj4S4MIDLAwd3gBqrQgqFutIwgt
58KzhKYn/J9+1q/G8ecKzRre7c7/PQbCHEH4A/XiIudyuSf49ziU+U7dq9rZIAiy
G2xMAKZnjANP0dQj8gaAJCD1qesyoIUXrHCuesrZEEY1gov6ZxUeR62KQgIFJDQ8
nrGgPRc/AjNcLhLKH5xaRqfbEv3WyYw1Ag4Fc1ZtIOgLbMr9BRcxnrhCAIBD4ASU
+63N5zxC/K0JOPy4iSa8+uMXoYD4eJIUI4e9cuAp976zCsrd6d2QEDZmly2/KGrc
TunlNQ49LfqV9LQWnumRoQ5vhlOHWAQmY48svf45PGeQrrbLUfV24uO4ZzwnCCCH
BUUUwTlasZi1zwHgZ1rmqOjemnGn6HJ9T64tFypUQKOiS5NxeAajszQLf3GfIE8Z
ibE+uxZQyvRexmyUt+RaOQfyAKtnczyOd9LU4/JqVtbKYtuxltw503gS+RuzxcHu
FEv/takSszbr9mKAj/pT0MEKE9nJLP2gcqw0j2fdjfWWejPGwWlxJ98sPlw8eh4K
NOtphFmgbjIUTrjfS6G+3cbOwQIDAQABo1cwVTATBgNVHSUEDDAKBggrBgEFBQcD
AjAdBgNVHQ4EFgQUzvw/tD0WsD5q2K2wSokjLEReY6wwHwYDVR0jBBgwFoAU9Nw4
MqfdGEeRWXI2H1ChuK2k9qEwDQYJKoZIhvcNAQELBQADggIBAIQp2CQyPjaqbXZc
hiR0VWwAyifttrHJJ59VCFovH4/LW8oPbg8w7JP4bfm9iTbo7yTqDV6BfOWat4Qf
T5o0HVcmxKEY7X6bEAmTFfSsNs6NTuaIE8QSFpJpKvLGIjulSqhayjSPuqJavluc
lGa1vUPeIqZAKPDFwrdqMXg/Q7DMhg9su7QPfNVu2E2Hrq++PaXPnWZlu3/yu5FH
2qjoS/xeG8QL8STzqVxqsmcGXkI8FYT2Goidb5eNPSqJflntgm0FzZ/YYvCpZbdC
8/Qjg+CnopfuyLS72iZvW4tSv/9plBsiu6UqhbjBz9xQZbBDpvUOyUvK+L8URmWB
21xTMtqdqk3iG3qAFGnaz0EM0Tg4MEopzYMieob2XoxjSH55ykj33LF/sZeNVPzK
gXi2bqLzL5I1kTPF+Irrg5z7FBTcXRVdPcvqjxGfbyVVmaxNmC26ozIF94rYUOIr
JeUB+pKG1xX/fhUAMeLvEkJ6GOl6ldnTqPJrNAZzwAqW5ra0H9kIbmf1fGPpezaa
KdtGUV3wYjChWAuSa0S3mP1qD+sRNS5NtR7efemmoUbR+hCg2Vyo5osRSJ9dkQJf
PNcoe7LEpZdYQvPI5v1fqVcFpOdOCckDdaGb3XPpd69LGdFD0jHOzF9eIavv9ewV
eiDIAGdPArZi+JWdNsp+TK4MJjcy
-----END CERTIFICATE-----`,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			data := struct {
				Cert string `validate:"required,certPEM"`
			}{
				Cert: tt.value,
			}

			ok, _ := New().Struct(data)

			assert.Equal(t, tt.want, ok)
		})
	}
}
