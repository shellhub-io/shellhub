import React from "react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useVaultStore } from "../../../stores/vaultStore";
import SecureVault from "../index";
import type { VaultKeyEntry } from "../../../types/vault";

vi.mock("../../../utils/date", () => ({
  formatDate: (d: string) => d,
}));

vi.mock("../../../components/common/PageHeader", () => ({
  default: ({
    title,
    children,
  }: {
    title: string;
    children?: React.ReactNode;
  }) => (
    <div>
      <h1>{title}</h1>
      {children}
    </div>
  ),
}));

vi.mock("../../../stores/vaultStore", () => ({
  useVaultStore: vi.fn((selector?: (s: Record<string, unknown>) => unknown) => {
    const state = (useVaultStore as unknown as { _state: Record<string, unknown> })._state;
    return selector ? selector(state) : state;
  }),
}));

// Mock heavy vault dialog/banner components to isolate page logic
vi.mock("../../../components/vault/VaultSetupDialog", () => ({
  default: ({
    open,
    onClose,
  }: {
    open: boolean;
    onClose: () => void;
  }) =>
    open
      ? (
        <div role="dialog" aria-label="Setup Vault">
          <button onClick={onClose}>Close Setup</button>
        </div>
      )
      : null,
}));

vi.mock("../../../components/vault/VaultUnlockDialog", () => ({
  default: ({
    open,
    onClose,
  }: {
    open: boolean;
    onClose: () => void;
  }) =>
    open
      ? (
        <div role="dialog" aria-label="Unlock Vault">
          <button onClick={onClose}>Close Unlock</button>
        </div>
      )
      : null,
}));

vi.mock("../../../components/vault/VaultSettingsSection", () => ({
  default: () => <div data-testid="vault-settings-section" />,
}));

// KeyDrawer and KeyDeleteDialog are also rendered by the page.  Mock them so
// we can assert they receive the right props without executing their internals.
vi.mock("../KeyDrawer", () => ({
  default: ({
    open,
    editKey,
    onClose,
  }: {
    open: boolean;
    editKey: VaultKeyEntry | null;
    onClose: () => void;
  }) =>
    open
      ? (
        <div
          role="dialog"
          aria-label={editKey ? "Edit Private Key" : "Add Private Key"}
        >
          <button onClick={onClose}>Close Drawer</button>
        </div>
      )
      : null,
}));

vi.mock("../KeyDeleteDialog", () => ({
  default: ({
    open,
    entry,
    onClose,
  }: {
    open: boolean;
    entry: VaultKeyEntry | null;
    onClose: () => void;
  }) =>
    open
      ? (
        <div role="dialog" aria-label="Delete Key Dialog">
          {entry && <span>{entry.name}</span>}
          <button onClick={onClose}>Close Delete</button>
        </div>
      )
      : null,
}));

// CopyButton is used inside the key table — stub it out
vi.mock("../../../components/common/CopyButton", () => ({
  default: ({ text }: { text: string }) => (
    <button aria-label={`Copy ${text}`}>Copy</button>
  ),
}));

const makeKey = (overrides: Partial<VaultKeyEntry> = {}): VaultKeyEntry => ({
  id: "key-1",
  name: "Production Server",
  data: "-----BEGIN OPENSSH PRIVATE KEY-----\ndata\n-----END OPENSSH PRIVATE KEY-----",
  hasPassphrase: false,
  fingerprint: "aa:bb:cc:dd",
  createdAt: "2024-01-01T00:00:00Z",
  updatedAt: "2024-01-01T00:00:00Z",
  ...overrides,
});

const mockRefreshStatus = vi.fn();

function setupStore(
  status: "uninitialized" | "locked" | "unlocked",
  keys: VaultKeyEntry[] = [],
) {
  (useVaultStore as unknown as { _state: Record<string, unknown> })._state = {
    status,
    keys,
    refreshStatus: mockRefreshStatus,
  };
}

afterEach(cleanup);

beforeEach(() => {
  vi.clearAllMocks();
});

describe("SecureVault", () => {
  describe("uninitialized state", () => {
    it("renders the Secure Vault setup landing page", () => {
      setupStore("uninitialized");
      render(<SecureVault />);
      expect(screen.getByText("Secure Vault")).toBeInTheDocument();
    });

    it("shows the 'Set Up Secure Vault' button", () => {
      setupStore("uninitialized");
      render(<SecureVault />);
      expect(
        screen.getByRole("button", { name: /set up secure vault/i }),
      ).toBeInTheDocument();
    });

    it("shows feature highlights: AES-256 Encryption, Zero Knowledge, Quick Connect", () => {
      setupStore("uninitialized");
      render(<SecureVault />);
      expect(screen.getByText("AES-256 Encryption")).toBeInTheDocument();
      expect(screen.getByText("Zero Knowledge")).toBeInTheDocument();
      expect(screen.getByText("Quick Connect")).toBeInTheDocument();
    });

    it("opens the setup dialog when 'Set Up Secure Vault' is clicked", async () => {
      setupStore("uninitialized");
      render(<SecureVault />);

      await userEvent.click(
        screen.getByRole("button", { name: /set up secure vault/i }),
      );

      expect(
        screen.getByRole("dialog", { name: /setup vault/i }),
      ).toBeInTheDocument();
    });

    it("closes the setup dialog when onClose is triggered", async () => {
      setupStore("uninitialized");
      render(<SecureVault />);

      await userEvent.click(
        screen.getByRole("button", { name: /set up secure vault/i }),
      );
      await userEvent.click(screen.getByRole("button", { name: /close setup/i }));

      expect(
        screen.queryByRole("dialog", { name: /setup vault/i }),
      ).not.toBeInTheDocument();
    });

    it("calls refreshStatus on mount", () => {
      setupStore("uninitialized");
      render(<SecureVault />);
      expect(mockRefreshStatus).toHaveBeenCalledTimes(1);
    });
  });

  describe("locked state", () => {
    it("renders the full-screen locked page with heading and highlights", () => {
      setupStore("locked");
      render(<SecureVault />);
      expect(screen.getByRole("heading", { name: /your vault is locked/i })).toBeInTheDocument();
      expect(screen.getByText("AES-256 Encryption")).toBeInTheDocument();
      expect(screen.getByText("Zero Knowledge")).toBeInTheDocument();
      expect(screen.getByText("Quick Connect")).toBeInTheDocument();
    });

    it("opens the unlock dialog when 'Unlock Vault' is clicked", async () => {
      setupStore("locked");
      render(<SecureVault />);

      await userEvent.click(screen.getByRole("button", { name: /unlock vault/i }));

      expect(
        screen.getByRole("dialog", { name: /unlock vault/i }),
      ).toBeInTheDocument();
    });

    it("closes the unlock dialog when onClose is triggered", async () => {
      setupStore("locked");
      render(<SecureVault />);

      await userEvent.click(screen.getByRole("button", { name: /unlock vault/i }));
      await userEvent.click(
        screen.getByRole("button", { name: /close unlock/i }),
      );

      expect(
        screen.queryByRole("dialog", { name: /unlock vault/i }),
      ).not.toBeInTheDocument();
    });

    it("does not render the keys table", () => {
      setupStore("locked");
      render(<SecureVault />);
      expect(screen.queryByRole("table")).not.toBeInTheDocument();
    });
  });

  describe("unlocked state — empty vault", () => {
    it("renders the empty state message", () => {
      setupStore("unlocked", []);
      render(<SecureVault />);
      expect(
        screen.getByText(/no keys yet/i),
      ).toBeInTheDocument();
    });

    it("renders 'Add Private Key' button in empty state", () => {
      setupStore("unlocked", []);
      render(<SecureVault />);
      expect(
        screen.getByRole("button", { name: /add private key/i }),
      ).toBeInTheDocument();
    });

    it("opens the Add Key drawer when 'Add Private Key' is clicked", async () => {
      setupStore("unlocked", []);
      render(<SecureVault />);

      await userEvent.click(
        screen.getByRole("button", { name: /add private key/i }),
      );

      expect(
        screen.getByRole("dialog", { name: /add private key/i }),
      ).toBeInTheDocument();
    });

    it("closes the Add Key drawer when onClose is triggered", async () => {
      setupStore("unlocked", []);
      render(<SecureVault />);

      await userEvent.click(
        screen.getByRole("button", { name: /add private key/i }),
      );
      await userEvent.click(
        screen.getByRole("button", { name: /close drawer/i }),
      );

      expect(
        screen.queryByRole("dialog", { name: /add private key/i }),
      ).not.toBeInTheDocument();
    });
  });

  describe("unlocked state — with keys", () => {
    const keys = [
      makeKey({ id: "key-1", name: "Production Server", fingerprint: "aa:bb:cc:dd" }),
      makeKey({
        id: "key-2",
        name: "Staging Server",
        fingerprint: "11:22:33:44",
        hasPassphrase: true,
      }),
    ];

    it("renders a table with a row per key", () => {
      setupStore("unlocked", keys);
      render(<SecureVault />);
      expect(screen.getByRole("table")).toBeInTheDocument();
      expect(screen.getByText("Production Server")).toBeInTheDocument();
      expect(screen.getByText("Staging Server")).toBeInTheDocument();
    });

    it("shows lock icon for keys with passphrase", () => {
      setupStore("unlocked", keys);
      render(<SecureVault />);
      expect(screen.getByTitle("Encrypted")).toBeInTheDocument();
    });

    it("does not show lock icon for keys without passphrase", () => {
      setupStore("unlocked", [makeKey({ id: "key-1", name: "Production Server", fingerprint: "aa:bb:cc:dd", hasPassphrase: false })]);
      render(<SecureVault />);
      expect(screen.queryByTitle("Encrypted")).not.toBeInTheDocument();
    });

    it("renders 'Add Private Key' button in header", () => {
      setupStore("unlocked", keys);
      render(<SecureVault />);
      expect(
        screen.getByRole("button", { name: /add private key/i }),
      ).toBeInTheDocument();
    });

    it("opens the Add Key drawer from the header button", async () => {
      setupStore("unlocked", keys);
      render(<SecureVault />);

      await userEvent.click(
        screen.getByRole("button", { name: /add private key/i }),
      );

      expect(
        screen.getByRole("dialog", { name: /add private key/i }),
      ).toBeInTheDocument();
    });

    it("opens the Edit Key drawer when the Edit button is clicked", async () => {
      setupStore("unlocked", keys);
      render(<SecureVault />);

      const editButtons = screen.getAllByTitle(/edit/i);
      await userEvent.click(editButtons[0]);

      expect(
        screen.getByRole("dialog", { name: /edit private key/i }),
      ).toBeInTheDocument();
    });

    it("closes the Edit Key drawer after onClose", async () => {
      setupStore("unlocked", keys);
      render(<SecureVault />);

      const editButtons = screen.getAllByTitle(/edit/i);
      await userEvent.click(editButtons[0]);
      await userEvent.click(
        screen.getByRole("button", { name: /close drawer/i }),
      );

      expect(
        screen.queryByRole("dialog", { name: /edit private key/i }),
      ).not.toBeInTheDocument();
    });

    it("opens the Delete dialog with the correct entry when Delete is clicked", async () => {
      setupStore("unlocked", keys);
      render(<SecureVault />);

      const deleteButtons = screen.getAllByTitle(/delete/i);
      await userEvent.click(deleteButtons[0]);

      const dialog = screen.getByRole("dialog", { name: /delete key dialog/i });
      expect(dialog).toBeInTheDocument();
      // The dialog mock renders the entry name inside the dialog
      expect(dialog).toHaveTextContent("Production Server");
    });

    it("closes the Delete dialog when onClose is triggered", async () => {
      setupStore("unlocked", keys);
      render(<SecureVault />);

      const deleteButtons = screen.getAllByTitle(/delete/i);
      await userEvent.click(deleteButtons[0]);
      await userEvent.click(
        screen.getByRole("button", { name: /close delete/i }),
      );

      expect(
        screen.queryByRole("dialog", { name: /delete key dialog/i }),
      ).not.toBeInTheDocument();
    });
  });

  describe("search filtering", () => {
    const keys = [
      makeKey({ id: "key-1", name: "Production Server", fingerprint: "aa:bb:cc:dd" }),
      makeKey({ id: "key-2", name: "Staging Server", fingerprint: "11:22:33:44" }),
    ];

    it("shows all rows when search is empty", () => {
      setupStore("unlocked", keys);
      render(<SecureVault />);
      expect(screen.getByText("Production Server")).toBeInTheDocument();
      expect(screen.getByText("Staging Server")).toBeInTheDocument();
    });

    it("filters rows by name", async () => {
      setupStore("unlocked", keys);
      render(<SecureVault />);

      await userEvent.type(
        screen.getByPlaceholderText(/search by name or fingerprint/i),
        "Production",
      );

      expect(screen.getByText("Production Server")).toBeInTheDocument();
      expect(screen.queryByText("Staging Server")).not.toBeInTheDocument();
    });

    it("filters rows by fingerprint", async () => {
      setupStore("unlocked", keys);
      render(<SecureVault />);

      await userEvent.type(
        screen.getByPlaceholderText(/search by name or fingerprint/i),
        "11:22",
      );

      expect(screen.queryByText("Production Server")).not.toBeInTheDocument();
      expect(screen.getByText("Staging Server")).toBeInTheDocument();
    });

    it("shows no-match message when search yields no results", async () => {
      setupStore("unlocked", keys);
      render(<SecureVault />);

      await userEvent.type(
        screen.getByPlaceholderText(/search by name or fingerprint/i),
        "nonexistent",
      );

      expect(
        screen.getByText(/no keys matching/i),
      ).toBeInTheDocument();
    });

    it("search is case-insensitive", async () => {
      setupStore("unlocked", keys);
      render(<SecureVault />);

      await userEvent.type(
        screen.getByPlaceholderText(/search by name or fingerprint/i),
        "production",
      );

      expect(screen.getByText("Production Server")).toBeInTheDocument();
    });
  });

  describe("VaultSettingsSection", () => {
    it("renders the settings section when unlocked", () => {
      setupStore("unlocked", []);
      render(<SecureVault />);
      expect(screen.getByTestId("vault-settings-section")).toBeInTheDocument();
    });

    it("does not render settings section in uninitialized state", () => {
      setupStore("uninitialized");
      render(<SecureVault />);
      expect(
        screen.queryByTestId("vault-settings-section"),
      ).not.toBeInTheDocument();
    });

    it("does not render settings section in locked state", () => {
      setupStore("locked");
      render(<SecureVault />);
      expect(
        screen.queryByTestId("vault-settings-section"),
      ).not.toBeInTheDocument();
    });
  });
});
