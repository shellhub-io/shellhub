package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/require"
)

type sentEvent struct {
	Type   string
	UID    string
	Status models.DeviceStatus
}

func (r *webhookReceiver) events(t *testing.T) []sentEvent {
	t.Helper()

	var events []sentEvent
	for _, got := range r.requests() {
		var event deliveredEvent
		require.NoError(t, json.Unmarshal(got.body, &event))
		if event.Type == string(models.WebhookEventDeviceRemoved) {
			require.NotNil(t, event.Data.Device.RemovedAt, "a removed device says when it was removed")
		}
		events = append(events, sentEvent{Type: event.Type, UID: event.Data.Device.UID, Status: event.Data.Device.Status})
	}

	return events
}

func subscribeToEveryDeviceEvent(t *testing.T, e *enrollmentE2E) *webhookReceiver {
	t.Helper()

	receiver := newWebhookReceiver(t)
	e.createWebhook(t, receiver.URL, models.WebhookEventTypes...)

	return receiver
}

func TestWebhookSourcesE2E_ARegisteredDeviceSendsPending(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	receiver := subscribeToEveryDeviceEvent(t, e)

	uid := e.enroll(t, "aa:bb:cc:dd:ee:01", "")
	e.enroll(t, "aa:bb:cc:dd:ee:01", "")
	e.deliver(t)

	require.Equal(t, []sentEvent{{Type: "device.pending", UID: uid, Status: models.DeviceStatusPending}}, receiver.events(t),
		"a device authenticating again while pending is not a new event")
}

func TestWebhookSourcesE2E_APendingEventCarriesTheNamespaceAndTheKeysTags(t *testing.T) {
	e := setupEnrollmentE2E(t)
	receiver := newWebhookReceiver(t)
	e.createWebhook(t, receiver.URL, models.WebhookEventDevicePending)

	tagID, err := e.st.TagCreate(context.Background(), &models.Tag{TenantID: e.tenantID, Name: "fleet"})
	require.NoError(t, err)
	e.provisioningKey(t, digest(0x01), "manual", models.ProvisioningKeyModeManual, models.ProvisioningKeyTypeUser, func(k *models.ProvisioningKey) {
		clearSecret(k)
		k.TagIDs = []string{tagID}
	})

	e.enroll(t, "aa:bb:cc:dd:ee:01", plaintextFor(0x01))
	e.deliver(t)

	got := receiver.requests()
	require.Len(t, got, 1)

	var event deliveredEvent
	require.NoError(t, json.Unmarshal(got[0].body, &event))
	require.Equal(t, "e2e", event.Data.Device.Namespace)
	require.Len(t, event.Data.Device.Tags, 1)
	require.Equal(t, "fleet", event.Data.Device.Tags[0].Name)
}

func TestWebhookSourcesE2E_DeletingAnAcceptedDeviceSendsRemovedAndReRegisteringSendsPending(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	uid := e.enroll(t, "aa:bb:cc:dd:ee:01", "")
	require.NoError(t, e.setStatus(t, uid, models.DeviceStatusAccepted))
	receiver := subscribeToEveryDeviceEvent(t, e)

	require.NoError(t, e.svc.DeleteDevice(context.Background(), models.UID(uid), e.tenantID))
	e.enroll(t, "aa:bb:cc:dd:ee:01", "")
	e.deliver(t)

	require.ElementsMatch(t, []sentEvent{
		{Type: "device.removed", UID: uid, Status: models.DeviceStatusRemoved},
		{Type: "device.pending", UID: uid, Status: models.DeviceStatusPending},
	}, receiver.events(t), "delivery order is not guaranteed")
}

func TestWebhookSourcesE2E_DeletingARemovedDeviceSendsNoSecondRemoved(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	uid := e.enroll(t, "aa:bb:cc:dd:ee:01", "")
	require.NoError(t, e.setStatus(t, uid, models.DeviceStatusAccepted))
	receiver := subscribeToEveryDeviceEvent(t, e)

	require.NoError(t, e.svc.DeleteDevice(context.Background(), models.UID(uid), e.tenantID))
	require.NoError(t, e.svc.DeleteDevice(context.Background(), models.UID(uid), e.tenantID))
	e.deliver(t)

	require.Equal(t, []sentEvent{{Type: "device.removed", UID: uid, Status: models.DeviceStatusRemoved}}, receiver.events(t))
}

func TestWebhookSourcesE2E_RejectingADeviceSendsRejected(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	receiver := newWebhookReceiver(t)
	e.createWebhook(t, receiver.URL, models.WebhookEventDeviceRejected)

	uid := e.enroll(t, "aa:bb:cc:dd:ee:01", "")
	require.NoError(t, e.setStatus(t, uid, models.DeviceStatusRejected))
	e.deliver(t)

	require.Equal(t, []sentEvent{{Type: "device.rejected", UID: uid, Status: models.DeviceStatusRejected}}, receiver.events(t))
}

func TestWebhookSourcesE2E_AnAutomaticProvisioningKeySendsAccepted(t *testing.T) {
	e := setupEnrollmentE2E(t)
	receiver := newWebhookReceiver(t)
	e.createWebhook(t, receiver.URL, models.WebhookEventDeviceAccepted)
	e.provisioningKey(t, digest(0x01), "auto", models.ProvisioningKeyModeAutomatic, models.ProvisioningKeyTypeUser, clearSecret)

	uid := e.enroll(t, "aa:bb:cc:dd:ee:01", plaintextFor(0x01))
	e.deliver(t)

	require.Equal(t, []sentEvent{{Type: "device.accepted", UID: uid, Status: models.DeviceStatusAccepted}}, receiver.events(t))
}

func TestWebhookSourcesE2E_AnEnrollmentCallbackSendsItsDecision(t *testing.T) {
	e := setupEnrollmentE2E(t)
	receiver := newWebhookReceiver(t)
	e.createWebhook(t, receiver.URL, models.WebhookEventDeviceAccepted)

	var callbackURL string
	decider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			CallbackURL string `json:"callback_url"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		callbackURL = payload.CallbackURL
		_ = json.NewEncoder(w).Encode(map[string]string{"decision": "defer"})
	}))
	t.Cleanup(decider.Close)

	e.provisioningKey(t, digest(0x50), "webhook", models.ProvisioningKeyModeWebhook, models.ProvisioningKeyTypeUser, func(k *models.ProvisioningKey) {
		k.WebhookURL = decider.URL
		k.WebhookSecret = "s3cr3t"
		clearSecret(k)
	})

	uid := e.enroll(t, "aa:bb:cc:dd:ee:50", plaintextFor(0x50))
	e.deliver(t)
	require.Empty(t, receiver.events(t), "a deferred decision accepts nothing yet")

	token := callbackURL[strings.LastIndex(callbackURL, "/")+1:]
	require.NoError(t, e.svc.ResolveEnrollmentCallback(context.Background(), &requests.EnrollmentCallback{Token: token, Decision: "accept"}))
	e.deliver(t)

	require.Equal(t, []sentEvent{{Type: "device.accepted", UID: uid, Status: models.DeviceStatusAccepted}}, receiver.events(t))
}

func TestWebhookSourcesE2E_AcceptingAPairingSendsAccepted(t *testing.T) {
	e := setupOwnershipE2E(t)
	receiver := newWebhookReceiver(t)
	e.createWebhook(t, receiver.URL, models.WebhookEventDeviceAccepted)

	accepted := e.pair(t, e.ownerID, "aa:bb:cc:dd:ee:01", "pk-1")
	e.deliver(t)

	require.Equal(t, []sentEvent{{Type: "device.accepted", UID: accepted.UID, Status: models.DeviceStatusAccepted}}, receiver.events(t))
}

func TestWebhookSourcesE2E_AReRegistrationThatFailsToSaveSendsNothing(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	ctx := context.Background()
	uid := e.enroll(t, "aa:bb:cc:dd:ee:01", "")
	require.NoError(t, e.setStatus(t, uid, models.DeviceStatusAccepted))
	require.NoError(t, e.svc.DeleteDevice(ctx, models.UID(uid), e.tenantID))
	receiver := subscribeToEveryDeviceEvent(t, e)

	_, err := e.db.NewRaw(`
		CREATE FUNCTION refuse_revival() RETURNS trigger AS $$
		BEGIN
			IF OLD.removed_at IS NOT NULL AND NEW.removed_at IS NULL THEN
				RAISE EXCEPTION 'revival refused';
			END IF;
			RETURN NEW;
		END $$ LANGUAGE plpgsql;
		CREATE TRIGGER refuse_revival BEFORE UPDATE ON devices FOR EACH ROW EXECUTE FUNCTION refuse_revival();
	`).Exec(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := e.db.NewRaw("DROP TRIGGER refuse_revival ON devices; DROP FUNCTION refuse_revival();").Exec(context.Background())
		require.NoError(t, err)
	})

	_, err = e.svc.AuthDevice(ctx, requests.DeviceAuth{
		TenantID:  e.tenantID,
		Hostname:  "host-aa:bb:cc:dd:ee:01",
		Identity:  &requests.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:01"},
		PublicKey: "pk-aa:bb:cc:dd:ee:01",
		RealIP:    "203.0.113.7",
	})
	require.Error(t, err)
	e.deliver(t)

	require.Empty(t, receiver.events(t), "the device stayed removed, so nothing became pending")
}

func TestWebhookSourcesE2E_DeletingAPendingDeviceSendsRemoved(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	uid := e.enroll(t, "aa:bb:cc:dd:ee:01", "")
	receiver := subscribeToEveryDeviceEvent(t, e)

	require.NoError(t, e.svc.DeleteDevice(context.Background(), models.UID(uid), e.tenantID))
	e.deliver(t)

	require.Equal(t, []sentEvent{{Type: "device.removed", UID: uid, Status: models.DeviceStatusRemoved}}, receiver.events(t))
}

func TestWebhookSourcesE2E_AMergeSendsRemovedForTheReplacedDevice(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	const mac = "aa:bb:cc:dd:ee:01"

	old := e.enrollAs(t, "field-box", mac, "pk-old", "")
	require.NoError(t, e.setStatus(t, old, models.DeviceStatusAccepted))
	reflashed := e.enrollAs(t, "fresh-image", mac, "pk-reflashed", "")
	receiver := subscribeToEveryDeviceEvent(t, e)

	require.NoError(t, e.setStatus(t, reflashed, models.DeviceStatusAccepted))
	e.deliver(t)

	require.ElementsMatch(t, []sentEvent{
		{Type: "device.removed", UID: old, Status: models.DeviceStatusRemoved},
		{Type: "device.accepted", UID: reflashed, Status: models.DeviceStatusAccepted},
	}, receiver.events(t))
}

func TestWebhookSourcesE2E_EphemeralCleanupSendsRemovedWithTheNamespaceAndTags(t *testing.T) {
	e := setupEnrollmentE2E(t)
	tagID, err := e.st.TagCreate(context.Background(), &models.Tag{TenantID: e.tenantID, Name: "fleet"})
	require.NoError(t, err)
	e.provisioningKey(t, digest(0x40), "ephemeral", models.ProvisioningKeyModeAutomatic, models.ProvisioningKeyTypeUser, func(k *models.ProvisioningKey) {
		clearSecret(k)
		k.Ephemeral = true
		k.EphemeralTimeout = 1
		k.TagIDs = []string{tagID}
	})
	uid := e.enroll(t, "aa:bb:cc:dd:ee:01", plaintextFor(0x40))
	require.Equal(t, models.DeviceStatusAccepted, e.status(t, uid))
	receiver := subscribeToEveryDeviceEvent(t, e)

	e.clock.Set(now.Add(time.Hour))
	require.NoError(t, e.svc.EphemeralCleanup()(context.Background()))
	e.deliver(t)

	require.Equal(t, []sentEvent{{Type: "device.removed", UID: uid, Status: models.DeviceStatusRemoved}}, receiver.events(t))

	var event deliveredEvent
	require.NoError(t, json.Unmarshal(receiver.requests()[0].body, &event))
	require.Equal(t, "e2e", event.Data.Device.Namespace)
	require.Len(t, event.Data.Device.Tags, 1)
	require.Equal(t, "fleet", event.Data.Device.Tags[0].Name)
}

func TestWebhookSourcesE2E_RemovingAMemberSendsRemovedForTheirDevices(t *testing.T) {
	e := setupOwnershipE2E(t)
	member := e.member(t, "member", authorizer.RoleOperator)
	owned := e.pair(t, member, "aa:bb:cc:dd:ee:01", "pk-1")
	receiver := subscribeToEveryDeviceEvent(t, e.enrollmentE2E)

	_, err := e.svc.RemoveNamespaceMember(context.Background(), &requests.NamespaceRemoveMember{UserID: e.ownerID, TenantID: e.tenantID, MemberID: member})
	require.NoError(t, err)
	e.deliver(t)

	require.Equal(t, []sentEvent{{Type: "device.removed", UID: owned.UID, Status: models.DeviceStatusRemoved}}, receiver.events(t))
}
