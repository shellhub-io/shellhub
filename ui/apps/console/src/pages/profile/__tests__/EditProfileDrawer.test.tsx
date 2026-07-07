import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor, cleanup } from "@testing-library/react";
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

beforeEach(() => {
  useAuthStore.setState({ updateProfile: vi.fn() } as Partial<ReturnType<typeof useAuthStore.getState>>);
});

afterEach(cleanup);

describe("EditProfileDrawer — recovery email validation guard", () => {
  it("shows error when primary email changes to case-insensitively match recovery email", async () => {
    render(<EditProfileDrawer {...defaultProps} />);
    fireEvent.change(screen.getByPlaceholderText("you@example.com"), {
      target: { value: "Recovery@Example.COM" },
    });
    await waitFor(() =>
      expect(
        screen.getByText("Must be different from your email"),
      ).toBeInTheDocument(),
    );
  });

  it("does not show recovery email error when primary email changes to a different address", async () => {
    render(<EditProfileDrawer {...defaultProps} />);
    fireEvent.change(screen.getByPlaceholderText("you@example.com"), {
      target: { value: "new@example.com" },
    });
    await waitFor(() =>
      expect(
        screen.queryByText("Must be different from your email"),
      ).not.toBeInTheDocument(),
    );
  });

  it("does not show recovery email error when recovery email is empty and primary email changes", async () => {
    render(<EditProfileDrawer {...defaultProps} currentRecoveryEmail="" />);
    fireEvent.change(screen.getByPlaceholderText("you@example.com"), {
      target: { value: "new@example.com" },
    });
    await waitFor(() =>
      expect(
        screen.queryByText("Must be different from your email"),
      ).not.toBeInTheDocument(),
    );
  });
});

describe("EditProfileDrawer — RHF resolver: no errors on empty optional-when-editing fields", () => {
  it("does not show 'Email is required' when email field is cleared (resolver skips empty)", async () => {
    render(<EditProfileDrawer {...defaultProps} />);
    fireEvent.change(screen.getByPlaceholderText("you@example.com"), {
      target: { value: "" },
    });
    await waitFor(() =>
      expect(screen.queryByText("Email is required")).not.toBeInTheDocument(),
    );
  });
});
