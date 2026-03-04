import { describe, it, expect, vi, beforeEach } from "vitest";
import { createAxiosError } from "../../test/createAxiosError";
import { useAuthStore } from "../authStore";

vi.mock("../../api/auth", () => ({
  login: vi.fn(),
  getAuthUser: vi.fn(),
  updateUser: vi.fn(),
  updatePassword: vi.fn(),
}));

import { login as apiLogin, getAuthUser } from "../../api/auth";

const mockedLogin = vi.mocked(apiLogin);
const mockedGetAuthUser = vi.mocked(getAuthUser);

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
  });
  vi.clearAllMocks();
});

describe("authStore", () => {
  describe("login", () => {
    it("sets token and user data on success", async () => {
      mockedLogin.mockResolvedValue({
        token: "jwt-token",
        user: "admin",
        id: "user-123",
        email: "admin@test.com",
        tenant: "tenant-456",
        name: "Admin User",
      });

      await useAuthStore.getState().login("admin", "password");

      const state = useAuthStore.getState();
      expect(state.token).toBe("jwt-token");
      expect(state.user).toBe("admin");
      expect(state.userId).toBe("user-123");
      expect(state.email).toBe("admin@test.com");
      expect(state.tenant).toBe("tenant-456");
      expect(state.loading).toBe(false);
      expect(state.error).toBeNull();
    });

    it("sets error on failure", async () => {
      mockedLogin.mockRejectedValue(new Error("401"));

      await useAuthStore.getState().login("admin", "wrong");

      const state = useAuthStore.getState();
      expect(state.token).toBeNull();
      expect(state.loading).toBe(false);
      expect(state.error).toBe("Invalid username or password");
    });

    it("sets unconfirmed-account message and clears token on 403", async () => {
      mockedLogin.mockRejectedValue(createAxiosError(403));

      // Seed a stale token to verify it is cleared on failure
      useAuthStore.setState({ token: "stale-token" });

      await useAuthStore.getState().login("admin", "password");

      const state = useAuthStore.getState();
      expect(state.token).toBeNull();
      expect(state.loading).toBe(false);
      expect(state.error).toBe(
        "Your account has not been confirmed. Please check your email for the activation link.",
      );
    });

    it("sets loading during request", async () => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      let resolveLogin: (v: any) => void;
      mockedLogin.mockReturnValue(
        new Promise((r) => {
          resolveLogin = r;
        }),
      );

      const promise = useAuthStore.getState().login("admin", "password");
      expect(useAuthStore.getState().loading).toBe(true);

      resolveLogin!({
        token: "t",
        user: "u",
        id: "i",
        email: "e",
        tenant: "t",
        name: "n",
      });
      await promise;

      expect(useAuthStore.getState().loading).toBe(false);
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
        error: "some error",
        username: "admin",
        recoveryEmail: "r@b.com",
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
      });

      // Should NOT persist transient state
      expect(persisted).not.toHaveProperty("loading");
      expect(persisted).not.toHaveProperty("error");
      expect(persisted).not.toHaveProperty("username");
      expect(persisted).not.toHaveProperty("recoveryEmail");
    });
  });
});
