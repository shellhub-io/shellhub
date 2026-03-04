import { describe, it, expect, vi, beforeEach } from "vitest";
import { CanceledError } from "axios";
import { createAxiosError } from "../../test/createAxiosError";
import { useSignUpStore } from "../signUpStore";

vi.mock("../../api/users", () => ({
  signUp: vi.fn(),
  resendEmail: vi.fn(),
  validateAccount: vi.fn(),
}));

import { signUp as apiSignUp, resendEmail as apiResendEmail, validateAccount as apiValidateAccount } from "../../api/users";

const mockedSignUp = vi.mocked(apiSignUp);
const mockedResendEmail = vi.mocked(apiResendEmail);
const mockedValidateAccount = vi.mocked(apiValidateAccount);

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

describe("signUpStore", () => {
  describe("signUp", () => {
    it("sets loading during request", async () => {
      let resolve: (v: { token: string; tenant: string }) => void;
      mockedSignUp.mockReturnValue(new Promise((r) => { resolve = r; }));

      const promise = useSignUpStore.getState().signUp({
        name: "Test", email: "t@t.com", username: "test", password: "pass1", email_marketing: false,
      });

      expect(useSignUpStore.getState().signUpLoading).toBe(true);

      resolve!({ token: "tok", tenant: "ten" });
      await promise;

      expect(useSignUpStore.getState().signUpLoading).toBe(false);
    });

    it("stores token and tenant on success and returns token", async () => {
      mockedSignUp.mockResolvedValue({ token: "jwt-token", tenant: "tenant-abc" });

      const result = await useSignUpStore.getState().signUp({
        name: "Test", email: "t@t.com", username: "test", password: "pass1", email_marketing: false,
      });

      expect(result).toBe("jwt-token");
      expect(useSignUpStore.getState().signUpToken).toBe("jwt-token");
      expect(useSignUpStore.getState().signUpTenant).toBe("tenant-abc");
    });

    it("returns null and stores null token when response has no token (normal flow)", async () => {
      mockedSignUp.mockResolvedValue({});

      const result = await useSignUpStore.getState().signUp({
        name: "Test", email: "t@t.com", username: "test", password: "pass1", email_marketing: false,
      });

      expect(result).toBeNull();
      expect(useSignUpStore.getState().signUpToken).toBeNull();
      expect(useSignUpStore.getState().signUpTenant).toBeNull();
    });

    it("sets signUpServerFields on 400/409 with field array and returns null", async () => {
      mockedSignUp.mockRejectedValue(createAxiosError(400, ["username", "email"]));

      const result = await useSignUpStore.getState().signUp({
        name: "Test", email: "t@t.com", username: "test", password: "pass1", email_marketing: false,
      });

      expect(result).toBeNull();
      expect(useSignUpStore.getState().signUpLoading).toBe(false);
      expect(useSignUpStore.getState().signUpServerFields).toEqual(["username", "email"]);
      expect(useSignUpStore.getState().signUpError).toBeNull();
    });

    it("sets signUpServerFields on 409 with field array and returns null", async () => {
      mockedSignUp.mockRejectedValue(createAxiosError(409, ["username"]));

      const result = await useSignUpStore.getState().signUp({
        name: "Test", email: "t@t.com", username: "test", password: "pass1", email_marketing: false,
      });

      expect(result).toBeNull();
      expect(useSignUpStore.getState().signUpServerFields).toEqual(["username"]);
      expect(useSignUpStore.getState().signUpError).toBeNull();
    });

    it("falls through to generic error when 400 body is not an array", async () => {
      mockedSignUp.mockRejectedValue(createAxiosError(400, { message: "validation error" }));

      const result = await useSignUpStore.getState().signUp({
        name: "Test", email: "t@t.com", username: "test", password: "pass1", email_marketing: false,
      });

      expect(result).toBeNull();
      expect(useSignUpStore.getState().signUpServerFields).toEqual([]);
      expect(useSignUpStore.getState().signUpError).toBe("An error occurred. Please try again.");
    });

    it("sets signUpError on non-field errors and returns null", async () => {
      mockedSignUp.mockRejectedValue(new Error("network error"));

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
      mockedSignUp.mockRejectedValue(new Error("network error"));

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
      mockedResendEmail.mockResolvedValue(undefined);

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
      let resolve: () => void;
      mockedResendEmail.mockReturnValue(new Promise<void>((r) => { resolve = r; }));

      const promise = useSignUpStore.getState().resendEmail("testuser");
      expect(useSignUpStore.getState().resendLoading).toBe(true);

      resolve!();
      await promise;

      expect(useSignUpStore.getState().resendLoading).toBe(false);
    });
  });

  describe("validateAccount", () => {
    it("transitions to success on 200", async () => {
      mockedValidateAccount.mockResolvedValue(undefined);

      await useSignUpStore.getState().validateAccount("t@t.com", "valid-token");

      expect(useSignUpStore.getState().validationStatus).toBe("success");
    });

    it("transitions to failed-token on 400 (expired token)", async () => {
      mockedValidateAccount.mockRejectedValue(createAxiosError(400));

      await useSignUpStore.getState().validateAccount("t@t.com", "expired-token");

      expect(useSignUpStore.getState().validationStatus).toBe("failed-token");
    });

    it("transitions to failed-token on 401 (wrong token)", async () => {
      mockedValidateAccount.mockRejectedValue(createAxiosError(401));

      await useSignUpStore.getState().validateAccount("t@t.com", "wrong-token");

      expect(useSignUpStore.getState().validationStatus).toBe("failed-token");
    });

    it("transitions to failed on 404 (user not found)", async () => {
      mockedValidateAccount.mockRejectedValue(createAxiosError(404));

      await useSignUpStore.getState().validateAccount("t@t.com", "unknown-user");

      expect(useSignUpStore.getState().validationStatus).toBe("failed");
    });

    it("transitions to failed on other errors", async () => {
      mockedValidateAccount.mockRejectedValue(createAxiosError(500));

      await useSignUpStore.getState().validateAccount("t@t.com", "bad-token");

      expect(useSignUpStore.getState().validationStatus).toBe("failed");
    });

    it("does not update state when the request is aborted", async () => {
      mockedValidateAccount.mockRejectedValue(new CanceledError());

      await useSignUpStore.getState().validateAccount("t@t.com", "tok");

      // Status must stay at "processing" — no terminal state set after an abort.
      expect(useSignUpStore.getState().validationStatus).toBe("processing");
    });

    it("sets processing during request", async () => {
      let resolve: () => void;
      mockedValidateAccount.mockReturnValue(new Promise<void>((r) => { resolve = r; }));

      const promise = useSignUpStore.getState().validateAccount("t@t.com", "tok");
      expect(useSignUpStore.getState().validationStatus).toBe("processing");

      resolve!();
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
