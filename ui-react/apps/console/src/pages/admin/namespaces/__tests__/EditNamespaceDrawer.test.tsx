import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import EditNamespaceDrawer from "../EditNamespaceDrawer";
import { useAdminEditNamespace } from "@/hooks/useAdminNamespaceMutations";
import type { Namespace } from "@/client";

vi.mock("@/hooks/useAdminNamespaceMutations", () => ({
  useAdminEditNamespace: vi.fn(),
}));

vi.mock("@/utils/styles", () => ({
  LABEL: "label",
  INPUT: "input",
}));

vi.mock("@/components/common/Drawer", async () => ({
  default: (await import("@/pages/admin/users/__tests__/mocks")).MockDrawer,
}));

const mockMutateAsync = vi.fn();

const mockNamespace: Namespace = {
  name: "my-namespace",
  owner: "owner-1",
  tenant_id: "tenant-abc",
  members: [],
  settings: {
    session_record: true,
    connection_announcement: "hello",
    device_auto_accept: false,
  },
  max_devices: 10,
  created_at: "2024-01-01T00:00:00Z",
  billing: null,
  devices_pending_count: 0,
  devices_accepted_count: 3,
  devices_rejected_count: 0,
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(useAdminEditNamespace).mockReturnValue({
    mutateAsync: mockMutateAsync,
    isPending: false,
  } as never);
});

function renderDrawer(
  overrides: Partial<{
    open: boolean;
    onClose: () => void;
    namespace: Namespace | null;
  }> = {},
) {
  const defaults = { open: true, onClose: vi.fn(), namespace: mockNamespace };
  const props = { ...defaults, ...overrides };
  return {
    onClose: props.onClose,
    ...render(<EditNamespaceDrawer {...props} />),
  };
}

describe("EditNamespaceDrawer", () => {
  describe("rendering — closed", () => {
    it("renders nothing when open is false", () => {
      renderDrawer({ open: false });
      expect(screen.queryByText("Edit Namespace")).not.toBeInTheDocument();
    });
  });

  describe("rendering — open", () => {
    it("renders the 'Edit Namespace' title", () => {
      renderDrawer();
      expect(screen.getByText("Edit Namespace")).toBeInTheDocument();
    });

    it("renders the Name input", () => {
      renderDrawer();
      expect(screen.getByLabelText(/^name$/i)).toBeInTheDocument();
    });

    it("renders the Max Devices input", () => {
      renderDrawer();
      expect(screen.getByLabelText(/^max devices$/i)).toBeInTheDocument();
    });

    it("renders the Session Recording checkbox", () => {
      renderDrawer();
      expect(screen.getByLabelText(/session recording/i)).toBeInTheDocument();
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
    it("pre-fills the Name field with the namespace name", () => {
      renderDrawer();
      expect(screen.getByLabelText(/^name$/i)).toHaveValue("my-namespace");
    });

    it("pre-fills the Max Devices field with the namespace max_devices", () => {
      renderDrawer();
      expect(screen.getByLabelText(/^max devices$/i)).toHaveValue(10);
    });

    it("pre-fills Session Recording checkbox as checked when session_record is true", () => {
      renderDrawer();
      expect(screen.getByLabelText(/session recording/i)).toBeChecked();
    });

    it("pre-fills Session Recording checkbox as unchecked when session_record is false", () => {
      renderDrawer({
        namespace: {
          ...mockNamespace,
          settings: { ...mockNamespace.settings!, session_record: false },
        },
      });
      expect(screen.getByLabelText(/session recording/i)).not.toBeChecked();
    });

    it("uses default max_devices of -1 when namespace has no max_devices", () => {
      renderDrawer({
        namespace: { ...mockNamespace, max_devices: -1 },
      });
      expect(screen.getByLabelText(/^max devices$/i)).toHaveValue(-1);
    });
  });

  describe("form enabling", () => {
    it("submit button is enabled when name is non-empty", () => {
      renderDrawer();
      expect(
        screen.getByRole("button", { name: /save changes/i }),
      ).not.toBeDisabled();
    });

    it("disables submit button when name is cleared", async () => {
      renderDrawer();
      await userEvent.clear(screen.getByLabelText(/^name$/i));
      expect(
        screen.getByRole("button", { name: /save changes/i }),
      ).toBeDisabled();
    });

    it("re-enables submit button when name is typed back in", async () => {
      renderDrawer();
      const nameInput = screen.getByLabelText(/^name$/i);
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, "new-name");
      expect(
        screen.getByRole("button", { name: /save changes/i }),
      ).not.toBeDisabled();
    });
  });

  describe("submit — success", () => {
    it("calls mutateAsync with the correct payload", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      renderDrawer();

      const nameInput = screen.getByLabelText(/^name$/i);
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, "updated-namespace");

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalledWith({
          path: { tenantID: "tenant-abc" },
          body: expect.objectContaining({
            name: "updated-namespace",
            max_devices: 10,
            settings: expect.objectContaining({
              session_record: true,
            }),
          }),
        });
      });
    });

    it("spreads the original namespace fields into the body", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      renderDrawer();

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({
              owner: "owner-1",
              tenant_id: "tenant-abc",
            }),
          }),
        );
      });
    });

    it("passes the updated session_record value when checkbox is toggled", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      renderDrawer();

      await userEvent.click(screen.getByLabelText(/session recording/i));

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({
              settings: expect.objectContaining({ session_record: false }),
            }),
          }),
        );
      });
    });

    it("calls onClose after successful submit", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      const { onClose } = renderDrawer();

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
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
        expect(
          screen.getByText("A namespace with this name already exists."),
        ).toBeInTheDocument();
      });
    });

    it("shows generic error for non-409 SDK errors", async () => {
      mockMutateAsync.mockRejectedValue({ status: 500 });
      renderDrawer();

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(
          screen.getByText(/failed to update namespace/i),
        ).toBeInTheDocument();
      });
    });

    it("shows generic error for non-SDK errors", async () => {
      mockMutateAsync.mockRejectedValue(new Error("network error"));
      renderDrawer();

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(
          screen.getByText(/failed to update namespace/i),
        ).toBeInTheDocument();
      });
    });

    it("renders error with role='alert'", async () => {
      mockMutateAsync.mockRejectedValue(new Error("network error"));
      renderDrawer();

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
    });

    it("does not call onClose when update fails", async () => {
      mockMutateAsync.mockRejectedValue(new Error("network error"));
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
    it("reloads namespace data when drawer is closed then reopened", async () => {
      const { rerender } = renderDrawer({ namespace: mockNamespace });

      const nameInput = screen.getByLabelText(/^name$/i);
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, "changed-name");

      rerender(
        <EditNamespaceDrawer
          open={false}
          onClose={vi.fn()}
          namespace={mockNamespace}
        />,
      );
      rerender(
        <EditNamespaceDrawer
          open={true}
          onClose={vi.fn()}
          namespace={mockNamespace}
        />,
      );

      expect(screen.getByLabelText(/^name$/i)).toHaveValue("my-namespace");
    });

    it("clears any error when closed then reopened", async () => {
      mockMutateAsync.mockRejectedValue(new Error("fail"));
      const { rerender } = renderDrawer({ namespace: mockNamespace });

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );
      await waitFor(() => screen.getByRole("alert"));

      rerender(
        <EditNamespaceDrawer
          open={false}
          onClose={vi.fn()}
          namespace={mockNamespace}
        />,
      );
      rerender(
        <EditNamespaceDrawer
          open={true}
          onClose={vi.fn()}
          namespace={mockNamespace}
        />,
      );

      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });

  describe("null namespace", () => {
    it("renders the drawer with empty name field when namespace is null", () => {
      renderDrawer({ namespace: null });
      expect(screen.getByLabelText(/^name$/i)).toHaveValue("");
    });

    it("renders the drawer with max_devices of -1 when namespace is null", () => {
      renderDrawer({ namespace: null });
      expect(screen.getByLabelText(/^max devices$/i)).toHaveValue(-1);
    });

    it("renders the Session Recording checkbox unchecked when namespace is null", () => {
      renderDrawer({ namespace: null });
      expect(screen.getByLabelText(/session recording/i)).not.toBeChecked();
    });
  });
});
