import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { useSignUpStore } from "@/stores/signUpStore";
import SignUp from "../SignUp";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

vi.mock("@/client", () => ({
  registerUser: vi.fn(),
  resendEmail: vi.fn(),
  getValidateAccount: vi.fn(),
}));

import { registerUser as registerUserSdk } from "@/client";

const mockedRegisterUser = vi.mocked(registerUserSdk);

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

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

function renderSignUp(search = "") {
  return render(
    <MemoryRouter initialEntries={[`/signup${search}`]}>
      <SignUp />
    </MemoryRouter>,
  );
}

/**
 * Fill all required fields with valid data and check privacy policy.
 * Does NOT submit.
 */
async function fillValidForm(
  user: ReturnType<typeof userEvent.setup>,
  overrides: {
    name?: string;
    username?: string;
    email?: string;
    password?: string;
    confirmPassword?: string;
    acceptMarketing?: boolean;
  } = {},
) {
  const {
    name = "Alice Smith",
    username = "alice",
    email = "alice@example.com",
    password = "Secret123",
    confirmPassword = "Secret123",
    acceptMarketing = false,
  } = overrides;

  await user.type(screen.getByLabelText(/^name$/i), name);
  await user.type(screen.getByLabelText(/^username$/i), username);
  await user.type(screen.getByLabelText(/^email$/i), email);
  await user.type(screen.getByLabelText(/^password$/i), password);
  await user.type(
    screen.getByLabelText(/^confirm password$/i),
    confirmPassword,
  );

  // Accept privacy policy (required)
  await user.click(screen.getByLabelText(/privacy policy/i));

  if (acceptMarketing) {
    await user.click(screen.getByLabelText(/receive news and updates/i));
  }
}

/* ------------------------------------------------------------------ */
/* Setup / teardown                                                    */
/* ------------------------------------------------------------------ */

afterEach(cleanup);

beforeEach(() => {
  mockNavigate.mockReset();
  mockedRegisterUser.mockReset();
  useSignUpStore.setState({
    signUpLoading: false,
    signUpError: null,
    signUpServerFields: [],
    signUpToken: null,
    signUpTenant: null,
  });
});

/* ================================================================== */
/* Tests                                                               */
/* ================================================================== */

describe("SignUp", () => {
  /* ---------------------------------------------------------------- */
  /* 1. Submit button disabled on initial render with empty defaults   */
  /* ---------------------------------------------------------------- */
  describe("initial render", () => {
    it("disables the submit button when all fields are empty", () => {
      renderSignUp();

      const submit = screen.getByRole("button", { name: /create account/i });
      expect(submit).toBeDisabled();
    });
  });

  /* ---------------------------------------------------------------- */
  /* 2. Filling valid data + checking privacy policy calls signUp     */
  /*    with correct payload including email_marketing                 */
  /* ---------------------------------------------------------------- */
  describe("successful submission", () => {
    it("calls signUp with correct payload including email_marketing when form is valid", async () => {
      mockedRegisterUser.mockResolvedValue(mockSdkResponse({}));
      const user = userEvent.setup();
      renderSignUp();

      await fillValidForm(user, { acceptMarketing: true });
      await user.click(screen.getByRole("button", { name: /create account/i }));

      await waitFor(() => expect(mockedRegisterUser).toHaveBeenCalledTimes(1));
      expect(mockedRegisterUser).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({
            name: "Alice Smith",
            username: "alice",
            email: "alice@example.com",
            password: "Secret123",
            email_marketing: true,
          }),
        }),
      );
    });

    it("calls signUp with email_marketing: false when marketing checkbox is unchecked", async () => {
      mockedRegisterUser.mockResolvedValue(mockSdkResponse({}));
      const user = userEvent.setup();
      renderSignUp();

      await fillValidForm(user, { acceptMarketing: false });
      await user.click(screen.getByRole("button", { name: /create account/i }));

      await waitFor(() => expect(mockedRegisterUser).toHaveBeenCalledTimes(1));
      expect(mockedRegisterUser).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({ email_marketing: false }),
        }),
      );
    });
  });

  /* ---------------------------------------------------------------- */
  /* 3. Field errors appear after blur not before                     */
  /* ---------------------------------------------------------------- */
  describe("field validation on blur", () => {
    it("does not show a name error before the field is touched", () => {
      renderSignUp();
      expect(screen.queryByText(/name must be/i)).not.toBeInTheDocument();
    });

    it("shows name error after blur when field is empty", async () => {
      const user = userEvent.setup();
      renderSignUp();

      const nameInput = screen.getByLabelText(/^name$/i);
      await user.click(nameInput);
      await user.tab(); // blur

      expect(await screen.findByText(/name must be/i)).toBeInTheDocument();
    });

    it("does not show username error before the field is touched", () => {
      renderSignUp();
      expect(screen.queryByText(/username must be/i)).not.toBeInTheDocument();
    });

    it("shows username error after blur when field is too short", async () => {
      const user = userEvent.setup();
      renderSignUp();

      await user.type(screen.getByLabelText(/^username$/i), "ab");
      await user.tab();

      expect(await screen.findByText(/username must be/i)).toBeInTheDocument();
    });
  });

  /* ---------------------------------------------------------------- */
  /* 4. Privacy unchecked keeps submit disabled                       */
  /* ---------------------------------------------------------------- */
  describe("privacy policy gate", () => {
    it("keeps submit disabled when all text fields are valid but privacy policy is unchecked", async () => {
      const user = userEvent.setup();
      renderSignUp();

      // Fill all required text fields but skip checking the privacy checkbox
      await user.type(screen.getByLabelText(/^name$/i), "Alice Smith");
      await user.type(screen.getByLabelText(/^username$/i), "alice");
      await user.type(screen.getByLabelText(/^email$/i), "alice@example.com");
      await user.type(screen.getByLabelText(/^password$/i), "Secret123");
      await user.type(
        screen.getByLabelText(/^confirm password$/i),
        "Secret123",
      );

      const submit = screen.getByRole("button", { name: /create account/i });
      expect(submit).toBeDisabled();
    });
  });

  /* ---------------------------------------------------------------- */
  /* 5. Password mismatch shows 'Passwords do not match'              */
  /* ---------------------------------------------------------------- */
  describe("password mismatch", () => {
    it("shows 'Passwords do not match' when confirmPassword differs from password", async () => {
      const user = userEvent.setup();
      renderSignUp();

      await user.type(screen.getByLabelText(/^password$/i), "Secret123");
      await user.type(
        screen.getByLabelText(/^confirm password$/i),
        "DifferentPass",
      );
      await user.tab();

      expect(
        await screen.findByText(/passwords do not match/i),
      ).toBeInTheDocument();
    });

    it("does not show mismatch error before confirmPassword is touched", async () => {
      const user = userEvent.setup();
      renderSignUp();

      await user.type(screen.getByLabelText(/^password$/i), "Secret123");
      // No blur on confirmPassword

      expect(
        screen.queryByText(/passwords do not match/i),
      ).not.toBeInTheDocument();
    });
  });

  /* ---------------------------------------------------------------- */
  /* 6. Server field errors render on correct fields, disable submit,  */
  /*    and clear after field edit                                      */
  /* ---------------------------------------------------------------- */
  describe("server field errors", () => {
    /**
     * Simulate a 400 server response with field errors, matching the shape
     * that isSdkError accepts: an error with numeric `status` that is also
     * Array-like (the store checks `Array.isArray(error)`).
     */
    function makeServerFieldError(fields: string[], status = 400) {
      // The store does: isSdkError(error) && Array.isArray(error)
      // isSdkError checks for numeric `status` on the object
      // Array.isArray requires the value to actually be an array
      const err = Object.assign([...fields], { status }) as unknown as Error;
      return err;
    }

    it("shows server-side username error on the username field and disables submit", async () => {
      mockedRegisterUser.mockRejectedValue(makeServerFieldError(["username"]));
      const user = userEvent.setup();
      renderSignUp();

      await fillValidForm(user);
      await user.click(screen.getByRole("button", { name: /create account/i }));

      // The server error should appear on the username field
      expect(
        await screen.findByText(/this username already exists/i),
      ).toBeInTheDocument();

      // Submit should be disabled because of the server field error
      expect(
        screen.getByRole("button", { name: /create account/i }),
      ).toBeDisabled();
    });

    it("shows server-side email error on the email field", async () => {
      mockedRegisterUser.mockRejectedValue(makeServerFieldError(["email"]));
      const user = userEvent.setup();
      renderSignUp();

      await fillValidForm(user);
      await user.click(screen.getByRole("button", { name: /create account/i }));

      expect(
        await screen.findByText(/this email is invalid or already in use/i),
      ).toBeInTheDocument();
    });

    it("clears the server field error after the user edits that field", async () => {
      mockedRegisterUser.mockRejectedValue(makeServerFieldError(["username"]));
      const user = userEvent.setup();
      renderSignUp();

      await fillValidForm(user);
      await user.click(screen.getByRole("button", { name: /create account/i }));

      await screen.findByText(/this username already exists/i);

      // Edit the username field — server error should be cleared. The clear is
      // synchronous with the keystroke, so waitForElementToBeRemoved can't be
      // used (the element is already gone by call time); assert absence instead.
      await user.type(screen.getByLabelText(/^username$/i), "a");

      await waitFor(() =>
        expect(
          screen.queryByText(/this username already exists/i),
        ).not.toBeInTheDocument(),
      );

      // The re-enable is the headline behavior: editing the field must clear the
      // server-field gate (clearSignUpServerField), not just RHF's error text.
      expect(
        screen.getByRole("button", { name: /create account/i }),
      ).toBeEnabled();
    });
  });
});
