import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useVaultStore } from "@/stores/vaultStore";
import KeyDeleteDialog from "../KeyDeleteDialog";
import type { VaultKeyEntry } from "@/types/vault";

vi.mock("@/stores/vaultStore", () => ({
  useVaultStore: vi.fn(),
}));

vi.mock("@/components/common/ConfirmDialog", async () => ({
  default: (await import("@/tests/mocks")).MockConfirmDialog,
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
      render(<KeyDeleteDialog open={false} entry={entry} onClose={vi.fn()} />);
      expect(screen.queryByText("Delete Private Key")).not.toBeInTheDocument();
    });

    it("renders dialog with title and entry name when open", () => {
      render(<KeyDeleteDialog open entry={entry} onClose={vi.fn()} />);
      expect(screen.getByText("Delete Private Key")).toBeInTheDocument();
      expect(screen.getByText("Production Server")).toBeInTheDocument();
    });

    it("renders nothing when entry is null", () => {
      render(<KeyDeleteDialog open entry={null} onClose={vi.fn()} />);
      expect(screen.queryByText("Production Server")).not.toBeInTheDocument();
    });
  });

  describe("cancel", () => {
    it("closes without deleting when Cancel is clicked", async () => {
      const onClose = vi.fn();
      render(<KeyDeleteDialog open entry={entry} onClose={onClose} />);

      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(onClose).toHaveBeenCalledTimes(1);
      expect(mockRemoveKey).not.toHaveBeenCalled();
    });
  });

  describe("confirm delete", () => {
    it("removes the entry by id and closes when Delete is confirmed", async () => {
      mockRemoveKey.mockResolvedValue(undefined);
      const onClose = vi.fn();
      render(<KeyDeleteDialog open entry={entry} onClose={onClose} />);

      await userEvent.click(screen.getByRole("button", { name: /delete/i }));

      await waitFor(() => {
        expect(mockRemoveKey).toHaveBeenCalledWith("key-1");
      });
      await waitFor(() => {
        expect(onClose).toHaveBeenCalledTimes(1);
      });
    });

    it("shows error message when removeKey fails", async () => {
      mockRemoveKey.mockRejectedValue(new Error("Storage full"));
      render(<KeyDeleteDialog open entry={entry} onClose={vi.fn()} />);

      await userEvent.click(screen.getByRole("button", { name: /delete/i }));

      await waitFor(() => {
        expect(screen.getByText("Storage full")).toBeInTheDocument();
      });
    });
  });
});
