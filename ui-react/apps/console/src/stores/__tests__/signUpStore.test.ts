import { describe, it, expect, vi, beforeEach } from "vitest";
import { useSignUpStore } from "../signUpStore";

vi.mock("../../client", () => ({
  registerUser: vi.fn(),
  resendEmail: vi.fn(),
  getValidateAccount: vi.fn(),
}));

import {
  registerUser as apiRegisterUser,
  resendEmail as apiResendEmail,
  getValidateAccount as apiGetValidateAccount,
} from "../../client";

const mockedRegisterUser = vi.mocked(apiRegisterUser);
const mockedResendEmail = vi.mocked(apiResendEmail);
const mockedGetValidateAccount = vi.mocked(apiGetValidateAccount);

type SdkResponse<T = unknown> = { data: T; request: Request; response: Response };

function mockSdkResponse<T>(data: T): SdkResponse<T> {
  return {
    data,
    request: new Request("http://localhost"),
    response: new Response(),
  };
}

beforeEach(() => {
  useSignUpStore.setState({
    signUpToken: null,
    signUpTenant: null,
    signUpLoading: false,
    signUpError: null,
    signUpServerFields: [],
    resendLoading: false,
    resendError: null,
    validationStatus: "idle",
  });
  vi.clearAllMocks();
});

/** Creates a mock SDK error with status and optional body. */
function createSdkError(status: number, body?: unknown) {
  return Object.assign(new Error(String(status)), { status, body: body ?? {} });
}

describe("signUpStore", () => {
  describe("signUp", () => {
    vi.spyOn(console, "warn").mockImplementation(() => {}); // Suppress expected warn logs during tests

    it("sets loading during request", async () => {
      let resolve: (v: SdkResponse<{ token: string; tenant: string }>) => void;
      mockedRegisterUser.mockReturnValue(new Promise<SdkResponse<{ token: string; tenant: string }>>((r) => { resolve = r; }));

      const promise = useSignUpStore.getState().signUp({
        name: "Test", email: "t@t.com", username: "test", password: "pass1", email_marketing: false,
      });

      expect(useSignUpStore.getState().signUpLoading).toBe(true);

      resolve!(mockSdkResponse({ token: "tok", tenant: "ten" }));
      await promise;

      expect(useSignUpStore.getState().signUpLoading).toBe(false);
    });

    it("stores token and tenant on success and returns token", async () => {
      mockedRegisterUser.mockResolvedValue(mockSdkResponse({ token: "jwt-token", tenant: "tenant-abc" }));

      const result = await useSignUpStore.getState().signUp({
        name: "Test", email: "t@t.com", username: "test", password: "pass1", email_marketing: false,
      });

      expect(result).toBe("jwt-token");
      expect(useSignUpStore.getState().signUpToken).toBe("jwt-token");
      expect(useSignUpStore.getState().signUpTenant).toBe("tenant-abc");
    });

    it("returns null and stores null token when response has no token (normal flow)", async () => {
      mockedRegisterUser.mockResolvedValue(mockSdkResponse({}));

      const result = await useSignUpStore.getState().signUp({
        name: "Test", email: "t@t.com", username: "test", password: "pass1", email_marketing: false,
      });

      expect(result).toBeNull();
      expect(useSignUpStore.getState().signUpToken).toBeNull();
      expect(useSignUpStore.getState().signUpTenant).toBeNull();
    });

    it("sets signUpServerFields on 400/409 with field array and returns null", async () => {
      mockedRegisterUser.mockRejectedValue(createSdkError(400, ["username", "email"]));

      const result = await useSignUpStore.getState().signUp({
        name: "Test", email: "t@t.com", username: "test", password: "pass1", email_marketing: false,
      });

      expect(result).toBeNull();
      expect(useSignUpStore.getState().signUpLoading).toBe(false);
      expect(useSignUpStore.getState().signUpServerFields).toEqual(["username", "email"]);
      expect(useSignUpStore.getState().signUpError).toBeNull();
    });

    it("sets signUpServerFields on 409 with field array and returns null", async () => {
      mockedRegisterUser.mockRejectedValue(createSdkError(409, ["username"]));

      const result = await useSignUpStore.getState().signUp({
        name: "Test", email: "t@t.com", username: "test", password: "pass1", email_marketing: false,
      });

      expect(result).toBeNull();
      expect(useSignUpStore.getState().signUpServerFields).toEqual(["username"]);
      expect(useSignUpStore.getState().signUpError).toBeNull();
    });

    it("falls through to generic error when 400 body is not an array", async () => {
      mockedRegisterUser.mockRejectedValue(createSdkError(400, { message: "validation error" }));

      const result = await useSignUpStore.getState().signUp({
        name: "Test", email: "t@t.com", username: "test", password: "pass1", email_marketing: false,
      });

      expect(result).toBeNull();
      expect(useSignUpStore.getState().signUpServerFields).toEqual([]);
      expect(useSignUpStore.getState().signUpError).toBe("An error occurred. Please try again.");
    });

    it("sets signUpError on non-field errors and returns null", async () => {
      mockedRegisterUser.mockRejectedValue(new Error("network error"));

      const result = await useSignUpStore.getState().signUp({
        name: "Test", email: "t@t.com", username: "test", password: "pass1", email_marketing: false,
      });

      expect(result).toBeNull();
      expect(useSignUpStore.getState().signUpLoading).toBe(false);
      expect(useSignUpStore.getState().signUpError).toBe("An error occurred. Please try again.");
      expect(useSignUpStore.getState().signUpServerFields).toEqual([]);
    });

    it("clears stale token and tenant at the start of a new attempt", async () => {
      useSignUpStore.setState({ signUpToken: "old-token", signUpTenant: "old-tenant" });
      mockedRegisterUser.mockRejectedValue(new Error("network error"));

      await useSignUpStore.getState().signUp({
        name: "Test", email: "t@t.com", username: "test", password: "pass1", email_marketing: false,
      });

      expect(useSignUpStore.getState().signUpToken).toBeNull();
      expect(useSignUpStore.getState().signUpTenant).toBeNull();
    });
  });

  describe("clearSignUpServerField", () => {
    it("removes a specific field from signUpServerFields", () => {
      useSignUpStore.setState({ signUpServerFields: ["username", "email"] });

      useSignUpStore.getState().clearSignUpServerField("username");

      expect(useSignUpStore.getState().signUpServerFields).toEqual(["email"]);
    });
  });

  describe("resetSignUpErrors", () => {
    it("clears signUpError and signUpServerFields", () => {
      useSignUpStore.setState({
        signUpError: "some error",
        signUpServerFields: ["email"],
      });

      useSignUpStore.getState().resetSignUpErrors();

      expect(useSignUpStore.getState().signUpError).toBeNull();
      expect(useSignUpStore.getState().signUpServerFields).toEqual([]);
    });
  });

  describe("resendEmail", () => {
    it("returns true on success", async () => {
      mockedResendEmail.mockResolvedValue(mockSdkResponse(undefined));

      const result = await useSignUpStore.getState().resendEmail("testuser");

      expect(result).toBe(true);
      expect(useSignUpStore.getState().resendLoading).toBe(false);
      expect(useSignUpStore.getState().resendError).toBeNull();
    });

    it("returns false and sets resendError on failure", async () => {
      mockedResendEmail.mockRejectedValue(new Error("server error"));

      const result = await useSignUpStore.getState().resendEmail("testuser");

      expect(result).toBe(false);
      expect(useSignUpStore.getState().resendLoading).toBe(false);
      expect(useSignUpStore.getState().resendError).toBe("Failed to resend email. Please try again.");
    });

    it("sets loading during request", async () => {
      let resolve: (v: SdkResponse) => void;
      mockedResendEmail.mockReturnValue(new Promise<SdkResponse>((r) => { resolve = r; }));

      const promise = useSignUpStore.getState().resendEmail("testuser");
      expect(useSignUpStore.getState().resendLoading).toBe(true);

      resolve!(mockSdkResponse(undefined));
      await promise;

      expect(useSignUpStore.getState().resendLoading).toBe(false);
    });
  });

  describe("validateAccount", () => {
    it("transitions to success on 200", async () => {
      mockedGetValidateAccount.mockResolvedValue(mockSdkResponse(undefined));

      await useSignUpStore.getState().validateAccount("t@t.com", "valid-token");

      expect(useSignUpStore.getState().validationStatus).toBe("success");
    });

    it("transitions to failed-token on 400 (expired token)", async () => {
      mockedGetValidateAccount.mockRejectedValue(createSdkError(400));

      await useSignUpStore.getState().validateAccount("t@t.com", "expired-token");

      expect(useSignUpStore.getState().validationStatus).toBe("failed-token");
    });

    it("transitions to failed-token on 401 (wrong token)", async () => {
      mockedGetValidateAccount.mockRejectedValue(createSdkError(401));

      await useSignUpStore.getState().validateAccount("t@t.com", "wrong-token");

      expect(useSignUpStore.getState().validationStatus).toBe("failed-token");
    });

    it("transitions to failed on 404 (user not found)", async () => {
      mockedGetValidateAccount.mockRejectedValue(createSdkError(404));

      await useSignUpStore.getState().validateAccount("t@t.com", "unknown-user");

      expect(useSignUpStore.getState().validationStatus).toBe("failed");
    });

    it("transitions to failed on other errors", async () => {
      mockedGetValidateAccount.mockRejectedValue(createSdkError(500));

      await useSignUpStore.getState().validateAccount("t@t.com", "bad-token");

      expect(useSignUpStore.getState().validationStatus).toBe("failed");
    });

    it("does not update state when the request is aborted", async () => {
      mockedGetValidateAccount.mockRejectedValue(new Error("aborted"));

      const controller = new AbortController();
      controller.abort();

      await useSignUpStore.getState().validateAccount("t@t.com", "tok", controller.signal);

      // Status must stay at "processing" — no terminal state set after an abort.
      expect(useSignUpStore.getState().validationStatus).toBe("processing");
    });

    it("sets processing during request", async () => {
      let resolve: (v: SdkResponse) => void;
      mockedGetValidateAccount.mockReturnValue(new Promise<SdkResponse>((r) => { resolve = r; }));

      const promise = useSignUpStore.getState().validateAccount("t@t.com", "tok");
      expect(useSignUpStore.getState().validationStatus).toBe("processing");

      resolve!(mockSdkResponse(undefined));
      await promise;
    });
  });

  describe("resetValidation", () => {
    it("resets validationStatus to idle", () => {
      useSignUpStore.setState({ validationStatus: "success" });

      useSignUpStore.getState().resetValidation();

      expect(useSignUpStore.getState().validationStatus).toBe("idle");
    });
  });

  describe("setValidationFailed", () => {
    it("sets validationStatus to failed", () => {
      useSignUpStore.setState({ validationStatus: "processing" });

      useSignUpStore.getState().setValidationFailed();

      expect(useSignUpStore.getState().validationStatus).toBe("failed");
    });
  });

  describe("resetResendError", () => {
    it("clears resendError", () => {
      useSignUpStore.setState({ resendError: "some error" });

      useSignUpStore.getState().resetResendError();

      expect(useSignUpStore.getState().resendError).toBeNull();
    });
  });
});
