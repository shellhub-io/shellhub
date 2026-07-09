import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useAuthStore } from "@/stores/authStore";
import { EditProfileDrawer } from "@/pages/Profile";

const defaultProps = {
  open: true,
  onClose: vi.fn(),
  currentName: "Test User",
  currentUsername: "testuser",
  currentEmail: "user@example.com",
  currentRecoveryEmail: "recovery@example.com",
};

describe("EditProfileDrawer", () => {
  beforeEach(() => {
    useAuthStore.setState({ updateProfile: vi.fn() } as Partial<
      ReturnType<typeof useAuthStore.getState>
    >);
  });

  async function changeEmail(value: string, props = defaultProps) {
    const user = userEvent.setup();
    render(<EditProfileDrawer {...props} />);
    const input = screen.getByLabelText("Email");
    await user.clear(input);
    await user.type(input, value);
    return user;
  }

  describe("recovery email validation guard", () => {
    it("shows error when primary email changes to case-insensitively match recovery email", async () => {
      await changeEmail("Recovery@Example.COM");

      await waitFor(() =>
        expect(
          screen.getByText("Must be different from your email"),
        ).toBeInTheDocument(),
      );
    });

    it("does not show recovery email error when primary email changes to a different address", async () => {
      await changeEmail("new@example.com");

      await waitFor(() =>
        expect(
          screen.queryByText("Must be different from your email"),
        ).not.toBeInTheDocument(),
      );
    });

    it("does not show recovery email error when recovery email is empty and primary email changes", async () => {
      await changeEmail("new@example.com", {
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
      render(<EditProfileDrawer {...defaultProps} />);
      const input = screen.getByLabelText("Email");
      await user.clear(input);

      await waitFor(() =>
        expect(screen.getByText("Email is required")).toBeInTheDocument(),
      );
      expect(screen.getByRole("button", { name: /^save$/i })).toBeDisabled();
    });
  });
});
