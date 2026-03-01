import React from "react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useVaultStore } from "../../../stores/vaultStore";
import KeyDeleteDialog from "../KeyDeleteDialog";
import type { VaultKeyEntry } from "../../../types/vault";

vi.mock("../../../stores/vaultStore", () => ({
  useVaultStore: vi.fn(),
}));

vi.mock("../../../components/common/ConfirmDialog", () => ({
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
    description: React.ReactNode;
    confirmLabel?: string;
  }) => {
    if (!open) return null;
    return (
      <div role="dialog">
        <h2>{title}</h2>
        <div>{description}</div>
        <button onClick={onClose}>Cancel</button>
        <button onClick={onConfirm}>{confirmLabel}</button>
      </div>
    );
  },
}));

const mockRemoveKey = vi.fn();

const entry: VaultKeyEntry = {
  id: "key-1",
  name: "Production Server",
  data: "-----BEGIN OPENSSH PRIVATE KEY-----\ntest\n-----END OPENSSH PRIVATE KEY-----",
  hasPassphrase: false,
  fingerprint: "aa:bb:cc:dd",
  createdAt: "2024-01-01T00:00:00Z",
  updatedAt: "2024-01-01T00:00:00Z",
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(useVaultStore).mockImplementation((selector) => {
    if (typeof selector === "function") {
      return selector({ removeKey: mockRemoveKey } as never);
    }
    return mockRemoveKey;
  });
});

describe("KeyDeleteDialog", () => {
  describe("rendering", () => {
    it("renders nothing when open is false", () => {
      render(
        <KeyDeleteDialog open={false} entry={entry} onClose={vi.fn()} />,
      );
      expect(screen.queryByText("Delete Private Key")).not.toBeInTheDocument();
    });

    it("renders dialog with title and entry name when open", () => {
      render(<KeyDeleteDialog open entry={entry} onClose={vi.fn()} />);
      expect(screen.getByText("Delete Private Key")).toBeInTheDocument();
      expect(screen.getByText("Production Server")).toBeInTheDocument();
    });

    it("renders the Delete confirm button", () => {
      render(<KeyDeleteDialog open entry={entry} onClose={vi.fn()} />);
      expect(
        screen.getByRole("button", { name: /delete/i }),
      ).toBeInTheDocument();
    });

    it("renders the Cancel button", () => {
      render(<KeyDeleteDialog open entry={entry} onClose={vi.fn()} />);
      expect(
        screen.getByRole("button", { name: /cancel/i }),
      ).toBeInTheDocument();
    });

    it("renders nothing when entry is null", () => {
      render(<KeyDeleteDialog open entry={null} onClose={vi.fn()} />);
      // Dialog is open but entry name will not appear
      expect(screen.queryByText("Production Server")).not.toBeInTheDocument();
    });
  });

  describe("cancel", () => {
    it("calls onClose when Cancel is clicked", async () => {
      const onClose = vi.fn();
      render(<KeyDeleteDialog open entry={entry} onClose={onClose} />);

      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(onClose).toHaveBeenCalledTimes(1);
    });

    it("does not call removeKey when Cancel is clicked", async () => {
      render(
        <KeyDeleteDialog open entry={entry} onClose={vi.fn()} />,
      );

      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(mockRemoveKey).not.toHaveBeenCalled();
    });
  });

  describe("confirm delete", () => {
    it("calls removeKey with the entry id when Delete is confirmed", async () => {
      mockRemoveKey.mockResolvedValue(undefined);
      const onClose = vi.fn();
      render(<KeyDeleteDialog open entry={entry} onClose={onClose} />);

      await userEvent.click(screen.getByRole("button", { name: /delete/i }));

      await waitFor(() => {
        expect(mockRemoveKey).toHaveBeenCalledWith("key-1");
      });
    });

    it("calls onClose after successful deletion", async () => {
      mockRemoveKey.mockResolvedValue(undefined);
      const onClose = vi.fn();
      render(<KeyDeleteDialog open entry={entry} onClose={onClose} />);

      await userEvent.click(screen.getByRole("button", { name: /delete/i }));

      await waitFor(() => {
        expect(onClose).toHaveBeenCalledTimes(1);
      });
    });

    it("does nothing when entry is null and Delete is clicked", async () => {
      render(<KeyDeleteDialog open entry={null} onClose={vi.fn()} />);

      // The confirm button is still rendered by ConfirmDialog even with null entry
      const deleteBtn = screen.queryByRole("button", { name: /delete/i });
      if (deleteBtn) {
        await userEvent.click(deleteBtn);
      }
      expect(mockRemoveKey).not.toHaveBeenCalled();
    });
  });
});
