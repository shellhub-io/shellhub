import type { UserAuth } from "@/client";

export function mockUserAuth(
  overrides: Partial<UserAuth> = {},
): UserAuth {
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
