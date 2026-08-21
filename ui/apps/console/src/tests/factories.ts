import type { Namespace, Tag, UserAuth } from "@/client";

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
