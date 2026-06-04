import type { ReactNode } from "react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

// ── Module mocks ──────────────────────────────────────────────────────────────

vi.mock("@/hooks/useContainerMutations", () => ({
  useUpdateContainerStatus: vi.fn(),
  useRemoveContainer: vi.fn(),
}));

// Flatten ConfirmDialog to a plain div so we can exercise the component's
// logic without animations, portals, or BaseDialog internals.
// The mock renders children so the error <p role="alert"> is visible.
vi.mock("@/components/common/ConfirmDialog", () => ({
  default: ({
    open,
    onClose,
    onConfirm,
    title,
    description,
    confirmLabel = "Confirm",
    children,
  }: {
    open: boolean;
    onClose: () => void;
    onConfirm: () => Promise<void> | void;
    title: string;
    description: ReactNode;
    confirmLabel?: string;
    children?: ReactNode;
  }) => {
    if (!open) return null;
    return (
      <div role="dialog">
        <h2>{title}</h2>
        <div>{description}</div>
        {children}
        <button onClick={onClose}>Cancel</button>
        <button onClick={() => void onConfirm()}>{confirmLabel}</button>
      </div>
    );
  },
}));

// ── Imports (after mocks) ─────────────────────────────────────────────────────

import {
  useUpdateContainerStatus,
  useRemoveContainer,
} from "@/hooks/useContainerMutations";
import ContainerActionDialog from "../ContainerActionDialog";

// ── Helpers ───────────────────────────────────────────────────────────────────

const mockStatusMutateAsync = vi.fn();
const mockRemoveMutateAsync = vi.fn();

const mockContainer = { uid: "container-uid-1", name: "my-container" };

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(useUpdateContainerStatus).mockReturnValue({
    mutateAsync: mockStatusMutateAsync,
  } as never);
  vi.mocked(useRemoveContainer).mockReturnValue({
    mutateAsync: mockRemoveMutateAsync,
  } as never);
});

type Action = "accept" | "reject" | "remove";

function renderDialog(
  overrides: Partial<{
    container: typeof mockContainer | null;
    action: Action;
    open: boolean;
    onClose: () => void;
    onSuccess: () => void;
  }> = {},
) {
  const defaults = {
    container: mockContainer,
    action: "accept" as Action,
    open: true,
    onClose: vi.fn(),
    onSuccess: vi.fn(),
  };
  const props = { ...defaults, ...overrides };
  return {
    onClose: props.onClose,
    onSuccess: props.onSuccess,
    ...render(<ContainerActionDialog {...props} />),
  };
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("ContainerActionDialog", () => {
  describe("rendering — closed", () => {
    it("renders nothing when open is false", () => {
      renderDialog({ open: false });
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });

  describe("rendering — title and confirm label per action", () => {
    it("shows 'Accept Container' title for accept action", () => {
      renderDialog({ action: "accept" });
      expect(screen.getByText("Accept Container")).toBeInTheDocument();
    });

    it("shows 'Accept' confirm button for accept action", () => {
      renderDialog({ action: "accept" });
      expect(
        screen.getByRole("button", { name: /^accept$/i }),
      ).toBeInTheDocument();
    });

    it("shows 'Reject Container' title for reject action", () => {
      renderDialog({ action: "reject" });
      expect(screen.getByText("Reject Container")).toBeInTheDocument();
    });

    it("shows 'Reject' confirm button for reject action", () => {
      renderDialog({ action: "reject" });
      expect(
        screen.getByRole("button", { name: /^reject$/i }),
      ).toBeInTheDocument();
    });

    it("shows 'Remove Container' title for remove action", () => {
      renderDialog({ action: "remove" });
      expect(screen.getByText("Remove Container")).toBeInTheDocument();
    });

    it("shows 'Remove' confirm button for remove action", () => {
      renderDialog({ action: "remove" });
      expect(
        screen.getByRole("button", { name: /^remove$/i }),
      ).toBeInTheDocument();
    });
  });

  describe("rendering — undone warning", () => {
    it("shows 'This action cannot be undone' only for remove action", () => {
      renderDialog({ action: "remove" });
      expect(
        screen.getByText(/this action cannot be undone/i),
      ).toBeInTheDocument();
    });

    it("does not show undone warning for accept action", () => {
      renderDialog({ action: "accept" });
      expect(
        screen.queryByText(/this action cannot be undone/i),
      ).not.toBeInTheDocument();
    });

    it("does not show undone warning for reject action", () => {
      renderDialog({ action: "reject" });
      expect(
        screen.queryByText(/this action cannot be undone/i),
      ).not.toBeInTheDocument();
    });
  });

  describe("accept — success", () => {
    it("calls updateContainerStatus with uid and status 'accept'", async () => {
      mockStatusMutateAsync.mockResolvedValue(undefined);
      renderDialog({ action: "accept" });

      await userEvent.click(screen.getByRole("button", { name: /^accept$/i }));

      await waitFor(() => {
        expect(mockStatusMutateAsync).toHaveBeenCalledWith({
          path: { uid: "container-uid-1", status: "accept" },
        });
      });
    });

    it("calls onClose after successful accept", async () => {
      mockStatusMutateAsync.mockResolvedValue(undefined);
      const { onClose } = renderDialog({ action: "accept" });

      await userEvent.click(screen.getByRole("button", { name: /^accept$/i }));

      await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    });

    it("calls onSuccess after successful accept", async () => {
      mockStatusMutateAsync.mockResolvedValue(undefined);
      const { onSuccess } = renderDialog({ action: "accept" });

      await userEvent.click(screen.getByRole("button", { name: /^accept$/i }));

      await waitFor(() => expect(onSuccess).toHaveBeenCalledTimes(1));
    });
  });

  describe("reject — success", () => {
    it("calls updateContainerStatus with uid and status 'reject'", async () => {
      mockStatusMutateAsync.mockResolvedValue(undefined);
      renderDialog({ action: "reject" });

      await userEvent.click(screen.getByRole("button", { name: /^reject$/i }));

      await waitFor(() => {
        expect(mockStatusMutateAsync).toHaveBeenCalledWith({
          path: { uid: "container-uid-1", status: "reject" },
        });
      });
    });

    it("calls onClose after successful reject", async () => {
      mockStatusMutateAsync.mockResolvedValue(undefined);
      const { onClose } = renderDialog({ action: "reject" });

      await userEvent.click(screen.getByRole("button", { name: /^reject$/i }));

      await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    });

    it("calls onSuccess after successful reject", async () => {
      mockStatusMutateAsync.mockResolvedValue(undefined);
      const { onSuccess } = renderDialog({ action: "reject" });

      await userEvent.click(screen.getByRole("button", { name: /^reject$/i }));

      await waitFor(() => expect(onSuccess).toHaveBeenCalledTimes(1));
    });
  });

  describe("remove — success", () => {
    it("calls removeContainer with the container uid", async () => {
      mockRemoveMutateAsync.mockResolvedValue(undefined);
      renderDialog({ action: "remove" });

      await userEvent.click(screen.getByRole("button", { name: /^remove$/i }));

      await waitFor(() => {
        expect(mockRemoveMutateAsync).toHaveBeenCalledWith({
          path: { uid: "container-uid-1" },
        });
      });
    });

    it("calls onClose after successful remove", async () => {
      mockRemoveMutateAsync.mockResolvedValue(undefined);
      const { onClose } = renderDialog({ action: "remove" });

      await userEvent.click(screen.getByRole("button", { name: /^remove$/i }));

      await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    });

    it("calls onSuccess after successful remove", async () => {
      mockRemoveMutateAsync.mockResolvedValue(undefined);
      const { onSuccess } = renderDialog({ action: "remove" });

      await userEvent.click(screen.getByRole("button", { name: /^remove$/i }));

      await waitFor(() => expect(onSuccess).toHaveBeenCalledTimes(1));
    });
  });

  describe("error handling — accept 402", () => {
    it("shows billing error message for accept + 402", async () => {
      mockStatusMutateAsync.mockRejectedValue({ status: 402 });
      renderDialog({ action: "accept" });

      await userEvent.click(screen.getByRole("button", { name: /^accept$/i }));

      await waitFor(() => {
        expect(screen.getByRole("alert")).toHaveTextContent(
          /check your billing status/i,
        );
      });
    });

    it("does not call onClose on accept + 402", async () => {
      mockStatusMutateAsync.mockRejectedValue({ status: 402 });
      const { onClose } = renderDialog({ action: "accept" });

      await userEvent.click(screen.getByRole("button", { name: /^accept$/i }));

      await waitFor(() => screen.getByRole("alert"));
      expect(onClose).not.toHaveBeenCalled();
    });
  });

  describe("error handling — accept 403", () => {
    it("shows namespace limit error message for accept + 403", async () => {
      mockStatusMutateAsync.mockRejectedValue({ status: 403 });
      renderDialog({ action: "accept" });

      await userEvent.click(screen.getByRole("button", { name: /^accept$/i }));

      await waitFor(() => {
        expect(screen.getByRole("alert")).toHaveTextContent(
          /maximum amount of accepted containers/i,
        );
      });
    });

    it("does not call onClose on accept + 403", async () => {
      mockStatusMutateAsync.mockRejectedValue({ status: 403 });
      const { onClose } = renderDialog({ action: "accept" });

      await userEvent.click(screen.getByRole("button", { name: /^accept$/i }));

      await waitFor(() => screen.getByRole("alert"));
      expect(onClose).not.toHaveBeenCalled();
    });
  });

  describe("error handling — accept 409", () => {
    it("shows name conflict error message for accept + 409", async () => {
      mockStatusMutateAsync.mockRejectedValue({ status: 409 });
      renderDialog({ action: "accept" });

      await userEvent.click(screen.getByRole("button", { name: /^accept$/i }));

      await waitFor(() => {
        expect(screen.getByRole("alert")).toHaveTextContent(
          /a container with that name already exists/i,
        );
      });
    });

    it("does not call onClose on accept + 409", async () => {
      mockStatusMutateAsync.mockRejectedValue({ status: 409 });
      const { onClose } = renderDialog({ action: "accept" });

      await userEvent.click(screen.getByRole("button", { name: /^accept$/i }));

      await waitFor(() => screen.getByRole("alert"));
      expect(onClose).not.toHaveBeenCalled();
    });
  });

  describe("error handling — generic", () => {
    it("shows generic error for unknown status on accept", async () => {
      mockStatusMutateAsync.mockRejectedValue({ status: 500 });
      renderDialog({ action: "accept" });

      await userEvent.click(screen.getByRole("button", { name: /^accept$/i }));

      await waitFor(() => {
        expect(screen.getByRole("alert")).toHaveTextContent(
          /failed to accept container/i,
        );
      });
    });

    it("shows generic error for non-SDK errors on accept", async () => {
      mockStatusMutateAsync.mockRejectedValue(new Error("network error"));
      renderDialog({ action: "accept" });

      await userEvent.click(screen.getByRole("button", { name: /^accept$/i }));

      await waitFor(() => {
        expect(screen.getByRole("alert")).toHaveTextContent(
          /failed to accept container/i,
        );
      });
    });

    it("shows generic error for reject failures", async () => {
      mockStatusMutateAsync.mockRejectedValue({ status: 500 });
      renderDialog({ action: "reject" });

      await userEvent.click(screen.getByRole("button", { name: /^reject$/i }));

      await waitFor(() => {
        expect(screen.getByRole("alert")).toHaveTextContent(
          /failed to reject container/i,
        );
      });
    });

    it("shows generic error for remove failures", async () => {
      mockRemoveMutateAsync.mockRejectedValue(new Error("server error"));
      renderDialog({ action: "remove" });

      await userEvent.click(screen.getByRole("button", { name: /^remove$/i }));

      await waitFor(() => {
        expect(screen.getByRole("alert")).toHaveTextContent(
          /failed to remove container/i,
        );
      });
    });

    it("does not call onClose on generic error", async () => {
      mockStatusMutateAsync.mockRejectedValue(new Error("error"));
      const { onClose } = renderDialog({ action: "accept" });

      await userEvent.click(screen.getByRole("button", { name: /^accept$/i }));

      await waitFor(() => screen.getByRole("alert"));
      expect(onClose).not.toHaveBeenCalled();
    });

    it("does not call onSuccess on error", async () => {
      mockStatusMutateAsync.mockRejectedValue(new Error("error"));
      const { onSuccess } = renderDialog({ action: "accept" });

      await userEvent.click(screen.getByRole("button", { name: /^accept$/i }));

      await waitFor(() => screen.getByRole("alert"));
      expect(onSuccess).not.toHaveBeenCalled();
    });
  });

  describe("null container", () => {
    it("does not call any mutation when container is null and confirmed", async () => {
      renderDialog({ container: null, action: "accept" });

      await userEvent.click(screen.getByRole("button", { name: /^accept$/i }));

      await waitFor(() => expect(mockStatusMutateAsync).not.toHaveBeenCalled());
      expect(mockRemoveMutateAsync).not.toHaveBeenCalled();
    });
  });

  describe("cancel", () => {
    it("calls onClose when Cancel is clicked", async () => {
      const { onClose } = renderDialog();
      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(onClose).toHaveBeenCalledTimes(1);
    });

    it("does not call any mutation when Cancel is clicked", async () => {
      renderDialog();
      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(mockStatusMutateAsync).not.toHaveBeenCalled();
      expect(mockRemoveMutateAsync).not.toHaveBeenCalled();
    });
  });

  describe("optional onSuccess callback", () => {
    it("does not throw when onSuccess is not provided and action succeeds", async () => {
      mockStatusMutateAsync.mockResolvedValue(undefined);
      const { onClose } = renderDialog({
        action: "accept",
        onSuccess: undefined,
      });

      await userEvent.click(screen.getByRole("button", { name: /^accept$/i }));

      await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    });
  });
});
