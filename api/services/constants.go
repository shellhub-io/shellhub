package services

import (
	"time"
)

// API Key expiration constants
const (
	APIKeyExpirationDays30  = 30
	APIKeyExpirationDays60  = 60
	APIKeyExpirationDays90  = 90
	APIKeyExpirationDays365 = 365
	APIKeyExpirationNever   = -1
)

// Cache duration constants
const (
	DeviceAuthCacheDuration = 30 * time.Second
	MFATokenCacheDuration   = 30 * time.Minute
	APIKeyCacheDuration     = 2 * time.Minute
	TokenCacheDuration      = 72 * time.Hour
)

// Member invitation constants
const (
	MemberInvitationExpirationDays = 7
	HoursPerDay                    = 24
)

// RSA key size constants
const (
	RSAKeySize = 4096
)

// Task processing constants
const (
	TaskProcessingSleepDuration = 100 * time.Millisecond
)

// Device constants
const (
	DeviceUsageLimit = 3
)
