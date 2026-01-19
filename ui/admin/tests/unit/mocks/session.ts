import { IAdminSession } from "@admin/interfaces/ISession";

export const mockSession: IAdminSession = {
  uid: "session-123",
  device_uid: "device-123",
  device: {
    uid: "device-123",
    name: "test-device",
    identity: {
      mac: "00:11:22:33:44:55",
    },
    info: {
      id: "device-123",
      pretty_name: "Ubuntu 22.04",
      version: "v1.0.0",
      arch: "x86_64",
      platform: "linux",
    },
    public_key: "ssh-rsa AAAA...",
    tenant_id: "tenant-123",
    last_seen: "2024-01-10T12:00:00Z",
    online: true,
    namespace: "default",
    status: "accepted" as const,
    created_at: "2024-01-01T00:00:00Z",
    status_updated_at: "2024-01-01T00:00:00Z",
    remote_addr: "192.168.1.100",
    position: { latitude: 0, longitude: 0 },
    tags: [],
  },
  tenant_id: "tenant-123",
  username: "testuser",
  ip_address: "192.168.1.100",
  started_at: "2024-01-10T12:00:00Z",
  last_seen: "2024-01-10T12:30:00Z",
  active: true,
  authenticated: true,
  recorded: false,
  type: "ssh",
  position: { latitude: 0, longitude: 0 },
  term: "xterm-256color",
};

export const mockSessions: IAdminSession[] = [
  { ...mockSession, uid: "session-1", username: "alice" },
  { ...mockSession, uid: "session-2", username: "bob", active: false, authenticated: false },
  { ...mockSession, uid: "session-3", username: "charlie" },
];
