import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { useAuthStore } from "@/stores/authStore";
import type { UserAdminResponse } from "@/client";
import EditUserModal from "../EditUserModal";

vi.mock("@/components/common/Modal", async () => ({
  default: (await import("@/tests/mocks")).MockModal,
}));

const Wrapper = createTestWrapper();

const mockUser: UserAdminResponse = {
  id: "u1",
  name: "Alice Smith",
  username: "alice",
  email: "alice@example.com",
  admin: false,
  status: "not-confirmed",
  created_at: "2024-01-01T00:00:00Z",
  last_login: "2024-01-01T00:00:00Z",
};

const confirmedUser: UserAdminResponse = {
  ...mockUser,
  status: "confirmed",
};

const updateSpy = vi.fn();

function renderModal(
  overrides: Partial<{
    open: boolean;
    onClose: () => void;
    user: UserAdminResponse | null;
  }> = {},
) {
  const defaults = { open: true, onClose: vi.fn(), user: mockUser };
  const props = { ...defaults, ...overrides };
  return {
    onClose: props.onClose,
    ...render(<EditUserModal {...props} />, { wrapper: Wrapper }),
  };
}

describe("EditUserModal", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    updateSpy.mockReset();
    useAuthStore.setState({ userId: "u2" });
    server.use(
      http.put("*/admin/api/users/:id", async ({ request, params }) => {
        updateSpy({
          path: { id: params.id },
          body: await request.json(),
        });
        return HttpResponse.json({});
      }),
    );
  });

  describe("form enabling", () => {
    it("submit button is enabled when all required fields are filled", async () => {
      renderModal();
      await waitFor(() =>
        expect(
          screen.getByRole("button", { name: /save changes/i }),
        ).not.toBeDisabled(),
      );
    });

    it("disables submit button when name is cleared", async () => {
      renderModal();
      const nameInput = screen.getByLabelText(/^name$/i);
      await userEvent.clear(nameInput);
      expect(
        screen.getByRole("button", { name: /save changes/i }),
      ).toBeDisabled();
    });

    it("disables submit button when username is cleared", async () => {
      renderModal();
      await userEvent.clear(screen.getByLabelText(/^username$/i));
      expect(
        screen.getByRole("button", { name: /save changes/i }),
      ).toBeDisabled();
    });

    it("disables submit button when email is cleared", async () => {
      renderModal();
      await userEvent.clear(screen.getByLabelText(/^email$/i));
      expect(
        screen.getByRole("button", { name: /save changes/i }),
      ).toBeDisabled();
    });
  });

  describe("confirmed checkbox constraint", () => {
    it("confirmed checkbox is disabled for an already-confirmed user", () => {
      renderModal({ user: confirmedUser });
      expect(screen.getByLabelText(/^confirmed$/i)).toBeDisabled();
    });

    it("confirmed checkbox is enabled for a pending (unconfirmed) user", () => {
      renderModal({ user: mockUser });
      expect(screen.getByLabelText(/^confirmed$/i)).not.toBeDisabled();
    });

    it("recognises status='confirmed' as a confirmed user", () => {
      const userWithStatus: UserAdminResponse = {
        ...mockUser,
        status: "confirmed",
      };
      renderModal({ user: userWithStatus });
      expect(screen.getByLabelText(/^confirmed$/i)).toBeDisabled();
    });
  });

  describe("admin checkbox — self-demotion constraint", () => {
    it("admin checkbox is disabled when editing your own admin account", () => {
      useAuthStore.setState({ userId: "u1" });
      const selfAdmin: UserAdminResponse = { ...mockUser, admin: true };
      renderModal({ user: selfAdmin });
      expect(screen.getByLabelText(/^admin user$/i)).toBeDisabled();
    });

    it("admin checkbox is enabled when editing another admin", () => {
      useAuthStore.setState({ userId: "u2" });
      const otherAdmin: UserAdminResponse = { ...mockUser, admin: true };
      renderModal({ user: otherAdmin });
      expect(screen.getByLabelText(/^admin user$/i)).not.toBeDisabled();
    });

    it("admin checkbox is enabled for a non-admin self user", () => {
      useAuthStore.setState({ userId: "u1" });
      const selfNonAdmin: UserAdminResponse = { ...mockUser, admin: false };
      renderModal({ user: selfNonAdmin });
      expect(screen.getByLabelText(/^admin user$/i)).not.toBeDisabled();
    });
  });

  describe("namespace limit controls", () => {
    it("pre-enables namespace limit when max_namespaces is set", () => {
      renderModal({ user: { ...mockUser, max_namespaces: 5 } });
      expect(screen.getByLabelText(/max namespaces/i)).toBeInTheDocument();
    });

    it("pre-checks disable namespace creation when max_namespaces is 0", () => {
      renderModal({ user: { ...mockUser, max_namespaces: 0 } });
      expect(
        screen.getByLabelText(/disable namespace creation/i),
      ).toBeChecked();
    });

    it("hides max namespaces input when disable is checked", async () => {
      renderModal({ user: { ...mockUser, max_namespaces: 5 } });
      await userEvent.click(
        screen.getByLabelText(/disable namespace creation/i),
      );
      expect(
        screen.queryByLabelText(/max namespaces/i),
      ).not.toBeInTheDocument();
    });
  });

  describe("submit — success", () => {
    it("calls adminUpdateUser with the correct payload", async () => {
      renderModal();

      const nameInput = screen.getByLabelText(/^name$/i);
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, "Alice Updated");

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(updateSpy).toHaveBeenCalledWith(
          expect.objectContaining({
            path: { id: "u1" },
            body: expect.objectContaining({
              name: "Alice Updated",
              username: "alice",
              email: "alice@example.com",
            }),
          }),
        );
      });
    });

    it("calls onClose after successful update", async () => {
      const { onClose } = renderModal();

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    });

    it("sends max_namespaces as undefined when limit is not enabled", async () => {
      renderModal({ user: { ...mockUser, max_namespaces: undefined } });

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(updateSpy).toHaveBeenCalledWith(
          expect.objectContaining({
            path: { id: "u1" },
          }),
        );
      });
    });
  });

  describe("submit — error handling", () => {
    it("shows conflict error message for 409 responses", async () => {
      server.use(
        http.put("*/admin/api/users/:id", () =>
          HttpResponse.json({}, { status: 409 }),
        ),
      );
      renderModal();

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(screen.getByText(/already exists/i)).toBeInTheDocument();
      });
    });

    it("shows generic error for 400 responses", async () => {
      server.use(
        http.put("*/admin/api/users/:id", () =>
          HttpResponse.json({}, { status: 400 }),
        ),
      );
      renderModal();

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(screen.getByText(/failed to update user/i)).toBeInTheDocument();
      });
    });

    it("shows generic error for unexpected failures", async () => {
      server.use(
        http.put("*/admin/api/users/:id", () => HttpResponse.error()),
      );
      renderModal();

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(screen.getByText(/failed to update user/i)).toBeInTheDocument();
      });
    });
  });

  describe("client-side validation", () => {
    it("allows submit with a blank password (password kept as-is)", async () => {
      renderModal();
      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );
      await waitFor(() => expect(updateSpy).toHaveBeenCalled());
    });

    it("rejects a too-short password on edit when the user is changing it", async () => {
      renderModal();
      const passwordInput = screen.getByLabelText(/^password$/i);
      await userEvent.type(passwordInput, "abc");
      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      expect(updateSpy).not.toHaveBeenCalled();
      expect(passwordInput).toHaveAttribute("aria-invalid", "true");
    });

    it("blocks submit when an existing field is edited to an invalid value", async () => {
      renderModal();
      const usernameInput = screen.getByLabelText(/^username$/i);
      await userEvent.clear(usernameInput);
      await userEvent.type(usernameInput, "Invalid Username!");

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );
      expect(updateSpy).not.toHaveBeenCalled();
      expect(usernameInput).toHaveAttribute("aria-invalid", "true");
    });
  });

  describe("state reset on reopen", () => {
    it("reloads user data when modal is closed then reopened", async () => {
      const { rerender } = renderModal({ user: mockUser });

      const nameInput = screen.getByLabelText(/^name$/i);
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, "Changed Name");

      rerender(
        <EditUserModal open={false} onClose={vi.fn()} user={mockUser} />,
      );
      rerender(
        <EditUserModal open={true} onClose={vi.fn()} user={mockUser} />,
      );

      expect(screen.getByLabelText(/^name$/i)).toHaveValue("Alice Smith");
    });

    it("clears any error when closed then reopened", async () => {
      server.use(
        http.put("*/admin/api/users/:id", () => HttpResponse.error()),
      );
      const { rerender } = renderModal({ user: mockUser });

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );
      await waitFor(() => screen.getByRole("alert"));

      rerender(
        <EditUserModal open={false} onClose={vi.fn()} user={mockUser} />,
      );
      rerender(
        <EditUserModal open={true} onClose={vi.fn()} user={mockUser} />,
      );

      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });
});
