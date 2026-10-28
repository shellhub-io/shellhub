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

async function changeField(label: string, value: string, props = defaultProps) {
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
});
