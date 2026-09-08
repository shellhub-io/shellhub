import { describe, it, expect, vi, beforeEach } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { useSignUpStore } from "../signUpStore";

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
});

const payload = {
  name: "Test",
  email: "t@t.com",
  username: "test",
  password: "pass1",
  email_marketing: false,
};

describe("signUpStore", () => {
  describe("signUp", () => {
    vi.spyOn(console, "warn").mockImplementation(() => {});

    it("sets loading during request", async () => {
      let resolveHandler!: (r: Response) => void;
      const handlerReady = new Promise<void>((ready) => {
        server.use(
          http.post(
            "*/api/register",
            () =>
              new Promise<Response>((resolve) => {
                resolveHandler = resolve;
                ready();
              }),
          ),
        );
      });

      const promise = useSignUpStore.getState().signUp(payload);
      await handlerReady;
      expect(useSignUpStore.getState().signUpLoading).toBe(true);

      resolveHandler(HttpResponse.json({ token: "tok", tenant: "ten" }));
      await promise;

      expect(useSignUpStore.getState().signUpLoading).toBe(false);
    });

    it("stores token and tenant on success and returns token", async () => {
      server.use(
        http.post("*/api/register", () =>
          HttpResponse.json({ token: "jwt-token", tenant: "tenant-abc" }),
        ),
      );

      const result = await useSignUpStore.getState().signUp(payload);

      expect(result).toBe("jwt-token");
      expect(useSignUpStore.getState().signUpToken).toBe("jwt-token");
      expect(useSignUpStore.getState().signUpTenant).toBe("tenant-abc");
    });

    it("returns null and stores null token when response has no token (normal flow)", async () => {
      server.use(http.post("*/api/register", () => HttpResponse.json({})));

      const result = await useSignUpStore.getState().signUp(payload);

      expect(result).toBeNull();
      expect(useSignUpStore.getState().signUpToken).toBeNull();
      expect(useSignUpStore.getState().signUpTenant).toBeNull();
    });

    it("sets signUpServerFields on a 400 carrying per-field detail and returns null", async () => {
      server.use(
        http.post("*/api/register", () =>
          HttpResponse.json(
            {
              message: "user invalid",
              fields: { username: "required", email: "invalid" },
            },
            { status: 400 },
          ),
        ),
      );

      const result = await useSignUpStore.getState().signUp(payload);

      expect(result).toBeNull();
      expect(useSignUpStore.getState().signUpLoading).toBe(false);
      expect(useSignUpStore.getState().signUpServerFields).toEqual([
        "username",
        "email",
      ]);
      expect(useSignUpStore.getState().signUpError).toBeNull();
    });

    it("sets signUpServerFields on a 409 carrying per-field detail and returns null", async () => {
      server.use(
        http.post("*/api/register", () =>
          HttpResponse.json(
            {
              message: "user duplicated",
              fields: { username: "duplicated" },
            },
            { status: 409 },
          ),
        ),
      );

      const result = await useSignUpStore.getState().signUp(payload);

      expect(result).toBeNull();
      expect(useSignUpStore.getState().signUpServerFields).toEqual([
        "username",
      ]);
      expect(useSignUpStore.getState().signUpError).toBeNull();
    });

    it("falls through to the status message when the 400 body carries no fields", async () => {
      server.use(
        http.post("*/api/register", () =>
          HttpResponse.json({ message: "validation error" }, { status: 400 }),
        ),
      );

      const result = await useSignUpStore.getState().signUp(payload);

      expect(result).toBeNull();
      expect(useSignUpStore.getState().signUpServerFields).toEqual([]);
      expect(useSignUpStore.getState().signUpError).toBe(
        "Some values are invalid. Review the form and try again.",
      );
    });

    it("sets signUpError on non-field errors and returns null", async () => {
      server.use(http.post("*/api/register", () => HttpResponse.error()));

      const result = await useSignUpStore.getState().signUp(payload);

      expect(result).toBeNull();
      expect(useSignUpStore.getState().signUpLoading).toBe(false);
      expect(useSignUpStore.getState().signUpError).toBe(
        "Something went wrong. Please try again.",
      );
      expect(useSignUpStore.getState().signUpServerFields).toEqual([]);
    });

    it("clears stale token and tenant at the start of a new attempt", async () => {
      useSignUpStore.setState({
        signUpToken: "old-token",
        signUpTenant: "old-tenant",
      });
      server.use(http.post("*/api/register", () => HttpResponse.error()));

      await useSignUpStore.getState().signUp(payload);

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
      server.use(
        http.post(
          "*/api/user/resend_email",
          () => new HttpResponse(null, { status: 200 }),
        ),
      );

      const result = await useSignUpStore.getState().resendEmail("testuser");

      expect(result).toBe(true);
      expect(useSignUpStore.getState().resendLoading).toBe(false);
      expect(useSignUpStore.getState().resendError).toBeNull();
    });

    it("returns false and sets resendError on failure", async () => {
      server.use(
        http.post("*/api/user/resend_email", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );

      const result = await useSignUpStore.getState().resendEmail("testuser");

      expect(result).toBe(false);
      expect(useSignUpStore.getState().resendLoading).toBe(false);
      expect(useSignUpStore.getState().resendError).toBe(
        "Failed to resend email. Please try again.",
      );
    });

    it("sets loading during request", async () => {
      let resolveHandler!: (r: Response) => void;
      const handlerReady = new Promise<void>((ready) => {
        server.use(
          http.post(
            "*/api/user/resend_email",
            () =>
              new Promise<Response>((resolve) => {
                resolveHandler = resolve;
                ready();
              }),
          ),
        );
      });

      const promise = useSignUpStore.getState().resendEmail("testuser");
      await handlerReady;
      expect(useSignUpStore.getState().resendLoading).toBe(true);

      resolveHandler(new HttpResponse(null, { status: 200 }));
      await promise;

      expect(useSignUpStore.getState().resendLoading).toBe(false);
    });
  });

  describe("validateAccount", () => {
    it("transitions to success on 200", async () => {
      server.use(
        http.get(
          "*/api/user/validation_account",
          () => new HttpResponse(null, { status: 200 }),
        ),
      );

      await useSignUpStore.getState().validateAccount("t@t.com", "valid-token");

      expect(useSignUpStore.getState().validationStatus).toBe("success");
    });

    it("transitions to failed-token on 400 (expired token)", async () => {
      server.use(
        http.get("*/api/user/validation_account", () =>
          HttpResponse.json({}, { status: 400 }),
        ),
      );

      await useSignUpStore
        .getState()
        .validateAccount("t@t.com", "expired-token");

      expect(useSignUpStore.getState().validationStatus).toBe("failed-token");
    });

    it("transitions to failed-token on 401 (wrong token)", async () => {
      server.use(
        http.get("*/api/user/validation_account", () =>
          HttpResponse.json({}, { status: 401 }),
        ),
      );

      await useSignUpStore.getState().validateAccount("t@t.com", "wrong-token");

      expect(useSignUpStore.getState().validationStatus).toBe("failed-token");
    });

    it("transitions to failed on 404 (user not found)", async () => {
      server.use(
        http.get("*/api/user/validation_account", () =>
          HttpResponse.json({}, { status: 404 }),
        ),
      );

      await useSignUpStore
        .getState()
        .validateAccount("t@t.com", "unknown-user");

      expect(useSignUpStore.getState().validationStatus).toBe("failed");
    });

    it("transitions to failed on other errors", async () => {
      server.use(
        http.get("*/api/user/validation_account", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );

      await useSignUpStore.getState().validateAccount("t@t.com", "bad-token");

      expect(useSignUpStore.getState().validationStatus).toBe("failed");
    });

    it("does not update state when the request is aborted", async () => {
      server.use(
        http.get("*/api/user/validation_account", () =>
          HttpResponse.json({}, { status: 200 }),
        ),
      );

      const controller = new AbortController();
      controller.abort();

      await useSignUpStore
        .getState()
        .validateAccount("t@t.com", "tok", controller.signal);

      expect(useSignUpStore.getState().validationStatus).toBe("processing");
    });

    it("sets processing during request", async () => {
      let resolveHandler!: (r: Response) => void;
      const handlerReady = new Promise<void>((ready) => {
        server.use(
          http.get(
            "*/api/user/validation_account",
            () =>
              new Promise<Response>((resolve) => {
                resolveHandler = resolve;
                ready();
              }),
          ),
        );
      });

      const promise = useSignUpStore
        .getState()
        .validateAccount("t@t.com", "tok");
      await handlerReady;
      expect(useSignUpStore.getState().validationStatus).toBe("processing");

      resolveHandler(new HttpResponse(null, { status: 200 }));
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
