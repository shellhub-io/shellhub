import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { useSignUpStore } from "@/stores/signUpStore";
import {
  clearPendingDeviceCode,
  setPendingDeviceCode,
} from "@/utils/navigation";
import SignUp from "../SignUp";

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

function renderSignUp(search = "") {
  return render(
    <MemoryRouter initialEntries={[`/signup${search}`]}>
      <SignUp />
    </MemoryRouter>,
  );
}

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

  await user.click(screen.getByLabelText(/privacy policy/i));

  if (acceptMarketing) {
    await user.click(screen.getByLabelText(/news and updates/i));
  }
}

function captureRegisterBody() {
  const captured: { body?: Record<string, unknown> } = {};
  server.use(
    http.post("*/api/register", async ({ request }) => {
      captured.body = (await request.json()) as Record<string, unknown>;
      return new HttpResponse(null, { status: 204 });
    }),
  );
  return captured;
}

beforeEach(() => {
  vi.clearAllMocks();
  clearPendingDeviceCode();
  useSignUpStore.setState({
    signUpLoading: false,
    signUpError: null,
    signUpServerFields: {},
    signUpToken: null,
    signUpTenant: null,
  });
  server.use(
    http.post("*/api/register", () => new HttpResponse(null, { status: 204 })),
  );
});

describe("SignUp", () => {
  it("tells the user the namespace and first device come after the account", () => {
    renderSignUp();

    expect(
      screen.getByText(
        /next you name a namespace and connect your first device/i,
      ),
    ).toBeInTheDocument();
    expect(
      screen.queryByText(/namespace is created with it/i),
    ).not.toBeInTheDocument();
  });

  describe("successful submission", () => {
    it("navigates to confirm-account after valid submission", async () => {
      const user = userEvent.setup();
      renderSignUp();

      await fillValidForm(user, { acceptMarketing: true });
      await user.click(screen.getByRole("button", { name: /create account/i }));

      await waitFor(() =>
        expect(mockNavigate).toHaveBeenCalledWith(
          expect.stringContaining("/confirm-account"),
        ),
      );
    });

    it("sends the pending device code with the registration", async () => {
      setPendingDeviceCode("wxyz-2k7q");
      const captured = captureRegisterBody();
      const user = userEvent.setup();
      renderSignUp();

      await fillValidForm(user);
      await user.click(screen.getByRole("button", { name: /create account/i }));

      await waitFor(() =>
        expect(captured.body).toMatchObject({ device_code: "WXYZ2K7Q" }),
      );
    });

    it("sends no device code when none is pending", async () => {
      const captured = captureRegisterBody();
      const user = userEvent.setup();
      renderSignUp();

      await fillValidForm(user);
      await user.click(screen.getByRole("button", { name: /create account/i }));

      await waitFor(() =>
        expect(captured.body?.email).toBe("alice@example.com"),
      );
      expect(captured.body).not.toHaveProperty("device_code");
    });
  });

  describe("field validation on blur", () => {
    it("shows name error only after blurring an empty field", async () => {
      const user = userEvent.setup();
      renderSignUp();

      expect(screen.queryByText(/name must be/i)).not.toBeInTheDocument();

      const nameInput = screen.getByLabelText(/^name$/i);
      await user.click(nameInput);
      await user.tab(); // blur

      expect(await screen.findByText(/name must be/i)).toBeInTheDocument();
    });

    it("shows username error only after blurring a too-short field", async () => {
      const user = userEvent.setup();
      renderSignUp();

      expect(screen.queryByText(/username must be/i)).not.toBeInTheDocument();

      await user.type(screen.getByLabelText(/^username$/i), "ab");
      await user.tab();

      expect(await screen.findByText(/username must be/i)).toBeInTheDocument();
    });
  });

  describe("privacy policy gate", () => {
    it("keeps submit disabled when all text fields are valid but privacy policy is unchecked", async () => {
      const user = userEvent.setup();
      renderSignUp();

      expect(
        screen.getByRole("button", { name: /create account/i }),
      ).toBeDisabled();

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

  describe("password mismatch", () => {
    it("shows 'Passwords do not match' only once confirmPassword is touched", async () => {
      const user = userEvent.setup();
      renderSignUp();

      await user.type(screen.getByLabelText(/^password$/i), "Secret123");

      expect(
        screen.queryByText(/passwords do not match/i),
      ).not.toBeInTheDocument();

      await user.type(
        screen.getByLabelText(/^confirm password$/i),
        "DifferentPass",
      );
      await user.tab();

      expect(
        await screen.findByText(/passwords do not match/i),
      ).toBeInTheDocument();
    });
  });

  describe("server field errors", () => {
    function setServerFieldError(fields: Record<string, string>) {
      server.use(
        http.post("*/api/register", () =>
          HttpResponse.json(
            { message: "user invalid", fields },
            { status: 400 },
          ),
        ),
      );
    }

    it("shows server-side username error on the username field and disables submit", async () => {
      setServerFieldError({ username: "invalid" });
      const user = userEvent.setup();
      renderSignUp();

      await fillValidForm(user);
      await user.click(screen.getByRole("button", { name: /create account/i }));

      expect(
        await screen.findByText(/this username already exists/i),
      ).toBeInTheDocument();

      expect(
        screen.getByRole("button", { name: /create account/i }),
      ).toBeDisabled();
    });

    it("shows server-side email error on the email field", async () => {
      setServerFieldError({ email: "invalid" });
      const user = userEvent.setup();
      renderSignUp();

      await fillValidForm(user);
      await user.click(screen.getByRole("button", { name: /create account/i }));

      expect(
        await screen.findByText(/this email is invalid or already in use/i),
      ).toBeInTheDocument();
    });

    it.each([
      ["disposable", /use a permanent email address/i],
      [
        "unreachable",
        /this domain cannot receive email\. check the spelling\./i,
      ],
    ])(
      "says why the server refused the email as %s, and keeps submit disabled until it changes",
      async (reason, message) => {
        setServerFieldError({ email: reason });
        const user = userEvent.setup();
        renderSignUp();

        await fillValidForm(user);
        await user.click(
          screen.getByRole("button", { name: /create account/i }),
        );

        expect(await screen.findByText(message)).toBeInTheDocument();
        expect(
          screen.getByRole("button", { name: /create account/i }),
        ).toBeDisabled();

        await user.type(screen.getByLabelText(/^email$/i), "m");

        await waitFor(() =>
          expect(screen.queryByText(message)).not.toBeInTheDocument(),
        );
        expect(
          screen.getByRole("button", { name: /create account/i }),
        ).toBeEnabled();
      },
    );

    it("clears the server field error after the user edits that field", async () => {
      setServerFieldError({ username: "invalid" });
      const user = userEvent.setup();
      renderSignUp();

      await fillValidForm(user);
      await user.click(screen.getByRole("button", { name: /create account/i }));

      await screen.findByText(/this username already exists/i);

      await user.type(screen.getByLabelText(/^username$/i), "a");

      await waitFor(() =>
        expect(
          screen.queryByText(/this username already exists/i),
        ).not.toBeInTheDocument(),
      );

      expect(
        screen.getByRole("button", { name: /create account/i }),
      ).toBeEnabled();
    });
  });
});
