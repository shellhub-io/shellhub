package entity

import "testing"

func TestDeviceSettingsFromModelNilDefaultsToEnabled(t *testing.T) {
	settings := DeviceSettingsFromModel(nil, "device-id")

	if !settings.AllowPassword {
		t.Fatal("expected AllowPassword to default to true")
	}
	if !settings.AllowPublicKey {
		t.Fatal("expected AllowPublicKey to default to true")
	}
	if !settings.AllowRoot {
		t.Fatal("expected AllowRoot to default to true")
	}
	if !settings.AllowEmptyPasswords {
		t.Fatal("expected AllowEmptyPasswords to default to true")
	}
	if !settings.AllowTTY {
		t.Fatal("expected AllowTTY to default to true")
	}
	if !settings.AllowTCPForwarding {
		t.Fatal("expected AllowTCPForwarding to default to true")
	}
	if !settings.AllowWebEndpoints {
		t.Fatal("expected AllowWebEndpoints to default to true")
	}
	if !settings.AllowSFTP {
		t.Fatal("expected AllowSFTP to default to true")
	}
	if !settings.AllowAgentForwarding {
		t.Fatal("expected AllowAgentForwarding to default to true")
	}
}
