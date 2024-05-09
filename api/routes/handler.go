package routes

import (
	svc "github.com/shellhub-io/shellhub/api/services"
)

type Handler struct {
	service svc.Service
}

func NewHandler(s svc.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) all() []Route {
	return []Route{
		*h.createAPIKey(),
		*h.listAPIKey(),
		*h.updateAPIKey(),
		*h.deleteAPIKey(),

		*h.authRequest(),
		*h.authDevice(),
		*h.authUser(),
		*h.getUserAuthInfo(),
		*h.getUserAuthToken(),
		*h.swapUserAuthToken(),
		*h.authPublicKey(),

		*h.getDevice(),
		*h.getDeviceByPublicAddress(),
		*h.lookupDevice(),
		*h.listDevice(),
		*h.updateDevice(),
		*h.renameDevice(),
		*h.offlineDevice(),
		*h.updateDeviceStatus(),
		*h.deleteDevice(),
		*h.createDeviceTag(),
		*h.updateDeviceTag(),
		*h.removeDeviceTag(),

		*h.createNamespace(),
		*h.getNamespace(),
		*h.listNamespaces(),
		*h.updateNamespace(),
		*h.deleteNamespace(),
		*h.addNamespaceMember(),
		*h.updateNamespaceMember(),
		*h.removeNamespaceMember(),
		*h.getSessionRecord(),
		*h.updateSessionRecordStatus(),

		*h.createSession(),
		*h.getSession(),
		*h.listSessions(),
		*h.authenticateSession(),
		*h.finishSession(),
		*h.keepAliveSession(),
		*h.playRecordedSession(),
		*h.recordSession(),
		*h.deleteRecordedSession(),

		*h.createPublicKey(),
		*h.createPrivateKey(),
		*h.evaluatePublicKey(),
		*h.getPublicKey(),
		*h.listPublicKeys(),
		*h.updatePublicKey(),
		*h.deletePublicKey(),
		*h.addPublicKeyTag(),
		*h.updatePublicKeyTags(),
		*h.removePublicKeyTag(),

		*h.stats(),
		*h.systemInfo(),
		*h.systemDownloadInstallScript(),

		*h.listTags(),
		*h.updateTag(),
		*h.deleteTag(),

		*h.updateUser(),
		*h.updateUserPassword(),
	}
}
