import type {
  AccessPolicy,
  Announcement,
  AnnouncementShort,
  Customer,
  Device,
  FirewallRulesResponse,
  GetLicenseResponse,
  GetStatusDevicesResponse,
  InstallKey,
  MembershipInvitation,
  Namespace,
  PublicKeyResponse,
  ServiceAccount,
  Session,
  SshIdentity,
  Subscription,
  Tag,
  UserAuth,
  Webendpoint,
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

export function mockAnnouncementFull(
  overrides: Partial<Announcement> = {},
): Announcement {
  return {
    uuid: "uuid-0001",
    title: "Welcome to ShellHub",
    content: "Announcement body",
    date: "2024-06-01T10:00:00.000Z",
    ...overrides,
  };
}

export function mockStats(
  overrides: Partial<GetStatusDevicesResponse> = {},
): GetStatusDevicesResponse {
  return {
    registered_devices: 0,
    online_devices: 0,
    pending_devices: 0,
    rejected_devices: 0,
    active_sessions: 0,
    ...overrides,
  };
}

export function mockContainer(overrides: Partial<Device> = {}): Device {
  return {
    uid: "container-uid-1",
    name: "my-container",
    status: "accepted",
    online: true,
    namespace: "my-namespace",
    tenant_id: "tenant-456",
    tags: [],
    last_seen: "2024-01-15T00:00:00Z",
    created_at: "2024-01-01T00:00:00Z",
    identity: { mac: "aa:bb:cc:dd:ee:ff" },
    info: {
      id: "docker",
      pretty_name: "Docker Container",
      arch: "x86_64",
      platform: "docker",
      version: "0.14.0",
    },
    remote_addr: "1.2.3.4",
    ...overrides,
  };
}

export function mockAccessPolicy(
  overrides: Partial<AccessPolicy> = {},
): AccessPolicy {
  return {
    id: "policy-1",
    name: "default-policy",
    subject: { type: "all-members", value: "" },
    filter: { hostname: ".*", tags: [] },
    logins: ["root"],
    source_ip: [],
    action: "allow",
    require_reauth: false,
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
    ...overrides,
  };
}

export function mockWebEndpoint(
  overrides: Partial<Webendpoint> = {},
): Webendpoint {
  return {
    address: "abc123",
    full_address: "abc123.endpoints.shellhub.io",
    namespace: "tenant-456",
    device_uid: "device-uid-1",
    host: "localhost",
    port: 8080,
    ttl: 3600,
    tls: { enabled: false, verify: false, domain: "" },
    expires_in: "2025-01-01T00:00:00Z",
    created_at: "2024-01-01T00:00:00Z",
    ...overrides,
  };
}

export function mockSubscription(
  overrides: Partial<Subscription> = {},
): Subscription {
  return {
    id: "sub-1",
    active: true,
    status: "active",
    end_at: Math.floor(Date.now() / 1000) + 86400 * 30,
    invoices: null,
    ...overrides,
  };
}

export function mockCustomer(overrides: Partial<Customer> = {}): Customer {
  return {
    id: "cus-1",
    name: "Test Customer",
    email: "billing@test.com",
    payment_methods: null,
    ...overrides,
  };
}

export function mockLicense(
  overrides: Partial<GetLicenseResponse> = {},
): GetLicenseResponse {
  return {
    id: "license-1",
    expired: false,
    about_to_expire: false,
    grace_period: false,
    issued_at: 1704067200,
    starts_at: 1704067200,
    expires_at: 1735689600,
    allowed_regions: [],
    customer: { id: "cus-1", name: "Test", email: "test@test.com" },
    features: {
      devices: -1,
      session_recording: true,
      firewall_rules: true,
      reports: true,
      login_link: true,
      billing: true,
    },
    ...overrides,
  };
}

export function mockInvitation(
  overrides: Partial<MembershipInvitation> = {},
): MembershipInvitation {
  return {
    namespace: { tenant_id: "tenant-456", name: "my-namespace" },
    user: { id: "user-123", email: "invited@test.com" },
    invited_by: "owner-123",
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
    expires_at: "2025-01-01T00:00:00Z",
    status: "pending",
    status_updated_at: "2024-01-01T00:00:00Z",
    role: "observer",
    ...overrides,
  };
}

export function mockServiceAccount(
  overrides: Partial<ServiceAccount> = {},
): ServiceAccount {
  return {
    id: "sa-1",
    name: "ci-bot",
    created_at: "2024-01-01T00:00:00Z",
    identities: [],
    ...overrides,
  };
}

export function mockSshIdentity(
  overrides: Partial<SshIdentity> = {},
): SshIdentity {
  return {
    id: "identity-1",
    principal_id: "user-123",
    principal_name: "admin",
    principal_email: "admin@test.com",
    principal_type: "human",
    fingerprint: "SHA256:abc123",
    name: "my-key",
    source: "manual",
    created_at: "2024-01-01T00:00:00Z",
    ...overrides,
  };
}

export function mockInstallKey(
  overrides: Partial<InstallKey> = {},
): InstallKey {
  return {
    id: "key-digest-1",
    tenant_id: "tenant-456",
    created_by: "user-123",
    name: "default-key",
    mode: "manual",
    reusable: true,
    usage_limit: 0,
    used_times: 0,
    ephemeral: false,
    tags: [],
    revoked: false,
    disabled: false,
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
    ...overrides,
  };
}
