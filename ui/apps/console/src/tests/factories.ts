import type {
  AnnouncementShort,
  Device,
  FirewallRulesResponse,
  Namespace,
  PublicKeyResponse,
  Session,
  Tag,
  UserAuth,
} from "@/client";

export function mockUserAuth(overrides: Partial<UserAuth> = {}): UserAuth {
  return {
    token: "jwt-token",
    id: "user-123",
    origin: "local",
    user: "admin",
    name: "Admin User",
    email: "admin@test.com",
    recovery_email: "recovery@test.com",
    tenant: "tenant-456",
    role: "owner",
    mfa: false,
    admin: false,
    max_namespaces: -1,
    ...overrides,
  };
}

export function mockNamespace(overrides: Partial<Namespace> = {}): Namespace {
  return {
    name: "my-namespace",
    owner: "user-123",
    tenant_id: "tenant-456",
    members: [
      {
        id: "user-123",
        added_at: "2024-01-01T00:00:00Z",
        role: "owner",
        email: "admin@test.com",
        account_status: "confirmed",
      },
    ],
    settings: {
      session_record: false,
      connection_announcement: "",
      ssh_access_mode: "legacy",
      ssh_legacy_allowed: true,
    },
    max_devices: 3,
    created_at: "2024-01-01T00:00:00Z",
    billing: null,
    devices_pending_count: 0,
    devices_accepted_count: 0,
    devices_rejected_count: 0,
    ...overrides,
  };
}

export function mockTag(overrides: Partial<Tag> = {}): Tag {
  return {
    name: "tag",
    tenant_id: "tenant-456",
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
    ...overrides,
  };
}

export function mockDevice(overrides: Partial<Device> = {}): Device {
  return {
    uid: "device-uid-1",
    name: "my-device",
    status: "accepted",
    online: true,
    namespace: "my-namespace",
    tenant_id: "tenant-456",
    tags: [],
    last_seen: "2024-01-15T00:00:00Z",
    created_at: "2024-01-01T00:00:00Z",
    identity: { mac: "aa:bb:cc:dd:ee:ff" },
    info: {
      id: "ubuntu",
      pretty_name: "Ubuntu 22.04",
      arch: "x86_64",
      platform: "native",
      version: "0.14.0",
    },
    remote_addr: "1.2.3.4",
    ...overrides,
  };
}

export function mockSession(overrides: Partial<Session> = {}): Session {
  const device = mockDevice();
  return {
    uid: "session-1",
    device_uid: device.uid,
    device,
    tenant_id: device.tenant_id,
    username: "root",
    ip_address: "192.168.0.1",
    started_at: "2024-01-01T00:00:00Z",
    last_seen: "2024-01-01T01:00:00Z",
    active: false,
    authenticated: true,
    recorded: false,
    type: "term",
    term: "xterm",
    web: false,
    position: { latitude: 0, longitude: 0 },
    events: { types: ["shell"], seats: [] },
    ...overrides,
  };
}

export function mockFirewallRule(
  overrides: Partial<FirewallRulesResponse> = {},
): FirewallRulesResponse {
  return {
    id: "rule-1",
    tenant_id: "tenant-456",
    priority: 1,
    action: "allow",
    active: true,
    source_ip: ".*",
    username: ".*",
    filter: { hostname: ".*", tags: [] },
    ...overrides,
  };
}

export function mockPublicKey(
  overrides: Partial<PublicKeyResponse> = {},
): PublicKeyResponse {
  return {
    data: btoa("ssh-rsa AAAA test"),
    fingerprint: "aa:bb:cc:dd",
    created_at: "2024-01-01T00:00:00Z",
    tenant_id: "tenant-456",
    name: "my-key",
    filter: { hostname: ".*", tags: [] },
    username: ".*",
    ...overrides,
  };
}

export function mockAnnouncement(
  overrides: Partial<AnnouncementShort> = {},
): AnnouncementShort {
  return {
    uuid: "uuid-0001",
    title: "Welcome to ShellHub",
    date: "2024-06-01T10:00:00.000Z",
    ...overrides,
  };
}
