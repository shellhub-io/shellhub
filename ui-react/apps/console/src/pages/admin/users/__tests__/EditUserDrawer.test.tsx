import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import EditUserDrawer, { type EditableUser } from "../EditUserDrawer";
import { useUpdateUser } from "../../../../hooks/useAdminUserMutations";
import { useAuthStore } from "../../../../stores/authStore";
vi.mock("../../../../hooks/useAdminUserMutations", () => ({
  useUpdateUser: vi.fn(),
}));

vi.mock("../../../../utils/styles", () => ({
  LABEL: "label",
  INPUT: "input",
}));

vi.mock("../../../../components/common/Drawer", async () => ({
  default: (await import("./mocks")).MockDrawer,
}));

const mockMutateAsync = vi.fn();

const mockUser: EditableUser = {
  id: "u1",
  name: "Alice Smith",
  username: "alice",
  email: "alice@example.com",
  admin: false,
  status: "not-confirmed",
};

const confirmedUser: EditableUser = {
  ...mockUser,
  status: "confirmed",
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(useUpdateUser).mockReturnValue({
    mutateAsync: mockMutateAsync,
  } as never);
  useAuthStore.setState({ username: "admin" } as never);
});

function renderDrawer(
  overrides: Partial<{
    open: boolean;
    onClose: () => void;
    user: EditableUser | null;
  }> = {},
) {
  const defaults = { open: true, onClose: vi.fn(), user: mockUser };
  const props = { ...defaults, ...overrides };
  return { onClose: props.onClose, ...render(<EditUserDrawer {...props} />) };
}

describe("EditUserDrawer", () => {
  describe("rendering — closed", () => {
    it("renders nothing when open is false", () => {
      renderDrawer({ open: false });
      expect(screen.queryByText("Edit User")).not.toBeInTheDocument();
    });
  });

  describe("rendering — open", () => {
    it("renders the 'Edit User' title", () => {
      renderDrawer();
      expect(screen.getByText("Edit User")).toBeInTheDocument();
    });

    it("renders the Name input", () => {
      renderDrawer();
      expect(screen.getByLabelText(/^name$/i)).toBeInTheDocument();
    });

    it("renders the Username input", () => {
      renderDrawer();
      expect(screen.getByLabelText(/^username$/i)).toBeInTheDocument();
    });

    it("renders the Email input", () => {
      renderDrawer();
      expect(screen.getByLabelText(/^email$/i)).toBeInTheDocument();
    });

    it("renders the Password input", () => {
      renderDrawer();
      expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument();
    });

    it("renders the 'Save Changes' submit button", () => {
      renderDrawer();
      expect(
        screen.getByRole("button", { name: /save changes/i }),
      ).toBeInTheDocument();
    });

    it("renders the Cancel button", () => {
      renderDrawer();
      expect(
        screen.getByRole("button", { name: /cancel/i }),
      ).toBeInTheDocument();
    });
  });

  describe("form pre-filling", () => {
    it("pre-fills the Name field with the user's name", () => {
      renderDrawer();
      expect(screen.getByLabelText(/^name$/i)).toHaveValue("Alice Smith");
    });

    it("pre-fills the Username field with the user's username", () => {
      renderDrawer();
      expect(screen.getByLabelText(/^username$/i)).toHaveValue("alice");
    });

    it("pre-fills the Email field with the user's email", () => {
      renderDrawer();
      expect(screen.getByLabelText(/^email$/i)).toHaveValue(
        "alice@example.com",
      );
    });

    it("leaves the Password field blank", () => {
      renderDrawer();
      expect(screen.getByLabelText(/^password$/i)).toHaveValue("");
    });

    it("pre-fills confirmed checkbox as unchecked when user is not confirmed", () => {
      renderDrawer({ user: mockUser });
      expect(screen.getByLabelText(/^confirmed$/i)).not.toBeChecked();
    });

    it("pre-fills admin checkbox as unchecked when user is not admin", () => {
      renderDrawer({ user: mockUser });
      expect(screen.getByLabelText(/^admin user$/i)).not.toBeChecked();
    });
  });

  describe("form enabling", () => {
    it("submit button is enabled when all required fields are filled", () => {
      renderDrawer();
      expect(
        screen.getByRole("button", { name: /save changes/i }),
      ).not.toBeDisabled();
    });

    it("disables submit button when name is cleared", async () => {
      renderDrawer();
      const nameInput = screen.getByLabelText(/^name$/i);
      await userEvent.clear(nameInput);
      expect(
        screen.getByRole("button", { name: /save changes/i }),
      ).toBeDisabled();
    });

    it("disables submit button when username is cleared", async () => {
      renderDrawer();
      await userEvent.clear(screen.getByLabelText(/^username$/i));
      expect(
        screen.getByRole("button", { name: /save changes/i }),
      ).toBeDisabled();
    });

    it("disables submit button when email is cleared", async () => {
      renderDrawer();
      await userEvent.clear(screen.getByLabelText(/^email$/i));
      expect(
        screen.getByRole("button", { name: /save changes/i }),
      ).toBeDisabled();
    });
  });

  describe("confirmed checkbox constraint", () => {
    it("confirmed checkbox is disabled for an already-confirmed user", () => {
      renderDrawer({ user: confirmedUser });
      expect(screen.getByLabelText(/^confirmed$/i)).toBeDisabled();
    });

    it("confirmed checkbox is enabled for a pending (unconfirmed) user", () => {
      renderDrawer({ user: mockUser });
      expect(screen.getByLabelText(/^confirmed$/i)).not.toBeDisabled();
    });

    it("recognises status='confirmed' as a confirmed user", () => {
      const userWithStatus: EditableUser = {
        ...mockUser,
        status: "confirmed",
      };
      renderDrawer({ user: userWithStatus });
      expect(screen.getByLabelText(/^confirmed$/i)).toBeDisabled();
    });
  });

  describe("admin checkbox — self-demotion constraint", () => {
    it("admin checkbox is disabled when editing your own admin account", () => {
      useAuthStore.setState({ username: "alice" } as never);
      const selfAdmin: EditableUser = { ...mockUser, admin: true };
      renderDrawer({ user: selfAdmin });
      expect(screen.getByLabelText(/^admin user$/i)).toBeDisabled();
    });

    it("admin checkbox is enabled when editing another admin", () => {
      useAuthStore.setState({ username: "admin" } as never);
      const otherAdmin: EditableUser = {
        ...mockUser,
        username: "bob",
        admin: true,
      };
      renderDrawer({ user: otherAdmin });
      expect(screen.getByLabelText(/^admin user$/i)).not.toBeDisabled();
    });

    it("admin checkbox is enabled for a non-admin self user", () => {
      useAuthStore.setState({ username: "alice" } as never);
      const selfNonAdmin: EditableUser = { ...mockUser, admin: false };
      renderDrawer({ user: selfNonAdmin });
      expect(screen.getByLabelText(/^admin user$/i)).not.toBeDisabled();
    });
  });

  describe("password visibility toggle", () => {
    it("shows password in plaintext when Show password button is clicked", async () => {
      renderDrawer();
      await userEvent.click(
        screen.getByRole("button", { name: /show password/i }),
      );
      expect(screen.getByLabelText(/^password$/i)).toHaveAttribute(
        "type",
        "text",
      );
    });

    it("hides password again when Hide password is clicked", async () => {
      renderDrawer();
      await userEvent.click(
        screen.getByRole("button", { name: /show password/i }),
      );
      await userEvent.click(
        screen.getByRole("button", { name: /hide password/i }),
      );
      expect(screen.getByLabelText(/^password$/i)).toHaveAttribute(
        "type",
        "password",
      );
    });
  });

  describe("namespace limit controls", () => {
    it("does not show namespace sub-options by default when max_namespaces is undefined", () => {
      renderDrawer({ user: { ...mockUser, max_namespaces: undefined } });
      expect(
        screen.queryByLabelText(/disable namespace creation/i),
      ).not.toBeInTheDocument();
    });

    it("pre-enables namespace limit when max_namespaces is set", () => {
      renderDrawer({ user: { ...mockUser, max_namespaces: 5 } });
      expect(screen.getByLabelText(/max namespaces/i)).toBeInTheDocument();
    });

    it("pre-checks disable namespace creation when max_namespaces is 0", () => {
      renderDrawer({ user: { ...mockUser, max_namespaces: 0 } });
      expect(
        screen.getByLabelText(/disable namespace creation/i),
      ).toBeChecked();
    });

    it("hides max namespaces input when disable is checked", async () => {
      renderDrawer({ user: { ...mockUser, max_namespaces: 5 } });
      await userEvent.click(
        screen.getByLabelText(/disable namespace creation/i),
      );
      expect(
        screen.queryByLabelText(/max namespaces/i),
      ).not.toBeInTheDocument();
    });
  });

  describe("submit — success", () => {
    it("calls mutateAsync with the correct payload", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      renderDrawer();

      const nameInput = screen.getByLabelText(/^name$/i);
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, "Alice Updated");

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalledWith({
          path: { id: "u1" },
          body: expect.objectContaining({
            name: "Alice Updated",
            username: "alice",
            email: "alice@example.com",
          }),
        });
      });
    });

    it("calls onClose after successful update", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      const { onClose } = renderDrawer();

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    });

    it("sends max_namespaces as undefined when limit is not enabled", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      renderDrawer({ user: { ...mockUser, max_namespaces: undefined } });

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalledWith({
          path: { id: "u1" },
          body: expect.objectContaining({ max_namespaces: undefined }),
        });
      });
    });
  });

  describe("submit — error handling", () => {
    it("shows conflict error message for 409 responses", async () => {
      mockMutateAsync.mockRejectedValue({ status: 409 });
      renderDrawer();

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(screen.getByText(/already exists/i)).toBeInTheDocument();
      });
    });

    it("shows generic error for 400 responses", async () => {
      mockMutateAsync.mockRejectedValue({ status: 400 });
      renderDrawer();

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(screen.getByText(/failed to update user/i)).toBeInTheDocument();
      });
    });

    it("shows generic error for unexpected failures", async () => {
      mockMutateAsync.mockRejectedValue(new Error("server error"));
      renderDrawer();

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(screen.getByText(/failed to update user/i)).toBeInTheDocument();
      });
    });

    it("renders error with role='alert'", async () => {
      mockMutateAsync.mockRejectedValue(new Error("server error"));
      renderDrawer();

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
    });

    it("does not call onClose when update fails", async () => {
      mockMutateAsync.mockRejectedValue(new Error("server error"));
      const { onClose } = renderDrawer();

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => screen.getByRole("alert"));
      expect(onClose).not.toHaveBeenCalled();
    });
  });

  describe("cancel", () => {
    it("calls onClose when Cancel is clicked", async () => {
      const { onClose } = renderDrawer();
      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(onClose).toHaveBeenCalledTimes(1);
    });

    it("does not call mutateAsync when Cancel is clicked", async () => {
      renderDrawer();
      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(mockMutateAsync).not.toHaveBeenCalled();
    });
  });

  describe("state reset on reopen", () => {
    it("reloads user data when drawer is closed then reopened", async () => {
      const { rerender } = renderDrawer({ user: mockUser });

      const nameInput = screen.getByLabelText(/^name$/i);
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, "Changed Name");

      rerender(
        <EditUserDrawer open={false} onClose={vi.fn()} user={mockUser} />,
      );
      rerender(
        <EditUserDrawer open={true} onClose={vi.fn()} user={mockUser} />,
      );

      expect(screen.getByLabelText(/^name$/i)).toHaveValue("Alice Smith");
    });

    it("clears any error when closed then reopened", async () => {
      mockMutateAsync.mockRejectedValue(new Error("fail"));
      const { rerender } = renderDrawer({ user: mockUser });

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );
      await waitFor(() => screen.getByRole("alert"));

      rerender(
        <EditUserDrawer open={false} onClose={vi.fn()} user={mockUser} />,
      );
      rerender(
        <EditUserDrawer open={true} onClose={vi.fn()} user={mockUser} />,
      );

      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });

  describe("null user", () => {
    it("renders the drawer with empty fields when user is null", () => {
      renderDrawer({ user: null });
      expect(screen.getByLabelText(/^name$/i)).toHaveValue("");
      expect(screen.getByLabelText(/^username$/i)).toHaveValue("");
      expect(screen.getByLabelText(/^email$/i)).toHaveValue("");
    });
  });
});
