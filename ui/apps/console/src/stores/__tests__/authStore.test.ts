import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { useAuthStore } from "../authStore";
import { mockUserAuth } from "@/tests/factories";
import { VALID_JWT } from "@/tests/seedAuthStore";

beforeEach(() => {
  useAuthStore.setState({
    token: null,
    user: null,
    userId: null,
    email: null,
    username: null,
    origin: null,
    recoveryEmail: null,
    tenant: null,
    role: null,
    name: null,
    loading: false,
    error: null,
    mfaEnabled: false,
    mfaToken: null,
    mfaRecoveryExpiry: null,
    isAdmin: false,
  });
});

describe("authStore", () => {
  describe("login", () => {
    it("sets token and user data on success", async () => {
      server.use(
        http.post("*/api/login", () => HttpResponse.json(mockUserAuth())),
      );

      await useAuthStore.getState().login("admin", "password");

      const state = useAuthStore.getState();
      expect(state.token).toBe(VALID_JWT);
      expect(state.user).toBe("admin");
      expect(state.userId).toBe("user-123");
      expect(state.email).toBe("admin@test.com");
      expect(state.tenant).toBe("tenant-456");
      expect(state.loading).toBe(false);
    });

    it("re-throws error and resets loading on failure", async () => {
      server.use(
        http.post("*/api/login", () => HttpResponse.json({}, { status: 401 })),
      );

      await expect(
        useAuthStore.getState().login("admin", "wrong"),
      ).rejects.toThrow();

      const state = useAuthStore.getState();
      expect(state.token).toBeNull();
      expect(state.loading).toBe(false);
    });

    it("re-throws on 403 (Login page handles redirect)", async () => {
      server.use(
        http.post("*/api/login", () => HttpResponse.json({}, { status: 403 })),
      );

      await expect(
        useAuthStore.getState().login("admin", "password"),
      ).rejects.toThrow();

      const state = useAuthStore.getState();
      expect(state.token).toBeNull();
      expect(state.loading).toBe(false);
    });

    it("clears mfaToken at start of login to prevent stale token reuse", async () => {
      useAuthStore.setState({ mfaToken: "stale-mfa-token" });

      let mfaTokenDuringRequest: string | null = "not-checked";
      server.use(
        http.post("*/api/login", () => {
          mfaTokenDuringRequest = useAuthStore.getState().mfaToken;
          return HttpResponse.json(mockUserAuth());
        }),
      );

      await useAuthStore.getState().login("admin", "password");

      expect(mfaTokenDuringRequest).toBeNull();
    });

    it("sets loading during request", async () => {
      let resolveHandler!: (r: Response) => void;
      const handlerReady = new Promise<void>((ready) => {
        server.use(
          http.post(
            "*/api/login",
            () =>
              new Promise<Response>((resolve) => {
                resolveHandler = resolve;
                ready();
              }),
          ),
        );
      });

      const promise = useAuthStore.getState().login("admin", "password");
      await handlerReady;
      expect(useAuthStore.getState().loading).toBe(true);

      resolveHandler(HttpResponse.json(mockUserAuth()));
      await promise;

      expect(useAuthStore.getState().loading).toBe(false);
    });

    it("detects MFA requirement when interceptor sets mfaToken before reject", async () => {
      server.use(
        http.post("*/api/login", () =>
          HttpResponse.json(
            {},
            {
              status: 401,
              headers: { "x-mfa-token": "mfa-temp-token" },
            },
          ),
        ),
      );

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

    describe("Chatwoot cleanup", () => {
      afterEach(() => {
        delete window.$chatwoot;
      });

      it("calls toggle('close') then reset() when window.$chatwoot is set", () => {
        const toggle = vi.fn();
        const reset = vi.fn();
        window.$chatwoot = {
          toggle,
          reset,
          setUser: vi.fn(),
          setCustomAttributes: vi.fn(),
          setConversationCustomAttributes: vi.fn(),
          deleteCustomAttribute: vi.fn(),
        };

        useAuthStore.getState().logout();

        expect(toggle).toHaveBeenCalledWith("close");
        expect(reset).toHaveBeenCalled();
        const toggleOrder = toggle.mock.invocationCallOrder[0];
        const resetOrder = reset.mock.invocationCallOrder[0];
        expect(toggleOrder).toBeLessThan(resetOrder);
      });

      it("does not throw when window.$chatwoot is undefined", () => {
        delete window.$chatwoot;
        expect(() => useAuthStore.getState().logout()).not.toThrow();
      });

      it("still clears auth state even when window.$chatwoot.reset throws", () => {
        window.$chatwoot = {
          toggle: vi.fn(),
          reset: vi.fn(() => {
            throw new Error("widget error");
          }),
          setUser: vi.fn(),
          setCustomAttributes: vi.fn(),
          setConversationCustomAttributes: vi.fn(),
          deleteCustomAttribute: vi.fn(),
        };

        useAuthStore.setState({ token: "jwt", user: "admin" });

        expect(() => useAuthStore.getState().logout()).not.toThrow();

        expect(useAuthStore.getState().token).toBeNull();
        expect(useAuthStore.getState().user).toBeNull();
      });
    });
  });

  describe("setSession", () => {
    it("updates token, tenant, and role", () => {
      useAuthStore.getState().setSession({
        token: "new-jwt",
        tenant: "new-tenant",
        role: "administrator",
      });

      const state = useAuthStore.getState();
      expect(state.token).toBe("new-jwt");
      expect(state.tenant).toBe("new-tenant");
      expect(state.role).toBe("administrator");
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
      server.use(
        http.get("*/api/auth/user", () =>
          HttpResponse.json(
            mockUserAuth({
              user: "admin",
              email: "admin@test.com",
              recovery_email: "backup@test.com",
              name: "Admin User",
            }),
          ),
        ),
      );

      await useAuthStore.getState().fetchUser();

      const state = useAuthStore.getState();
      expect(state.username).toBe("admin");
      expect(state.email).toBe("admin@test.com");
      expect(state.recoveryEmail).toBe("backup@test.com");
      expect(state.name).toBe("Admin User");
    });

    it("maps the SSO origin into the store", async () => {
      server.use(
        http.get("*/api/auth/user", () =>
          HttpResponse.json(mockUserAuth({ origin: "saml" })),
        ),
      );

      await useAuthStore.getState().fetchUser();

      expect(useAuthStore.getState().origin).toBe("saml");
    });

    it("silently ignores errors (interceptor handles redirect)", async () => {
      server.use(
        http.get("*/api/auth/user", () =>
          HttpResponse.json({}, { status: 401 }),
        ),
      );

      await useAuthStore.getState().fetchUser();
    });
  });

  describe("loginWithToken", () => {
    it("maps the SSO origin into the store", async () => {
      server.use(
        http.get("*/api/auth/user", () =>
          HttpResponse.json(mockUserAuth({ origin: "saml" })),
        ),
      );

      await useAuthStore.getState().loginWithToken(VALID_JWT);

      expect(useAuthStore.getState().origin).toBe("saml");
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
        origin: "saml",
        tenant: "t",
        role: "owner",
        name: "Admin",
        loading: true,
        username: "admin",
        recoveryEmail: "r@b.com",
        mfaEnabled: true,
        mfaToken: "mfa-temp-token",
        mfaRecoveryExpiry: "1234567890",
      };

      const persisted = partialize(full) as Record<string, unknown>;

      expect(persisted).toEqual({
        token: "jwt",
        user: "admin",
        userId: "123",
        email: "a@b.com",
        origin: "saml",
        tenant: "t",
        role: "owner",
        name: "Admin",
        mfaEnabled: true,
      });

      expect(persisted).not.toHaveProperty("loading");
      expect(persisted).not.toHaveProperty("username");
      expect(persisted).not.toHaveProperty("recoveryEmail");

      expect(persisted).not.toHaveProperty("mfaToken");
      expect(persisted).not.toHaveProperty("mfaRecoveryExpiry");
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
      server.use(
        http.post("*/api/user/mfa/auth", () =>
          HttpResponse.json(mockUserAuth()),
        ),
      );

      await useAuthStore.getState().loginWithMfa("123456");

      const state = useAuthStore.getState();
      expect(state.token).toBe(VALID_JWT);
      expect(state.mfaToken).toBeNull();
      expect(state.mfaEnabled).toBe(true);
      expect(state.loading).toBe(false);
    });

    it("throws error when no mfaToken available", async () => {
      useAuthStore.setState({ mfaToken: null });

      await expect(
        useAuthStore.getState().loginWithMfa("123456"),
      ).rejects.toThrow("No MFA token available");
    });

    it("sets error on invalid code", async () => {
      server.use(
        http.post("*/api/user/mfa/auth", () =>
          HttpResponse.json({}, { status: 401 }),
        ),
      );

      await expect(
        useAuthStore.getState().loginWithMfa("999999"),
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
      server.use(
        http.post("*/api/user/mfa/recover", () =>
          HttpResponse.json(mockUserAuth({ token: "recovered-jwt" }), {
            headers: { "x-expires-at": futureExpiry.toString() },
          }),
        ),
      );

      await useAuthStore.getState().recoverWithCode("recovery-code-abc");

      const state = useAuthStore.getState();
      expect(state.token).toBe("recovered-jwt");
      expect(state.mfaRecoveryExpiry).toBe(futureExpiry);
      expect(state.loading).toBe(false);
    });

    it("clears mfaToken on successful recovery to prevent stale token re-use", async () => {
      const futureExpiry = Math.floor(Date.now() / 1000) + 3600;
      useAuthStore.setState({ mfaToken: "mfa-temp-token", user: "admin" });
      server.use(
        http.post("*/api/user/mfa/recover", () =>
          HttpResponse.json(mockUserAuth({ token: "recovered-jwt" }), {
            headers: { "x-expires-at": futureExpiry.toString() },
          }),
        ),
      );

      await useAuthStore.getState().recoverWithCode("recovery-code-abc");

      expect(useAuthStore.getState().mfaToken).toBeNull();
    });

    it("throws error when no username available", async () => {
      useAuthStore.setState({ user: null, username: null });

      await expect(
        useAuthStore.getState().recoverWithCode("recovery-code"),
      ).rejects.toThrow("Username or email is required");
    });

    it("sets error on invalid recovery code", async () => {
      server.use(
        http.post("*/api/user/mfa/recover", () =>
          HttpResponse.json({}, { status: 400 }),
        ),
      );

      await expect(
        useAuthStore.getState().recoverWithCode("invalid-code"),
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
      server.use(
        http.post("*/api/login", () =>
          HttpResponse.json(mockUserAuth({ mfa: true })),
        ),
      );

      await useAuthStore.getState().login("admin", "password");

      expect(useAuthStore.getState().mfaEnabled).toBe(true);
    });

    it("sets mfaEnabled false when mfa is false", async () => {
      server.use(
        http.post("*/api/login", () => HttpResponse.json(mockUserAuth())),
      );

      await useAuthStore.getState().login("admin", "password");

      expect(useAuthStore.getState().mfaEnabled).toBe(false);
    });
  });
});
