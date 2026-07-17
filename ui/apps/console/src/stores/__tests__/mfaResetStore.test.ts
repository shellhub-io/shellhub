import { describe, it, expect, vi, beforeEach } from "vitest";
import { useMfaResetStore } from "../mfaResetStore";
import { useAuthStore } from "../authStore";
import type { UserAuth } from "@/client";

vi.mock("@/client", () => ({
  requestResetMfa: vi.fn(),
  resetMfa: vi.fn(),
  // Dependencies pulled in transitively by authStore
  login: vi.fn(),
  getUserInfo: vi.fn(),
  updateUser: vi.fn(),
  deleteUser: vi.fn(),
  authMfa: vi.fn(),
  mfaRecover: vi.fn(),
}));

import { requestResetMfa, resetMfa } from "@/client";

const mockedRequestResetMfa = vi.mocked(requestResetMfa);
const mockedResetMfa = vi.mocked(resetMfa);

type SdkResponse<T = unknown> = {
  data: T;
  request: Request;
  response: Response;
};

function mockSdkResponse<T>(data: T): SdkResponse<T> {
  return {
    data,
    request: new Request("http://localhost"),
    response: new Response(),
  };
}

function mockUserAuth(overrides: Partial<UserAuth> = {}): UserAuth {
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
  vi.clearAllMocks();
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
      mockedRequestResetMfa.mockResolvedValueOnce(
        mockSdkResponse({ token: "reset-token" }),
      );

      await useMfaResetStore.getState().requestMfaReset("admin");

      const state = useMfaResetStore.getState();
      expect(state.mfaResetToken).toBe("reset-token");
      expect(state.mfaResetIdentifier).toBe("admin");
      expect(state.loading).toBe(false);
      expect(state.error).toBeNull();
    });

    it("sets loading during request", async () => {
      let resolve!: (v: SdkResponse<{ token: string }>) => void;
      mockedRequestResetMfa.mockReturnValueOnce(
        new Promise<SdkResponse<{ token: string }>>((r) => {
          resolve = r;
        }),
      );

      const promise = useMfaResetStore.getState().requestMfaReset("admin");
      expect(useMfaResetStore.getState().loading).toBe(true);

      resolve(mockSdkResponse({ token: "reset-token" }));
      await promise;

      expect(useMfaResetStore.getState().loading).toBe(false);
    });

    it("sets error and throws on failure", async () => {
      mockedRequestResetMfa.mockRejectedValueOnce(new Error("network error"));

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
      mockedResetMfa.mockResolvedValueOnce(
        mockSdkResponse(mockUserAuth({ token: "reset-token", mfa: false })),
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
      mockedResetMfa.mockResolvedValueOnce(mockSdkResponse(mockUserAuth()));

      await useMfaResetStore.getState().completeMfaReset("AAA11", "BBB22");

      const state = useMfaResetStore.getState();
      expect(state.mfaResetToken).toBeNull();
      expect(state.mfaResetIdentifier).toBeNull();
      expect(state.loading).toBe(false);
    });

    it("calls resetMfa with correct arguments", async () => {
      mockedResetMfa.mockResolvedValueOnce(mockSdkResponse(mockUserAuth()));

      await useMfaResetStore.getState().completeMfaReset("AAA11", "BBB22");

      expect(mockedResetMfa).toHaveBeenCalledWith({
        path: { "user-id": "user-123" },
        body: { main_email_code: "AAA11", recovery_email_code: "BBB22" },
        throwOnError: true,
      });
    });

    it("sets error and throws on failure", async () => {
      mockedResetMfa.mockRejectedValueOnce(new Error("invalid codes"));

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
