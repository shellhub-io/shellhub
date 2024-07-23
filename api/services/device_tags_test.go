package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/errors"
	mocksGeoIp "github.com/shellhub-io/shellhub/pkg/geoip/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

const (
	invalidUID = "Fails to find the device invalid uid"
)

func TestCreateTag(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	cases := []struct {
		description   string
		uid           models.UID
		deviceName    string
		requiredMocks func()
		expected      error
	}{
		{
			description: "Fails to find the device invalid uid",
			uid:         "invalid_uid",
			deviceName:  "device1",
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID("invalid_uid")).Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceNotFound(models.UID("invalid_uid"), errors.New("error", "", 0)),
		},
		{
			description: "Fails duplicated name",
			uid:         models.UID("uid"),
			deviceName:  "device1",
			requiredMocks: func() {
				device := &models.Device{
					UID:      "uid",
					TenantID: "tenant",
					Tags:     []string{"device1"},
				}

				mock.On("DeviceGet", ctx, models.UID("uid")).Return(device, nil).Once()
			},
			expected: NewErrTagDuplicated("device1", nil),
		},
		{
			description: "Successful create a tag for the device",
			uid:         models.UID("uid"),
			deviceName:  "device6",
			requiredMocks: func() {
				device := &models.Device{
					UID:      "uid",
					TenantID: "tenant",
					Tags:     []string{"device1"},
				}

				mock.On("DeviceGet", ctx, models.UID(device.UID)).Return(device, nil).Once()
				mock.On("DevicePushTag", ctx, models.UID(device.UID), "device6").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			locator := &mocksGeoIp.Locator{}
			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, WithLocator(locator))

			err := service.CreateDeviceTag(ctx, tc.uid, tc.deviceName)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestRemoveTag(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	cases := []struct {
		description   string
		uid           models.UID
		deviceName    string
		requiredMocks func()
		expected      error
	}{
		{
			description: invalidUID,
			uid:         "invalid_uid",
			deviceName:  "device1",
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID("invalid_uid")).Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceNotFound(models.UID("invalid_uid"), errors.New("error", "", 0)),
		},
		{
			description: "fail when device does not contain the tag",
			uid:         models.UID("uid"),
			deviceName:  "device2",
			requiredMocks: func() {
				device := &models.Device{
					UID:      "uid",
					TenantID: "tenant",
					Tags:     []string{"device1"},
				}

				mock.On("DeviceGet", ctx, models.UID("uid")).Return(device, nil).Once()
			},
			expected: NewErrTagNotFound("device2", nil),
		},
		{
			description: "fail delete a tag",
			uid:         models.UID("uid"),
			deviceName:  "device1",
			requiredMocks: func() {
				device := &models.Device{
					UID:      "uid",
					TenantID: "tenant",
					Tags:     []string{"device1"},
				}

				mock.On("DeviceGet", ctx, models.UID("uid")).Return(device, nil).Once()
				mock.On("DevicePullTag", ctx, models.UID("uid"), "device1").Return(errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "successful delete a tag",
			uid:         models.UID("uid"),
			deviceName:  "device1",
			requiredMocks: func() {
				device := &models.Device{
					UID:      "uid",
					TenantID: "tenant",
					Tags:     []string{"device1"},
				}

				mock.On("DeviceGet", ctx, models.UID("uid")).Return(device, nil).Once()
				mock.On("DevicePullTag", ctx, models.UID("uid"), "device1").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			locator := &mocksGeoIp.Locator{}
			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, WithLocator(locator))

			err := service.RemoveDeviceTag(ctx, tc.uid, tc.deviceName)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestDeviceUpdateTag(t *testing.T) {
	storemock := new(mocks.Store)

	cases := []struct {
		description   string
		uid           models.UID
		tags          []string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when tags exceeds the limit",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tags:        []string{"device1", "device2", "device3", "device4"},
			requiredMocks: func() {
			},
			expected: NewErrTagLimit(DeviceMaxTags, nil),
		},
		{
			description: "fails when device is not found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tags:        []string{"device1", "device2", "device3"},
			requiredMocks: func() {
				storemock.On("DeviceGet", context.TODO(), models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c")).Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceNotFound("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c", errors.New("error", "", 0)),
		},
		{
			description: "fails when an unexpected error occours",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tags:        []string{"device1", "device2", "device3"},
			requiredMocks: func() {
				device := &models.Device{
					UID:      "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					TenantID: "tenant",
				}
				storemock.On("DeviceGet", context.TODO(), models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c")).Return(device, nil).Once()

				tags := []string{"device1", "device2", "device3"}
				storemock.On("DeviceSetTags", context.TODO(), models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"), tags).Return(int64(0), int64(0), errors.New("error", "layer", 1)).Once()
			},
			expected: errors.New("error", "layer", 1),
		},
		{
			description: "successful update tags for the device",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tags:        []string{"device1", "device2", "device3"},
			requiredMocks: func() {
				device := &models.Device{
					UID:      "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					TenantID: "tenant",
				}
				storemock.On("DeviceGet", context.TODO(), models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c")).Return(device, nil).Once()

				tags := []string{"device1", "device2", "device3"}
				storemock.On("DeviceSetTags", context.TODO(), models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"), tags).Return(int64(1), int64(3), nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			locator := &mocksGeoIp.Locator{}
			service := NewService(store.Store(storemock), privateKey, publicKey, storecache.NewNullCache(), clientMock, WithLocator(locator))

			err := service.UpdateDeviceTag(context.TODO(), tc.uid, tc.tags)
			assert.Equal(t, tc.expected, err)
		})
	}

	storemock.AssertExpectations(t)
}
