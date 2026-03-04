import { describe, it, expect, vi, beforeEach } from "vitest";
import { useAuthStore } from "../authStore";

vi.mock("../../api/client", () => ({
  default: {
    post: vi.fn(),
    get: vi.fn(),
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  },
}));

vi.mock("../../api/auth", () => ({
  getAuthUser: vi.fn(),
  updateUser: vi.fn(),
  updatePassword: vi.fn(),
}));

vi.mock("../../api/mfa", () => ({
  validateMfa: vi.fn(),
  recoverMfa: vi.fn(),
}));

import apiClient from "../../api/client";
import { getAuthUser } from "../../api/auth";
import { validateMfa, recoverMfa } from "../../api/mfa";

const mockedApiClientPost = vi.mocked(apiClient.post);
const mockedGetAuthUser = vi.mocked(getAuthUser);
const mockedValidateMfa = vi.mocked(validateMfa);
const mockedRecoverMfa = vi.mocked(recoverMfa);

beforeEach(() => {
  useAuthStore.setState({
    token: null,
    user: null,
    userId: null,
    email: null,
    username: null,
    recoveryEmail: null,
    tenant: null,
    role: null,
    name: null,
    loading: false,
    error: null,
    mfaEnabled: false,
    mfaToken: null,
    mfaRecoveryExpiry: null,
  });
  vi.clearAllMocks();
});

describe("authStore", () => {
  describe("login", () => {
    it("sets token and user data on success", async () => {
      mockedApiClientPost.mockResolvedValueOnce({
        data: {
          token: "jwt-token",
          user: "admin",
          id: "user-123",
          email: "admin@test.com",
          tenant: "tenant-456",
          name: "Admin User",
        },
        headers: {},
      });

      await useAuthStore.getState().login("admin", "password");

      const state = useAuthStore.getState();
      expect(state.token).toBe("jwt-token");
      expect(state.user).toBe("admin");
      expect(state.userId).toBe("user-123");
      expect(state.email).toBe("admin@test.com");
      expect(state.tenant).toBe("tenant-456");
      expect(state.loading).toBe(false);
    });

    it("re-throws error and resets loading on failure", async () => {
      mockedApiClientPost.mockRejectedValueOnce(new Error("401"));

      await expect(
        useAuthStore.getState().login("admin", "wrong"),
      ).rejects.toThrow("401");

      const state = useAuthStore.getState();
      expect(state.token).toBeNull();
      expect(state.loading).toBe(false);
    });

    it("clears mfaToken at start of login to prevent stale token reuse", async () => {
      // Pre-populate a stale mfaToken from a previous login attempt
      useAuthStore.setState({ mfaToken: "stale-mfa-token" });

      // Capture mfaToken state at the start of the request
      let mfaTokenDuringRequest: string | null = "not-checked";
      mockedApiClientPost.mockImplementationOnce(async () => {
        mfaTokenDuringRequest = useAuthStore.getState().mfaToken;
        return {
          data: {
            token: "jwt-token",
            user: "admin",
            id: "user-123",
            email: "admin@test.com",
            tenant: "tenant-456",
            name: "Admin User",
          },
          headers: {},
        };
      });

      await useAuthStore.getState().login("admin", "password");

      // mfaToken must be null when the request is made (not stale from previous session)
      expect(mfaTokenDuringRequest).toBeNull();
    });

    it("sets loading during request", async () => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      let resolveLogin: (v: any) => void;
      mockedApiClientPost.mockReturnValueOnce(
        new Promise((r) => {
          resolveLogin = r;
        }),
      );

      const promise = useAuthStore.getState().login("admin", "password");
      expect(useAuthStore.getState().loading).toBe(true);

      resolveLogin!({
        data: { token: "t", user: "u", id: "i", email: "e", tenant: "t", name: "n" },
        headers: {},
      });
      await promise;

      expect(useAuthStore.getState().loading).toBe(false);
    });

    it("detects MFA requirement when interceptor sets mfaToken before reject", async () => {
      mockedApiClientPost.mockImplementationOnce(() => {
        useAuthStore.getState().setMfaToken("mfa-temp-token");
        return Promise.reject(new Error("401"));
      });

      await useAuthStore.getState().login("admin", "password");

      const state = useAuthStore.getState();
      expect(state.user).toBe("admin");
      expect(state.mfaEnabled).toBe(true);
      expect(state.error).toBeNull();
      expect(state.loading).toBe(false);
    });
  });

  describe("logout", () => {
    it("resets all state to initial values", () => {
      useAuthStore.setState({
        token: "jwt",
        user: "admin",
        userId: "123",
        email: "a@b.com",
        tenant: "t",
        role: "owner",
        name: "Admin",
      });

      useAuthStore.getState().logout();

      const state = useAuthStore.getState();
      expect(state.token).toBeNull();
      expect(state.user).toBeNull();
      expect(state.userId).toBeNull();
      expect(state.email).toBeNull();
      expect(state.tenant).toBeNull();
      expect(state.role).toBeNull();
      expect(state.name).toBeNull();
    });
  });

  describe("setSession", () => {
    it("updates token, tenant, and role", () => {
      useAuthStore
        .getState()
        .setSession({ token: "new-jwt", tenant: "new-tenant", role: "admin" });

      const state = useAuthStore.getState();
      expect(state.token).toBe("new-jwt");
      expect(state.tenant).toBe("new-tenant");
      expect(state.role).toBe("admin");
    });

    it("preserves existing role when not provided", () => {
      useAuthStore.setState({ role: "owner" });

      useAuthStore
        .getState()
        .setSession({ token: "new-jwt", tenant: "new-tenant" });

      expect(useAuthStore.getState().role).toBe("owner");
    });
  });

  describe("fetchUser", () => {
    it("updates user data from API", async () => {
      mockedGetAuthUser.mockResolvedValue({
        user: "admin",
        id: "user-123",
        email: "admin@test.com",
        recovery_email: "backup@test.com",
        name: "Admin User",
        tenant: "tenant-456",
      });

      await useAuthStore.getState().fetchUser();

      const state = useAuthStore.getState();
      expect(state.username).toBe("admin");
      expect(state.email).toBe("admin@test.com");
      expect(state.recoveryEmail).toBe("backup@test.com");
      expect(state.name).toBe("Admin User");
    });

    it("silently ignores errors (interceptor handles redirect)", async () => {
      mockedGetAuthUser.mockRejectedValue(new Error("401"));

      // Should not throw
      await useAuthStore.getState().fetchUser();
    });
  });

  describe("partialize (persistence)", () => {
    it("only persists the expected fields", () => {
      const store = useAuthStore as unknown as {
        persist: { getOptions: () => { partialize: (s: unknown) => unknown } };
      };
      const partialize = store.persist.getOptions().partialize;

      const full = {
        token: "jwt",
        user: "admin",
        userId: "123",
        email: "a@b.com",
        tenant: "t",
        role: "owner",
        name: "Admin",
        loading: true,
        username: "admin",
        recoveryEmail: "r@b.com",
        // MFA fields
        mfaEnabled: true,
        mfaToken: "mfa-temp-token",
        mfaRecoveryExpiry: "1234567890",
        mfaResetUserId: "some-user-id",
      };

      const persisted = partialize(full) as Record<string, unknown>;

      expect(persisted).toEqual({
        token: "jwt",
        user: "admin",
        userId: "123",
        email: "a@b.com",
        tenant: "t",
        role: "owner",
        name: "Admin",
        mfaEnabled: true, // MFA enabled status SHOULD persist
      });

      // Should NOT persist transient state
      expect(persisted).not.toHaveProperty("loading");
      expect(persisted).not.toHaveProperty("username");
      expect(persisted).not.toHaveProperty("recoveryEmail");

      // Should NOT persist sensitive MFA session/flow state
      expect(persisted).not.toHaveProperty("mfaToken");
      expect(persisted).not.toHaveProperty("mfaRecoveryExpiry");
      expect(persisted).not.toHaveProperty("mfaResetUserId");
    });
  });

  describe("loginWithMfa", () => {
    beforeEach(() => {
      useAuthStore.setState({
        mfaToken: "mfa-temp-token-123",
        user: "admin",
      });
    });

    it("completes MFA login with valid code", async () => {
      mockedValidateMfa.mockResolvedValue({
        token: "jwt-token",
        user: "admin",
        id: "user-123",
        email: "admin@test.com",
        tenant: "tenant-456",
        name: "Admin User",
      });

      await useAuthStore.getState().loginWithMfa("123456");

      const state = useAuthStore.getState();
      expect(state.token).toBe("jwt-token");
      expect(state.mfaToken).toBeNull(); // Temp token cleared
      expect(state.mfaEnabled).toBe(true);
      expect(state.loading).toBe(false);
      expect(mockedValidateMfa).toHaveBeenCalledWith({
        token: "mfa-temp-token-123",
        code: "123456",
      });
    });

    it("throws error when no mfaToken available", async () => {
      useAuthStore.setState({ mfaToken: null });

      await expect(
        useAuthStore.getState().loginWithMfa("123456")
      ).rejects.toThrow("No MFA token available");
    });

    it("sets error on invalid code", async () => {
      mockedValidateMfa.mockRejectedValue(new Error("Invalid code"));

      await expect(
        useAuthStore.getState().loginWithMfa("999999")
      ).rejects.toThrow("Invalid verification code");

      const state = useAuthStore.getState();
      expect(state.loading).toBe(false);
      expect(state.error).toBe("Invalid verification code");
    });
  });

  describe("recoverWithCode", () => {
    beforeEach(() => {
      useAuthStore.setState({
        user: "admin",
      });
    });

    it("authenticates with valid recovery code", async () => {
      const futureExpiry = Math.floor(Date.now() / 1000) + 3600;
      mockedRecoverMfa.mockResolvedValue({
        data: {
          token: "recovered-jwt",
          user: "admin",
          id: "user-123",
          email: "admin@test.com",
          tenant: "tenant-456",
          name: "Admin User",
        },
        expiresAt: futureExpiry.toString(),
      });

      await useAuthStore.getState().recoverWithCode("recovery-code-abc");

      const state = useAuthStore.getState();
      expect(state.token).toBe("recovered-jwt");
      expect(state.mfaRecoveryExpiry).toBe(futureExpiry);
      expect(state.loading).toBe(false);
      expect(mockedRecoverMfa).toHaveBeenCalledWith({
        identifier: "admin",
        recovery_code: "recovery-code-abc",
      });
    });

    it("clears mfaToken on successful recovery to prevent stale token re-use", async () => {
      const futureExpiry = Math.floor(Date.now() / 1000) + 3600;
      useAuthStore.setState({ mfaToken: "mfa-temp-token", user: "admin" });
      mockedRecoverMfa.mockResolvedValue({
        data: {
          token: "recovered-jwt",
          user: "admin",
          id: "user-123",
          email: "admin@test.com",
          tenant: "tenant-456",
          name: "Admin User",
        },
        expiresAt: futureExpiry.toString(),
      });

      await useAuthStore.getState().recoverWithCode("recovery-code-abc");

      expect(useAuthStore.getState().mfaToken).toBeNull();
    });

    it("throws error when no username available", async () => {
      useAuthStore.setState({ user: null, username: null });

      await expect(
        useAuthStore.getState().recoverWithCode("recovery-code")
      ).rejects.toThrow("Username or email is required");
    });

    it("sets error on invalid recovery code", async () => {
      mockedRecoverMfa.mockRejectedValue(new Error("Invalid"));

      await expect(
        useAuthStore.getState().recoverWithCode("invalid-code")
      ).rejects.toThrow("Invalid recovery code or username");

      const state = useAuthStore.getState();
      expect(state.error).toBe("Invalid recovery code or username");
    });
  });

  describe("updateMfaStatus", () => {
    it("updates MFA enabled status", () => {
      useAuthStore.getState().updateMfaStatus(true);

      expect(useAuthStore.getState().mfaEnabled).toBe(true);
    });

    it("can disable MFA status", () => {
      useAuthStore.setState({ mfaEnabled: true });

      useAuthStore.getState().updateMfaStatus(false);

      expect(useAuthStore.getState().mfaEnabled).toBe(false);
    });
  });

  describe("setMfaToken", () => {
    it("sets MFA token", () => {
      useAuthStore.getState().setMfaToken("mfa-token-123");

      expect(useAuthStore.getState().mfaToken).toBe("mfa-token-123");
    });
  });

  describe("login with mfa field", () => {
    it("sets mfaEnabled true when server reports mfa: true", async () => {
      mockedApiClientPost.mockResolvedValueOnce({
        data: {
          token: "jwt-token",
          user: "admin",
          id: "user-123",
          email: "admin@test.com",
          tenant: "tenant-456",
          name: "Admin User",
          mfa: true,
        },
        headers: {},
      });

      await useAuthStore.getState().login("admin", "password");

      expect(useAuthStore.getState().mfaEnabled).toBe(true);
    });

    it("sets mfaEnabled false when mfa field is absent", async () => {
      mockedApiClientPost.mockResolvedValueOnce({
        data: {
          token: "jwt-token",
          user: "admin",
          id: "user-123",
          email: "admin@test.com",
          tenant: "tenant-456",
          name: "Admin User",
        },
        headers: {},
      });

      await useAuthStore.getState().login("admin", "password");

      expect(useAuthStore.getState().mfaEnabled).toBe(false);
    });
  });
});
