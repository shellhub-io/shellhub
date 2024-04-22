package tunnel

import (
	"context"
	"net/http"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/connman"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	log "github.com/sirupsen/logrus"
)

// connectionHandler will be triggered whenever a new connection between an agent and the SSH server starts.
// As part of the function, the handler must perform the follow operations:
//
//  1. Updates the agent's Device.ConnectedAt field with the UTC time at which the function is called.
//  2. Sets the device from the "last_seen" cache task, starting the KeepAlive.
//  3. Increase the count of "connected_device" cache task for the agent's tenant.
func connectionHandler(client internalclient.Client, cache cache.Cache) httptunnel.ConnectionHandler {
	return func(req *http.Request) (*connman.Info, error) {
		info, err := connman.NewInfo(req.Header.Get("X-Tenant-ID"), req.Header.Get("X-Device-UID"))
		if err != nil {
			log.WithError(err).Error("failed to retrieve connection info")

			return nil, err
		}

		s, err := client.UpdateDeviceConnectionStats(info.TenantID, info.DeviceUID, time.Now(), time.Time{})
		if err != nil || s != 200 {
			log.
				WithFields(log.Fields{
					"tenant_id": info.TenantID,
					"uid":       info.DeviceUID,
					"status":    s,
				}).
				WithError(err).
				Error("failed to update device's connected_at")

			return nil, err
		}

		// TODO: context
		if err := cache.SetLastSeen(context.TODO(), info.TenantID, info.DeviceUID, info.ConnectedAt); err != nil {
			log.
				WithFields(log.Fields{
					"tenant_id": info.TenantID,
					"uid":       info.DeviceUID,
				}).
				WithError(err).
				Warn("unable to set device's last seen")
		}

		if err := client.NotifyConnectedDevicesDecrease(info.TenantID, info.DeviceUID); err != nil {
			log.
				WithFields(log.Fields{
					"tenant_id": info.TenantID,
					"uid":       info.DeviceUID,
				}).
				WithError(err).
				Error("failed to send a notification to update a connected_device status")
		}

		status := req.Header.Get("X-Device-Status")
		if err := client.NotifyConnectedDevicesIncrease(info.TenantID, status); err != nil {
			log.
				WithFields(log.Fields{
					"tenant_id": info.TenantID,
					"uid":       info.DeviceUID,
				}).
				WithError(err).
				Error("failed to send a notification to update a connected_device status")
		}

		return info, nil
	}
}

// keepAliveHandler function is responsible for informing the server that an agent is still online.
func keepAliveHandler(cache cache.Cache) httptunnel.KeepAliveHandler {
	return func(ctx context.Context, info *connman.Info) {
		if err := cache.SetLastSeen(ctx, info.TenantID, info.DeviceUID, clock.Now()); err != nil {
			log.
				WithFields(log.Fields{
					"tenant_id": info.TenantID,
					"uid":       info.DeviceUID,
				}).
				WithError(err).
				Warn("unable to set device's last seen")
		}
	}
}

// closeHandler will be triggered when a connection between an agent and the SSH server ends notifying the server
// that the associated agent is now offline. As part of the notification, the handler must perform the follow operations:
//
//  1. Updates the agent's Device.DisconnectedAt field with the UTC time at which the function is called.
//  2. Removes the device from the "last_seen" cache task.
//  3. Decreases the count of "connected_device" cache task for the agent's tenant.
func closeHandler(client internalclient.Client, cache cache.Cache) httptunnel.CloseHandler {
	return func(ctx context.Context, info *connman.Info) {
		if err := cache.DelLastSeen(ctx, info.TenantID, info.DeviceUID); err != nil {
			log.
				WithFields(log.Fields{
					"tenant_id": info.TenantID,
					"uid":       info.DeviceUID,
				}).
				WithError(err).
				Warn("unable to del device's last seen")
		}

		s, err := client.UpdateDeviceConnectionStats(info.TenantID, info.DeviceUID, time.Time{}, time.Now())
		if err != nil || s != 200 {
			log.
				WithFields(log.Fields{
					"tenant_id": info.TenantID,
					"uid":       info.DeviceUID,
				}).
				WithError(err).
				Error("failed to update device's disconnected_at")

			return
		}

		if err := client.NotifyConnectedDevicesDecrease(info.TenantID, info.DeviceUID); err != nil {
			log.
				WithFields(log.Fields{
					"tenant_id": info.TenantID,
					"uid":       info.DeviceUID,
				}).
				WithError(err).
				Error("failed to send a notification to update a connected_device status")
		}
	}
}
