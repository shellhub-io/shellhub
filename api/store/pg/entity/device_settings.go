package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/uptrace/bun"
)

type DeviceSettings struct {
	bun.BaseModel `bun:"table:device_settings"`

	ID                   string    `bun:"id,pk,type:uuid,nullzero,default:gen_random_uuid()"`
	DeviceID             string    `bun:"device_id,type:varchar,unique"`
	AllowPassword        bool      `bun:"allow_password"`
	AllowPublicKey       bool      `bun:"allow_public_key"`
	AllowRoot            bool      `bun:"allow_root"`
	AllowEmptyPasswords  bool      `bun:"allow_empty_passwords"`
	AllowTTY             bool      `bun:"allow_tty"`
	AllowTCPForwarding   bool      `bun:"allow_tcp_forwarding"`
	AllowWebEndpoints    bool      `bun:"allow_web_endpoints"`
	AllowSFTP            bool      `bun:"allow_sftp"`
	AllowAgentForwarding bool      `bun:"allow_agent_forwarding"`
	CreatedAt            time.Time `bun:"created_at"`
	UpdatedAt            time.Time `bun:"updated_at"`
}

func DeviceSettingsFromModel(ssh *models.SSHSettings, deviceID string) DeviceSettings {
	if ssh == nil {
		return DeviceSettings{
			ID:                   uuid.Generate(),
			DeviceID:             deviceID,
			AllowPassword:        true,
			AllowPublicKey:       true,
			AllowRoot:            true,
			AllowEmptyPasswords:  true,
			AllowTTY:             true,
			AllowTCPForwarding:   true,
			AllowWebEndpoints:    true,
			AllowSFTP:            true,
			AllowAgentForwarding: true,
		}
	}

	return DeviceSettings{
		ID:                   uuid.Generate(),
		DeviceID:             deviceID,
		AllowPassword:        ssh.AllowPassword,
		AllowPublicKey:       ssh.AllowPublicKey,
		AllowRoot:            ssh.AllowRoot,
		AllowEmptyPasswords:  ssh.AllowEmptyPasswords,
		AllowTTY:             ssh.AllowTTY,
		AllowTCPForwarding:   ssh.AllowTCPForwarding,
		AllowWebEndpoints:    ssh.AllowWebEndpoints,
		AllowSFTP:            ssh.AllowSFTP,
		AllowAgentForwarding: ssh.AllowAgentForwarding,
	}
}

func DeviceSettingsToModel(settings *DeviceSettings) *models.SSHSettings {
	if settings == nil {
		return nil
	}

	return &models.SSHSettings{
		AllowPassword:        settings.AllowPassword,
		AllowPublicKey:       settings.AllowPublicKey,
		AllowRoot:            settings.AllowRoot,
		AllowEmptyPasswords:  settings.AllowEmptyPasswords,
		AllowTTY:             settings.AllowTTY,
		AllowTCPForwarding:   settings.AllowTCPForwarding,
		AllowWebEndpoints:    settings.AllowWebEndpoints,
		AllowSFTP:            settings.AllowSFTP,
		AllowAgentForwarding: settings.AllowAgentForwarding,
	}
}
