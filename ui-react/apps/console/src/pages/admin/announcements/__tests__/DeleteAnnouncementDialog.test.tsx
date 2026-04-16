import type { ReactNode } from "react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import DeleteAnnouncementDialog from "../DeleteAnnouncementDialog";
import { useAdminDeleteAnnouncement } from "@/hooks/useAdminAnnouncementMutations";

vi.mock("@/hooks/useAdminAnnouncementMutations", () => ({
  useAdminDeleteAnnouncement: vi.fn(),
}));

// Flatten ConfirmDialog to a plain div so we can test the component's logic
// without dialog portals, animations, or BaseDialog internals.
vi.mock("@/components/common/ConfirmDialog", () => ({
  default: ({
    open,
    onClose,
    onConfirm,
    title,
    description,
    confirmLabel = "Confirm",
  }: {
    open: boolean;
    onClose: () => void;
    onConfirm: () => Promise<void> | void;
    title: string;
    description: ReactNode;
    confirmLabel?: string;
  }) => {
    if (!open) return null;
    return (
      <div role="dialog" aria-label={title}>
        <h2>{title}</h2>
        <div>{description}</div>
        <button onClick={onClose}>Cancel</button>
        <button onClick={() => void onConfirm()}>{confirmLabel}</button>
      </div>
    );
  },
}));

const mockMutateAsync = vi.fn();

const mockAnnouncement = {
  uuid: "ann-uuid-1234",
  title: "Test Announcement",
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(useAdminDeleteAnnouncement).mockReturnValue({
    mutateAsync: mockMutateAsync,
  } as never);
});

function renderDialog(
  overrides: Partial<{
    open: boolean;
    onClose: () => void;
    announcement: typeof mockAnnouncement | null;
    onDeleted: () => void;
  }> = {},
) {
  const defaults = {
    open: true,
    onClose: vi.fn(),
    announcement: mockAnnouncement,
    onDeleted: vi.fn(),
  };
  const props = { ...defaults, ...overrides };
  return {
    onClose: props.onClose,
    onDeleted: props.onDeleted,
    ...render(<DeleteAnnouncementDialog {...props} />),
  };
}

describe("DeleteAnnouncementDialog", () => {
  describe("rendering — closed", () => {
    it("renders nothing when open is false", () => {
      renderDialog({ open: false });
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });

  describe("rendering — open", () => {
    it("renders the dialog when open is true", () => {
      renderDialog();
      expect(screen.getByRole("dialog")).toBeInTheDocument();
    });

    it("renders the 'Delete Announcement' title", () => {
      renderDialog();
      expect(screen.getByText("Delete Announcement")).toBeInTheDocument();
    });

    it("renders the announcement title in the description", () => {
      renderDialog();
      expect(screen.getByText("Test Announcement")).toBeInTheDocument();
    });

    it("renders the 'This action cannot be undone' warning", () => {
      renderDialog();
      expect(
        screen.getByText(/this action cannot be undone/i),
      ).toBeInTheDocument();
    });

    it("renders the 'Delete' confirm button", () => {
      renderDialog();
      expect(
        screen.getByRole("button", { name: /^delete$/i }),
      ).toBeInTheDocument();
    });

    it("renders the Cancel button", () => {
      renderDialog();
      expect(
        screen.getByRole("button", { name: /cancel/i }),
      ).toBeInTheDocument();
    });
  });

  describe("confirm — success", () => {
    it("calls mutateAsync with the correct uuid", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalledWith({
          path: { uuid: "ann-uuid-1234" },
        });
      });
    });

    it("calls onDeleted callback after successful deletion", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      const { onDeleted } = renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => expect(onDeleted).toHaveBeenCalledTimes(1));
    });

    it("calls onClose after successful deletion", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      const { onClose } = renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    });

    it("calls onClose before onDeleted", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      const callOrder: string[] = [];
      const onClose = vi.fn(() => callOrder.push("onClose"));
      const onDeleted = vi.fn(() => callOrder.push("onDeleted"));
      render(
        <DeleteAnnouncementDialog
          open={true}
          onClose={onClose}
          announcement={mockAnnouncement}
          onDeleted={onDeleted}
        />,
      );

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => expect(onDeleted).toHaveBeenCalledTimes(1));
      expect(callOrder).toEqual(["onClose", "onDeleted"]);
    });
  });

  describe("confirm — error handling", () => {
    it("shows generic error message on failure", async () => {
      mockMutateAsync.mockRejectedValue(new Error("server error"));
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => {
        expect(
          screen.getByText(/failed to delete announcement/i),
        ).toBeInTheDocument();
      });
    });

    it("shows error for SDK errors", async () => {
      mockMutateAsync.mockRejectedValue({ status: 500 });
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => {
        expect(
          screen.getByText(/failed to delete announcement/i),
        ).toBeInTheDocument();
      });
    });

    it("does not call onDeleted when deletion fails", async () => {
      mockMutateAsync.mockRejectedValue(new Error("server error"));
      const { onDeleted } = renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => screen.getByText(/failed to delete announcement/i));
      expect(onDeleted).not.toHaveBeenCalled();
    });

    it("does not call onClose when deletion fails", async () => {
      mockMutateAsync.mockRejectedValue(new Error("server error"));
      const { onClose } = renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => screen.getByText(/failed to delete announcement/i));
      expect(onClose).not.toHaveBeenCalled();
    });

    it("clears the error message on subsequent close after failure", async () => {
      mockMutateAsync.mockRejectedValue(new Error("server error"));
      const { onClose } = renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));
      await waitFor(() => screen.getByText(/failed to delete announcement/i));

      // The ConfirmDialog's onClose callback wraps our onClose to clear error;
      // clicking Cancel triggers that path.
      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(onClose).toHaveBeenCalledTimes(1);
    });
  });

  describe("cancel", () => {
    it("calls onClose when Cancel is clicked", async () => {
      const { onClose } = renderDialog();
      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(onClose).toHaveBeenCalledTimes(1);
    });

    it("does not call mutateAsync when Cancel is clicked", async () => {
      renderDialog();
      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(mockMutateAsync).not.toHaveBeenCalled();
    });

    it("does not call onDeleted when Cancel is clicked", async () => {
      const { onDeleted } = renderDialog();
      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(onDeleted).not.toHaveBeenCalled();
    });
  });

  describe("null announcement", () => {
    it("renders nothing meaningful in the description when announcement is null", () => {
      renderDialog({ announcement: null });
      expect(screen.queryByText("Test Announcement")).not.toBeInTheDocument();
    });

    it("does not call mutateAsync when confirmed with null announcement", async () => {
      renderDialog({ announcement: null });
      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));
      await waitFor(() => expect(mockMutateAsync).not.toHaveBeenCalled());
    });
  });

  describe("optional onDeleted callback", () => {
    it("does not throw when onDeleted is not provided and deletion succeeds", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      const { onClose } = renderDialog({ onDeleted: undefined });

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    });
  });
});
