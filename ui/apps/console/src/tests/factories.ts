import type {
  AccessPolicy,
  Device,
  FirewallRulesResponse,
  GetLicenseResponse,
  GetStatusDevicesResponse,
  ProvisioningKey,
  Namespace,
  PublicKeyResponse,
  Session,
  Tag,
  UserAuth,
  Webendpoint,
} from "@/client";

/**
 * Builds a signed-in user for a test. Every field has a value, so a case names only what it is about
 * and the rest stays out of the way.
 */
export function mockUserAuth(overrides: Partial<UserAuth> = {}): UserAuth {
  return {
    token:
      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjk5OTk5OTk5OTksInN1YiI6InRlc3QifQ.fake-sig",
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
    auth_methods: ["local"],
    ...overrides,
  };
}

/**
 * Builds a namespace for a test. Every field has a value, so a case names only what it is about
 * and the rest stays out of the way.
 */
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
    devices_removed_count: 0,
    ...overrides,
  };
}

/**
 * Builds a tag for a test. Every field has a value, so a case names only what it is about
 * and the rest stays out of the way.
 */
export function mockTag(overrides: Partial<Tag> = {}): Tag {
  return {
    name: "tag",
    tenant_id: "tenant-456",
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
    ...overrides,
  };
}

/**
 * Builds an accepted device for a test. Every field has a value, so a case names only what it is about
 * and the rest stays out of the way.
 */
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

/**
 * Builds a closed session, with a device attached for a test. Every field has a value, so a case names only what it is about
 * and the rest stays out of the way.
 */
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
    web: false,
    position: { latitude: 0, longitude: 0 },
    events: { types: ["pty-req", "shell"], seats: [0], first: "pty-req" },
    ...overrides,
  };
}

/**
 * Builds a firewall rule for a test. Every field has a value, so a case names only what it is about
 * and the rest stays out of the way.
 */
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

/**
 * Builds a public key for a test. Every field has a value, so a case names only what it is about
 * and the rest stays out of the way.
 */
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

/**
 * Builds the device status counts for a test. Every field has a value, so a case names only what it is about
 * and the rest stays out of the way.
 */
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

/**
 * Builds a container, which is a device with the container platform for a test. Every field has a value, so a case names only what it is about
 * and the rest stays out of the way.
 */
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

/**
 * Builds an access policy for a test. Every field has a value, so a case names only what it is about
 * and the rest stays out of the way.
 */
export function mockAccessPolicy(
  overrides: Partial<AccessPolicy> = {},
): AccessPolicy {
  return {
    id: "policy-1",
    name: "default-policy",
    subject: { type: "all-members", value: "" },
    filter: { tags: [] },
    logins: ["root"],
    source_ip: [],
    action: "allow",
    require_reauth: false,
    subject_matches: true,
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
    ...overrides,
  };
}

/**
 * Builds a web endpoint for a test. Every field has a value, so a case names only what it is about
 * and the rest stays out of the way.
 */
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

/**
 * Builds an installed licence, unexpired for a test. Every field has a value, so a case names only what it is about
 * and the rest stays out of the way.
 */
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

/**
 * Builds a provisioning key with an allowance left for a test. Every field has a value, so a case names only what it is about
 * and the rest stays out of the way.
 */
export function mockProvisioningKey(
  overrides: Partial<ProvisioningKey> = {},
): ProvisioningKey {
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
