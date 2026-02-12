import { ISession } from "@/interfaces/ISession";
import { mockDeviceForSession } from "./device";

/**
 * Mock session data for testing.
 * Provides a complete session object with associated device.
 */
export const mockSession: ISession = {
  uid: "session-1",
  device_uid: "device-1",
  username: "user",
  authenticated: true,
  active: true,
  recorded: true,
  ip_address: "192.168.1.1",
  device: {
    uid: "device-1",
    name: "00-00-00-00-00-01",
  } as typeof mockDeviceForSession,
  tenant_id: "fake-tenant-data",
  term: "none",
  type: "none",
  started_at: "2026-01-08T00:00:00.000Z",
  last_seen: "2026-01-08T00:00:00.000Z",
  position: { longitude: 0, latitude: 0 },
};

/**
 * Mock detailed session for session details view.
 * Includes full device information and session metadata.
 */
export const mockDetailedSession: ISession = {
  uid: "1",
  device_uid: "1",
  device: mockDeviceForSession,
  tenant_id: "fake-tenant-data",
  username: "test",
  ip_address: "192.168.0.1",
  started_at: "2025-01-02T00:00:00.000Z",
  last_seen: "2025-01-02T00:00:00.000Z",
  active: true,
  authenticated: true,
  recorded: true,
  type: "ssh",
  term: "xterm-256color",
  position: { latitude: 0, longitude: 0 },
};
