package interfaces

import (
	"github.com/shellhub-io/shellhub/api/apicontext"
)

type Handler interface {
	AuthRequest(c apicontext.Context) error
	AuthDevice(c apicontext.Context) error
	AuthUser(c apicontext.Context) error
	AuthUserInfo(c apicontext.Context) error
	AuthGetToken(c apicontext.Context) error
	AuthSwapToken(c apicontext.Context) error
	AuthPublicKey(c apicontext.Context) error
	// AuthMiddleware(next apicontext.HandlerFunc) apicontext.HandlerFunc
	// DecodeMap(input, output interface{}) error
	GetDeviceList(c apicontext.Context) error
	GetDevice(c apicontext.Context) error
	DeleteDevice(c apicontext.Context) error
	RenameDevice(c apicontext.Context) error
	OfflineDevice(c apicontext.Context) error
	LookupDevice(c apicontext.Context) error
	UpdatePendingStatus(c apicontext.Context) error
	GetNamespaceList(c apicontext.Context) error
	CreateNamespace(c apicontext.Context) error
	GetNamespace(c apicontext.Context) error
	DeleteNamespace(c apicontext.Context) error
	EditNamespace(c apicontext.Context) error
	AddNamespaceUser(c apicontext.Context) error
	RemoveNamespaceUser(c apicontext.Context) error
	EditSessionRecordStatus(c apicontext.Context) error
	GetSessionRecord(c apicontext.Context) error
	GetSessionList(c apicontext.Context) error
	GetSession(c apicontext.Context) error
	SetSessionAuthenticated(c apicontext.Context) error
	CreateSession(c apicontext.Context) error
	FinishSession(c apicontext.Context) error
	RecordSession(c apicontext.Context) error
	PlaySession(c apicontext.Context) error
	DeleteRecordedSession(c apicontext.Context) error
	GetPublicKeys(c apicontext.Context) error
	GetPublicKey(c apicontext.Context) error
	CreatePublicKey(c apicontext.Context) error
	UpdatePublicKey(c apicontext.Context) error
	DeletePublicKey(c apicontext.Context) error
	CreatePrivateKey(c apicontext.Context) error
	EvaluateKeyHostname(c apicontext.Context) error
	GetStats(c apicontext.Context) error
	UpdateUserData(c apicontext.Context) error
	UpdateUserPassword(c apicontext.Context) error
}
