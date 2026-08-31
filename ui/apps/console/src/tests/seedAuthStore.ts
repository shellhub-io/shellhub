import { useAuthStore } from "@/stores/authStore";
import { mockUserAuth } from "./factories";

type AuthData = Pick<
  ReturnType<typeof useAuthStore.getState>,
  | "name"
  | "user"
  | "username"
  | "email"
  | "recoveryEmail"
  | "mfaEnabled"
  | "origin"
  | "loading"
  | "token"
  | "userId"
  | "tenant"
  | "role"
  | "isAdmin"
>;

/**
 * Puts a signed-in user in the auth store, so a test can start from an authenticated state
 * without going through sign-in.
 */
export function seedAuthStore(overrides: Partial<AuthData> = {}) {
  const auth = mockUserAuth();
  useAuthStore.setState({
    name: auth.name,
    user: auth.user,
    username: auth.user,
    email: auth.email,
    recoveryEmail: auth.recovery_email,
    mfaEnabled: auth.mfa,
    origin: auth.origin,
    loading: false,
    token: auth.token,
    userId: auth.id,
    tenant: auth.tenant,
    role: auth.role,
    isAdmin: auth.admin,
    ...overrides,
  });
}
