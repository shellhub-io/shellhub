import type { ComponentProps } from "react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { useAuthStore } from "@/stores/authStore";
import { EditProfileModal } from "../AccountProfile";

const realUpdateProfile = useAuthStore.getState().updateProfile;

const defaultProps = {
  open: true,
  onClose: vi.fn(),
  currentName: "Test User",
  currentEmail: "user@example.com",
  currentRecoveryEmail: "recovery@example.com",
};

async function changeField(
  label: string,
  value: string,
  props: ComponentProps<typeof EditProfileModal> = defaultProps,
) {
  const user = userEvent.setup();
  render(<EditProfileModal {...props} />);
  const input = screen.getByLabelText(label);
  await user.clear(input);
  await user.type(input, value);
  return { user, input };
}

describe("EditProfileModal", () => {
  beforeEach(() => {
    useAuthStore.setState({ updateProfile: vi.fn() } as Partial<
      ReturnType<typeof useAuthStore.getState>
    >);
  });

  describe("recovery email validation guard", () => {
    it("shows error when primary email changes to case-insensitively match recovery email", async () => {
      await changeField("Email", "Recovery@Example.COM");

      await waitFor(() =>
        expect(
          screen.getByText("Must be different from your email"),
        ).toBeInTheDocument(),
      );
    });

    it("does not show recovery email error when primary email changes to a different address", async () => {
      await changeField("Email", "new@example.com");

      await waitFor(() =>
        expect(
          screen.queryByText("Must be different from your email"),
        ).not.toBeInTheDocument(),
      );
    });

    it("does not show recovery email error when recovery email is empty and primary email changes", async () => {
      await changeField("Email", "new@example.com", {
        ...defaultProps,
        currentRecoveryEmail: "",
      });

      await waitFor(() =>
        expect(
          screen.queryByText("Must be different from your email"),
        ).not.toBeInTheDocument(),
      );
    });
  });

  describe("name padding", () => {
    it("saves a padded name trimmed", async () => {
      const { user } = await changeField("Name", "  Alice  ");

      await user.click(screen.getByRole("button", { name: /^save$/i }));

      await waitFor(() =>
        expect(useAuthStore.getState().updateProfile).toHaveBeenCalledWith({
          name: "Alice",
        }),
      );
    });

    it("saves a stored name with its padding removed", async () => {
      const props = { ...defaultProps, currentName: " Alice " };
      const { user } = await changeField("Name", "Alice", props);

      await user.click(screen.getByRole("button", { name: /^save$/i }));

      await waitFor(() =>
        expect(useAuthStore.getState().updateProfile).toHaveBeenCalledWith({
          name: "Alice",
        }),
      );
    });

    it("blocks saving a blank stored name that only gains padding", async () => {
      await changeField("Name", "    ", {
        ...defaultProps,
        currentName: "   ",
      });

      expect(await screen.findByText("Name is required")).toBeInTheDocument();
      expect(screen.getByRole("button", { name: /^save$/i })).toBeDisabled();
    });
  });

  describe("required fields", () => {
    it("shows 'Email is required' and blocks submit when a required field is cleared", async () => {
      const user = userEvent.setup();
      render(<EditProfileModal {...defaultProps} />);
      const input = screen.getByLabelText("Email");
      await user.clear(input);

      await waitFor(() =>
        expect(screen.getByText("Email is required")).toBeInTheDocument(),
      );
      expect(screen.getByRole("button", { name: /^save$/i })).toBeDisabled();
    });
  });

  describe("refused email", () => {
    beforeEach(() => {
      vi.clearAllMocks();
      useAuthStore.setState({ updateProfile: realUpdateProfile } as Partial<
        ReturnType<typeof useAuthStore.getState>
      >);
    });

    function refuse(fields: Record<string, string>) {
      server.use(
        http.patch("*/api/users", () =>
          HttpResponse.json(
            { message: "bad request", fields },
            { status: 400 },
          ),
        ),
      );
    }

    it.each([
      ["Email", "email", "disposable", "Use a permanent email address"],
      [
        "Recovery Email",
        "recovery_email",
        "unreachable",
        "This domain cannot receive email. Check the spelling.",
      ],
    ])(
      "shows why the server refused the %s beside it",
      async (label, field, reason, message) => {
        refuse({ [field]: reason });
        const { user, input } = await changeField(
          label,
          "someone@refused.test",
        );
        await user.type(screen.getByLabelText("Current Password"), "secret");
        await user.click(screen.getByRole("button", { name: /^save$/i }));

        await screen.findByText(message);
        expect(input).toHaveAccessibleDescription(
          expect.stringContaining(message),
        );
        expect(
          screen.queryByText(/some fields have invalid values/i),
        ).not.toBeInTheDocument();
        expect(defaultProps.onClose).not.toHaveBeenCalled();
      },
    );
  });

  describe("current password", () => {
    it("is asked for only once an email changes", async () => {
      const user = userEvent.setup();
      render(<EditProfileModal {...defaultProps} />);

      await user.type(screen.getByLabelText("Name"), "x");
      expect(
        screen.queryByLabelText("Current Password"),
      ).not.toBeInTheDocument();

      await user.clear(screen.getByLabelText("Recovery Email"));
      await user.type(
        screen.getByLabelText("Recovery Email"),
        "other@example.com",
      );
      expect(screen.getByLabelText("Current Password")).toBeInTheDocument();
      expect(screen.getByRole("button", { name: /^save$/i })).toBeDisabled();
    });

    it.each([
      ["Email", "Email is required"],
      ["Recovery Email", "Recovery email can't be removed"],
    ])("is not asked for while the %s is cleared", async (label, message) => {
      const user = userEvent.setup();
      render(<EditProfileModal {...defaultProps} />);

      await user.clear(screen.getByLabelText(label));

      expect(await screen.findByText(message)).toBeInTheDocument();
      expect(
        screen.queryByLabelText("Current Password"),
      ).not.toBeInTheDocument();
    });

    it("is gone after the modal closes and opens again", async () => {
      const user = userEvent.setup();
      const { rerender } = render(<EditProfileModal {...defaultProps} />);
      await user.clear(screen.getByLabelText("Email"));
      await user.type(screen.getByLabelText("Email"), "new@example.com");
      expect(screen.getByLabelText("Current Password")).toBeInTheDocument();

      rerender(<EditProfileModal {...defaultProps} open={false} />);
      rerender(<EditProfileModal {...defaultProps} />);

      expect(await screen.findByLabelText("Email")).toHaveValue(
        "user@example.com",
      );
      expect(
        screen.queryByLabelText("Current Password"),
      ).not.toBeInTheDocument();
    });

    it("is sent with the changed email", async () => {
      const { user } = await changeField("Email", "new@example.com");
      await user.type(screen.getByLabelText("Current Password"), "secret");

      await user.click(screen.getByRole("button", { name: /^save$/i }));

      await waitFor(() =>
        expect(useAuthStore.getState().updateProfile).toHaveBeenCalledWith({
          email: "new@example.com",
          current_password: "secret",
        }),
      );
    });

    it("is marked incorrect when the server refuses it", async () => {
      useAuthStore.setState({ updateProfile: realUpdateProfile } as Partial<
        ReturnType<typeof useAuthStore.getState>
      >);
      server.use(
        http.patch("*/api/users", () =>
          HttpResponse.json({ message: "forbidden" }, { status: 403 }),
        ),
      );
      const { user } = await changeField("Email", "new@example.com");
      const password = screen.getByLabelText("Current Password");
      await user.type(password, "wrong");

      await user.click(screen.getByRole("button", { name: /^save$/i }));

      await screen.findByText("Current password is incorrect.");
      expect(password).toHaveAccessibleDescription(
        expect.stringContaining("Current password is incorrect."),
      );
    });

    it("names the lockout a wrong one starts", async () => {
      useAuthStore.setState({ updateProfile: realUpdateProfile } as Partial<
        ReturnType<typeof useAuthStore.getState>
      >);
      server.use(
        http.patch("*/api/users", () =>
          HttpResponse.json(
            { message: "too many wrong attempts" },
            {
              status: 429,
              headers: { "X-Account-Lockout": String(Math.floor(Date.now() / 1000) + 270) },
            },
          ),
        ),
      );
      const { user } = await changeField("Email", "new@example.com");
      await user.type(screen.getByLabelText("Current Password"), "wrong");

      await user.click(screen.getByRole("button", { name: /^save$/i }));

      expect(
        await screen.findByText("Too many wrong attempts. Try again in 4 minutes."),
      ).toBeInTheDocument();
    });
  });

  describe("email held for confirmation", () => {
    it("reports the new email pending when the account keeps the old one", async () => {
      const onEmailPending = vi.fn();
      useAuthStore.setState({
        email: "user@example.com",
        updateProfile: vi.fn(),
      } as Partial<ReturnType<typeof useAuthStore.getState>>);
      const { user } = await changeField("Email", "new@example.com", {
        ...defaultProps,
        onEmailPending,
      });
      await user.type(screen.getByLabelText("Current Password"), "secret");

      await user.click(screen.getByRole("button", { name: /^save$/i }));

      await waitFor(() =>
        expect(onEmailPending).toHaveBeenCalledWith("new@example.com"),
      );
    });

    it("reports nothing pending when the new email took effect", async () => {
      const onEmailPending = vi.fn();
      useAuthStore.setState({
        email: "new@example.com",
        updateProfile: vi.fn(),
      } as Partial<ReturnType<typeof useAuthStore.getState>>);
      const { user } = await changeField("Email", "new@example.com", {
        ...defaultProps,
        onEmailPending,
      });
      await user.type(screen.getByLabelText("Current Password"), "secret");

      await user.click(screen.getByRole("button", { name: /^save$/i }));

      await waitFor(() => expect(defaultProps.onClose).toHaveBeenCalled());
      expect(onEmailPending).not.toHaveBeenCalled();
    });
  });

  describe("recovery email held for confirmation", () => {
    it("reports the new recovery email pending when the account keeps the old one", async () => {
      const onRecoveryEmailPending = vi.fn();
      useAuthStore.setState({
        recoveryEmail: "recovery@example.com",
        updateProfile: vi.fn(),
      } as Partial<ReturnType<typeof useAuthStore.getState>>);
      const { user } = await changeField(
        "Recovery Email",
        "backup@example.com",
        {
          ...defaultProps,
          onRecoveryEmailPending,
        },
      );
      await user.type(screen.getByLabelText("Current Password"), "secret");

      await user.click(screen.getByRole("button", { name: /^save$/i }));

      await waitFor(() =>
        expect(onRecoveryEmailPending).toHaveBeenCalledWith(
          "backup@example.com",
        ),
      );
    });
  });

  describe("authenticator code", () => {
    const withMfa = { ...defaultProps, mfaEnabled: true };

    it("is asked for only once an email changes on an account with MFA", async () => {
      const user = userEvent.setup();
      render(<EditProfileModal {...withMfa} />);

      await user.type(screen.getByLabelText("Name"), "x");
      expect(
        screen.queryByLabelText("Authenticator or Recovery Code"),
      ).not.toBeInTheDocument();

      await user.clear(screen.getByLabelText("Email"));
      await user.type(screen.getByLabelText("Email"), "new@example.com");
      await user.type(screen.getByLabelText("Current Password"), "secret");
      expect(screen.getByLabelText("Authenticator or Recovery Code")).toBeInTheDocument();
      expect(screen.getByRole("button", { name: /^save$/i })).toBeDisabled();
    });

    it("is not asked for on an account without MFA", async () => {
      await changeField("Email", "new@example.com");

      expect(
        screen.queryByLabelText("Authenticator or Recovery Code"),
      ).not.toBeInTheDocument();
    });

    it("is sent with the changed email", async () => {
      const { user } = await changeField("Email", "new@example.com", withMfa);
      await user.type(screen.getByLabelText("Current Password"), "secret");
      await user.type(screen.getByLabelText("Authenticator or Recovery Code"), "123456");

      await user.click(screen.getByRole("button", { name: /^save$/i }));

      await waitFor(() =>
        expect(useAuthStore.getState().updateProfile).toHaveBeenCalledWith({
          email: "new@example.com",
          current_password: "secret",
          mfa_code: "123456",
        }),
      );
    });

    it("sends a recovery code trimmed", async () => {
      const { user } = await changeField("Email", "new@example.com", withMfa);
      await user.type(screen.getByLabelText("Current Password"), "secret");
      await user.type(
        screen.getByLabelText("Authenticator or Recovery Code"),
        " ABCDEFGHIJKLMNOP ",
      );

      await user.click(screen.getByRole("button", { name: /^save$/i }));

      await waitFor(() =>
        expect(useAuthStore.getState().updateProfile).toHaveBeenCalledWith({
          email: "new@example.com",
          current_password: "secret",
          mfa_code: "ABCDEFGHIJKLMNOP",
        }),
      );
    });

    it.each([
      ["is invalid", "The code is incorrect."],
      [
        "too many invalid codes, try again later",
        "Too many incorrect codes. Try again later.",
      ],
    ])("says why the server refused it (%s)", async (reason, message) => {
      useAuthStore.setState({ updateProfile: realUpdateProfile } as Partial<
        ReturnType<typeof useAuthStore.getState>
      >);
      server.use(
        http.patch("*/api/users", () =>
          HttpResponse.json(
            { message: "bad request", fields: { mfa_code: reason } },
            { status: 400 },
          ),
        ),
      );
      const { user } = await changeField("Email", "new@example.com", withMfa);
      await user.type(screen.getByLabelText("Current Password"), "secret");
      const code = screen.getByLabelText("Authenticator or Recovery Code");
      await user.type(code, "000000");

      await user.click(screen.getByRole("button", { name: /^save$/i }));

      await screen.findByText(message);
      expect(code).toHaveAccessibleDescription(
        expect.stringContaining(message),
      );
    });
  });

  it("locks both emails of an account its identity provider manages", () => {
    render(<EditProfileModal {...defaultProps} emailsManagedExternally />);

    expect(screen.getByLabelText("Email")).toBeDisabled();
    expect(screen.getByLabelText("Recovery Email")).toBeDisabled();
  });
});
