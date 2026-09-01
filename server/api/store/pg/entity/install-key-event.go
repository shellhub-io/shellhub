package entity

import (
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/ssh"
)

func fingerprintFromPEM(pemKey string) string {
	if pemKey == "" {
		return ""
	}

	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return ""
	}

	var (
		pub any
		err error
	)

	switch block.Type {
	case "RSA PUBLIC KEY":
		pub, err = x509.ParsePKCS1PublicKey(block.Bytes)
	default:
		pub, err = x509.ParsePKIXPublicKey(block.Bytes)
	}

	if err != nil {
		return ""
	}

	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return ""
	}

	return ssh.FingerprintSHA256(sshPub)
}

// InstallKeyEvent is a row of install_key_events, recording one use of an install key and what
// was decided about it.
type InstallKeyEvent struct {
	bun.BaseModel `bun:"table:install_key_events"`

	ID             string     `bun:"id,pk"`
	InstallKeyID   string     `bun:"install_key_id"`
	NamespaceID    string     `bun:"namespace_id"`
	DeviceUID      string     `bun:"device_uid"`
	Hostname       string     `bun:"hostname"`
	MAC            string     `bun:"mac,nullzero"`
	InfoID         string     `bun:"info_id,nullzero"`
	InfoPrettyName string     `bun:"info_pretty_name,nullzero"`
	InfoVersion    string     `bun:"info_version,nullzero"`
	InfoArch       string     `bun:"info_arch,nullzero"`
	InfoPlatform   string     `bun:"info_platform,nullzero"`
	SourceIP       string     `bun:"source_ip,nullzero"`
	PublicKey      string     `bun:"public_key,nullzero"`
	Ephemeral      bool       `bun:"ephemeral"`
	ReRegistration bool       `bun:"re_registration"`
	CreatedAt      time.Time  `bun:"created_at"`
	DecidedStatus  string     `bun:"decided_status,nullzero"`
	DecidedAt      *time.Time `bun:"decided_at,nullzero"`
	// DeviceStatus and IsCurrent are populated by the live join / window in InstallKeyEventList; they
	// are not stored columns (used only for the live accept/reject action).
	DeviceStatus string `bun:"device_status,scanonly"`
	IsCurrent    bool   `bun:"is_current,scanonly"`
}

// InstallKeyEventFromModel projects an event into its row form.
func InstallKeyEventFromModel(model *models.InstallKeyEvent) *InstallKeyEvent {
	event := &InstallKeyEvent{
		ID:             model.ID,
		InstallKeyID:   model.InstallKeyID,
		NamespaceID:    model.TenantID,
		DeviceUID:      model.DeviceUID,
		Hostname:       model.Hostname,
		MAC:            model.MAC,
		SourceIP:       model.SourceIP,
		PublicKey:      model.PublicKey,
		Ephemeral:      model.Ephemeral,
		ReRegistration: model.ReRegistration,
		CreatedAt:      model.Timestamp,
		DecidedStatus:  string(model.DecidedStatus),
		DecidedAt:      model.DecidedAt,
	}

	if model.Info != nil {
		event.InfoID = model.Info.ID
		event.InfoPrettyName = model.Info.PrettyName
		event.InfoVersion = model.Info.Version
		event.InfoArch = model.Info.Arch
		event.InfoPlatform = model.Info.Platform
	}

	return event
}

// InstallKeyEventToModel rebuilds an event from its row.
func InstallKeyEventToModel(entity *InstallKeyEvent) *models.InstallKeyEvent {
	event := &models.InstallKeyEvent{
		ID:             entity.ID,
		InstallKeyID:   entity.InstallKeyID,
		TenantID:       entity.NamespaceID,
		DeviceUID:      entity.DeviceUID,
		Hostname:       entity.Hostname,
		MAC:            entity.MAC,
		SourceIP:       entity.SourceIP,
		PublicKey:      entity.PublicKey,
		Fingerprint:    fingerprintFromPEM(entity.PublicKey),
		Ephemeral:      entity.Ephemeral,
		ReRegistration: entity.ReRegistration,
		Timestamp:      entity.CreatedAt,
		DecidedStatus:  models.DeviceStatus(entity.DecidedStatus),
		DecidedAt:      entity.DecidedAt,
		DeviceStatus:   models.DeviceStatus(entity.DeviceStatus),
		IsCurrent:      entity.IsCurrent,
	}

	if entity.InfoID != "" || entity.InfoPrettyName != "" || entity.InfoVersion != "" || entity.InfoArch != "" || entity.InfoPlatform != "" {
		event.Info = &models.DeviceInfo{
			ID:         entity.InfoID,
			PrettyName: entity.InfoPrettyName,
			Version:    entity.InfoVersion,
			Arch:       entity.InfoArch,
			Platform:   entity.InfoPlatform,
		}
	}

	return event
}
