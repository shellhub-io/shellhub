package businesses

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNamespaceBuilder(t *testing.T) {
	ctx := context.TODO()
	mock := new(mocks.Store)

	err := errors.New("error")

	tests := []struct {
		description  string
		user         string
		tenant       string
		name         string
		session      bool
		maxDevices   int
		riquredMocks func()
		expected     error
	}{
		{
			description: "fail when user does not exist",
			user:        "id",
			tenant:      "tenant",
			name:        "namespace",
			session:     false,
			maxDevices:  3,
			riquredMocks: func() {
				mock.On("UserGetByID", ctx, "id", false).Return(nil, 0, err).Once()
			},
			expected: NewErrUserNotFound("id", err),
		},
		{
			description: "fail when namespace already exists",
			user:        "id",
			tenant:      "tenant",
			name:        "namespace",
			session:     false,
			maxDevices:  3,
			riquredMocks: func() {
				mock.On("UserGetByID", ctx, "id", false).Return(&models.User{}, 0, nil).Once()                           // nolint: exhaustruct
				mock.On("NamespaceGetByName", ctx, "namespace").Return(&models.Namespace{Name: "namespace"}, nil).Once() // nolint: exhaustruct
			},
			expected: NewErrNamespaceDuplicated(nil),
		},
		{
			description: "success to create namespace's model",
			user:        "id",
			tenant:      "tenant",
			name:        "namespace",
			session:     false,
			maxDevices:  3,
			riquredMocks: func() {
				mock.On("UserGetByID", ctx, "id", false).Return(&models.User{}, 0, nil).Once() // nolint: exhaustruct
				mock.On("NamespaceGetByName", ctx, "namespace").Return(nil, nil).Once()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.riquredMocks()

			_, err := Namespace(ctx, mock).
				FromUser(test.user).
				WithTenantID(test.tenant).
				WithName(test.name).
				WithSessionRecord(test.session).
				Create()

			assert.Equal(t, err, test.expected)
		})
	}
}

func ExampleCreateNamespace() {
	ctx := context.TODO()

	mock := new(mocks.Store)
	mock.On("UserGetByID", ctx, "id", false).Return(new(models.User), 0, nil).Once()
	mock.On("NamespaceGetByName", ctx, "namespace").Return(nil, errors.New("error")).Once()

	_, err := Namespace(ctx, mock).
		FromUser("id").
		WithTenantID(uuid.Generate()).
		WithName("namespace").
		WithSessionRecord(false).
		Create()
	if err != nil {
		fmt.Println(err)
		// Output:
	}
}

func ExampleCreateNamespace_err() {
	ctx := context.TODO()

	mock := new(mocks.Store)
	mock.On("UserGetByID", ctx, "id", false).Return(nil, 0, errors.New("error")).Once()

	_, err := Namespace(ctx, mock).
		FromUser("id").
		WithTenantID(uuid.Generate()).
		WithName("namespace").
		WithSessionRecord(false).
		Create()
	if err != nil {
		fmt.Println(err)
		// Output: user not found: error
	}
}
