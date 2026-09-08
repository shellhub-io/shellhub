import { describe, it, expect, beforeEach } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { useMfaResetStore } from "../mfaResetStore";
import { useAuthStore } from "../authStore";
import { mockUserAuth } from "@/tests/factories";

beforeEach(() => {
  useMfaResetStore.setState({
    mfaResetToken: null,
    mfaResetIdentifier: null,
    loading: false,
    error: null,
  });
  useAuthStore.setState({
    token: null,
    user: null,
    userId: null,
    email: null,
    tenant: null,
    name: null,
    isAdmin: false,
    mfaEnabled: false,
    loading: false,
    error: null,
    mfaToken: null,
    mfaRecoveryExpiry: null,
    username: null,
    recoveryEmail: null,
    role: null,
  });
});

describe("mfaResetStore", () => {
  describe("initial state", () => {
    it("initializes with clean state", () => {
      const state = useMfaResetStore.getState();
      expect(state.mfaResetToken).toBeNull();
      expect(state.mfaResetIdentifier).toBeNull();
      expect(state.loading).toBe(false);
      expect(state.error).toBeNull();
    });

    it("does not use Zustand persist", () => {
      expect(
        (useMfaResetStore as unknown as Record<string, unknown>).persist,
      ).toBeUndefined();
    });
  });

  describe("requestMfaReset", () => {
    it("stores the opaque token from the API response", async () => {
      server.use(
        http.post("*/api/user/mfa/reset", () =>
          HttpResponse.json({ token: "reset-token" }),
        ),
      );

      await useMfaResetStore.getState().requestMfaReset("admin");

      const state = useMfaResetStore.getState();
      expect(state.mfaResetToken).toBe("reset-token");
      expect(state.mfaResetIdentifier).toBe("admin");
      expect(state.loading).toBe(false);
      expect(state.error).toBeNull();
    });

    it("sets loading during request", async () => {
      let resolveHandler!: (r: Response) => void;
      const handlerReady = new Promise<void>((ready) => {
        server.use(
          http.post(
            "*/api/user/mfa/reset",
            () =>
              new Promise<Response>((resolve) => {
                resolveHandler = resolve;
                ready();
              }),
          ),
        );
      });

      const promise = useMfaResetStore.getState().requestMfaReset("admin");
      await handlerReady;
      expect(useMfaResetStore.getState().loading).toBe(true);

      resolveHandler(HttpResponse.json({ token: "reset-token" }));
      await promise;

      expect(useMfaResetStore.getState().loading).toBe(false);
    });

    it("sets error and throws on failure", async () => {
      server.use(
        http.post("*/api/user/mfa/reset", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );

      await expect(
        useMfaResetStore.getState().requestMfaReset("admin"),
      ).rejects.toThrow("Reset request failed");

      const state = useMfaResetStore.getState();
      expect(state.loading).toBe(false);
      expect(state.error).toBe(
        "Unable to send reset emails. Please check your identifier.",
      );
      expect(state.mfaResetToken).toBeNull();
    });
  });

  describe("completeMfaReset", () => {
    beforeEach(() => {
      useMfaResetStore.setState({ mfaResetToken: "user-123" });
    });

    it("throws when no mfaResetToken", async () => {
      useMfaResetStore.setState({ mfaResetToken: null });

      await expect(
        useMfaResetStore.getState().completeMfaReset("code1", "code2"),
      ).rejects.toThrow("No reset token available");

      expect(useMfaResetStore.getState().error).toBe(
        "Invalid reset session. Please start over.",
      );
    });

    it("sets auth state in authStore on success", async () => {
      server.use(
        http.put("*/api/user/mfa/reset/:userId", () =>
          HttpResponse.json(mockUserAuth({ token: "reset-token", mfa: false })),
        ),
      );

      await useMfaResetStore.getState().completeMfaReset("AAA11", "BBB22");

      const auth = useAuthStore.getState();
      expect(auth.token).toBe("reset-token");
      expect(auth.user).toBe("admin");
      expect(auth.userId).toBe("user-123");
      expect(auth.email).toBe("admin@test.com");
      expect(auth.tenant).toBe("tenant-456");
      expect(auth.isAdmin).toBe(false);
      expect(auth.mfaEnabled).toBe(false);
    });

    it("clears mfaResetToken and mfaResetIdentifier on success", async () => {
      useMfaResetStore.setState({
        mfaResetToken: "user-123",
        mfaResetIdentifier: "admin",
      });
      server.use(
        http.put("*/api/user/mfa/reset/:userId", () =>
          HttpResponse.json(mockUserAuth()),
        ),
      );

      await useMfaResetStore.getState().completeMfaReset("AAA11", "BBB22");

      const state = useMfaResetStore.getState();
      expect(state.mfaResetToken).toBeNull();
      expect(state.mfaResetIdentifier).toBeNull();
      expect(state.loading).toBe(false);
    });

    it("sets error and throws on failure", async () => {
      server.use(
        http.put("*/api/user/mfa/reset/:userId", () =>
          HttpResponse.json({}, { status: 400 }),
        ),
      );

      await expect(
        useMfaResetStore.getState().completeMfaReset("WRONG", "CODES"),
      ).rejects.toThrow("Invalid codes");

      const state = useMfaResetStore.getState();
      expect(state.loading).toBe(false);
      expect(state.error).toBe(
        "Invalid verification codes. Please check and try again.",
      );
      expect(state.mfaResetToken).toBe("user-123");
    });
  });

  describe("reset", () => {
    it("clears all state to initial values", () => {
      useMfaResetStore.setState({
        mfaResetToken: "user-abc",
        mfaResetIdentifier: "admin",
        loading: true,
        error: "some error",
      });

      useMfaResetStore.getState().reset();

      const state = useMfaResetStore.getState();
      expect(state.mfaResetToken).toBeNull();
      expect(state.mfaResetIdentifier).toBeNull();
      expect(state.loading).toBe(false);
      expect(state.error).toBeNull();
    });
  });
});
