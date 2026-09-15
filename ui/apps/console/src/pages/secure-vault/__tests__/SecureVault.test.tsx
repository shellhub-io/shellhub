import React from "react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { useVaultStore } from "@/stores/vaultStore";
import SecureVault from "../index";
import ConnectDrawer from "@/components/ConnectDrawer";
import type { VaultKeyEntry } from "@/types/vault";

vi.mock("@/components/common/PageHeader", () => ({
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

vi.mock("@/stores/vaultStore", () => ({
  useVaultStore: vi.fn((selector?: (s: Record<string, unknown>) => unknown) => {
    const state = (
      useVaultStore as unknown as { _state: Record<string, unknown> }
    )._state;
    return selector ? selector(state) : state;
  }),
}));

vi.mock("@/stores/terminalStore", () => ({
  useTerminalStore: vi.fn(
    (selector?: (s: Record<string, unknown>) => unknown) => {
      const state = { open: vi.fn() };
      return selector ? selector(state) : state;
    },
  ),
}));

vi.mock("@/utils/sshKeys", () => ({
  validatePrivateKey: vi.fn(() => ({ valid: false, encrypted: false })),
  getFingerprint: vi.fn(() => "fp"),
}));

vi.mock("@/components/common/Drawer", async () => ({
  default: (await import("@/tests/mocks")).MockDrawer,
}));

vi.mock("@/components/vault/VaultLockedBanner", () => ({
  default: ({ onUnlock }: { onUnlock: () => void }) => (
    <div data-testid="vault-locked-banner">
      <button type="button" onClick={onUnlock}>
        Unlock Vault
      </button>
    </div>
  ),
}));

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
}));

vi.mock("@/components/common/fields/InputField", () => ({
  default: ({
    id,
    label,
    value,
    onChange,
    placeholder,
  }: {
    id: string;
    label: string;
    value: string;
    onChange: (v: string) => void;
    placeholder?: string;
    autoFocus?: boolean;
  }) => (
    <div>
      <label htmlFor={id}>{label}</label>
      <input
        id={id}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
      />
    </div>
  ),
}));

vi.mock("@/components/common/fields/PasswordField", () => ({
  default: ({
    id,
    label,
    value,
    onChange,
    placeholder,
  }: {
    id: string;
    label: string;
    value: string;
    onChange: (v: string) => void;
    placeholder?: string;
    autoComplete?: string;
    suppressPasswordManager?: boolean;
    hint?: string;
  }) => (
    <div>
      <label htmlFor={id}>{label}</label>
      <input
        id={id}
        type="password"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
      />
    </div>
  ),
}));

vi.mock("@/components/common/fields/FieldLabel", () => ({
  default: ({
    children,
    htmlFor,
  }: {
    children: React.ReactNode;
    htmlFor?: string;
  }) => <label htmlFor={htmlFor}>{children}</label>,
}));

vi.mock("@/components/common/fields/RadioCard", () => ({
  default: ({
    value,
    label,
  }: {
    value: string;
    label: string;
    icon?: React.ReactNode;
    description?: string;
  }) => <div data-testid={`radio-card-${value}`}>{label}</div>,
}));

vi.mock("@/components/common/fields/RadioGroupField", () => ({
  default: ({
    label,
    value,
    onChange,
    children,
    containerClassName,
  }: {
    label: string;
    value: string;
    onChange: (v: string) => void;
    children: React.ReactNode;
    containerClassName?: string;
  }) => (
    <div>
      <span>{label}</span>
      <div className={containerClassName}>{children}</div>
      <select
        aria-label={label}
        value={value}
        onChange={(e) => onChange(e.target.value)}
      >
        <option value="password">Password</option>
        <option value="key">Private Key</option>
        <option value="vault">Vault</option>
        <option value="manual">Manual</option>
      </select>
    </div>
  ),
}));

vi.mock("@/components/common/fields/RadioSegment", () => ({
  default: ({
    value,
    label,
  }: {
    value: string;
    label: string;
    icon?: React.ReactNode;
  }) => <div data-testid={`radio-segment-${value}`}>{label}</div>,
}));

vi.mock("@/components/vault/VaultSetupDialog", () => ({
  default: ({ open, onClose }: { open: boolean; onClose: () => void }) =>
    open ? (
      <div role="dialog" aria-label="Setup Vault">
        <button type="button" onClick={onClose}>
          Close Setup
        </button>
      </div>
    ) : null,
}));

vi.mock("@/components/vault/VaultUnlockDialog", () => ({
  default: ({ open, onClose }: { open: boolean; onClose: () => void }) =>
    open ? (
      <div role="dialog" aria-label="Unlock Vault">
        <button type="button" onClick={onClose}>
          Close Unlock
        </button>
      </div>
    ) : null,
}));

vi.mock("@/components/vault/VaultSettingsSection", () => ({
  default: () => <div data-testid="vault-settings-section" />,
}));

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
    open ? (
      <div
        role="dialog"
        aria-label={editKey ? "Edit Private Key" : "Add Private Key"}
      >
        <button type="button" onClick={onClose}>
          Close Drawer
        </button>
      </div>
    ) : null,
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
    open ? (
      <div role="dialog" aria-label="Delete Key Dialog">
        {entry && <span>{entry.name}</span>}
        <button type="button" onClick={onClose}>
          Close Delete
        </button>
      </div>
    ) : null,
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
  autoLockNonce: number = 0,
) {
  (useVaultStore as unknown as { _state: Record<string, unknown> })._state = {
    status,
    keys,
    refreshStatus: mockRefreshStatus,
    autoLockNonce,
  };
}

function getState() {
  return (useVaultStore as unknown as { _state: Record<string, unknown> })
    ._state;
}
beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
  server.use(
    http.get("*/api/namespaces/:tenant", () => HttpResponse.json(null)),
    http.get("*/api/auth/token/:tenant", () =>
      HttpResponse.json({ token: "jwt-token", role: "owner" }),
    ),
    http.post(
      "*/api/ssh-identities",
      () => new HttpResponse(null, { status: 204 }),
    ),
  );
});

describe("SecureVault", () => {
  describe("uninitialized state", () => {
    it("calls refreshStatus on mount", () => {
      setupStore("uninitialized");
      render(<SecureVault />);
      expect(mockRefreshStatus).toHaveBeenCalledTimes(1);
    });
  });

  describe("locked state", () => {
    it("does not render the keys table", () => {
      setupStore("locked");
      render(<SecureVault />);
      expect(screen.queryByRole("table")).not.toBeInTheDocument();
    });
  });

  describe("unlocked state — with keys", () => {
    const keys = [
      makeKey({
        id: "key-1",
        name: "Production Server",
        fingerprint: "aa:bb:cc:dd",
      }),
      makeKey({
        id: "key-2",
        name: "Staging Server",
        fingerprint: "11:22:33:44",
        hasPassphrase: true,
      }),
    ];

    it("shows lock icon for keys with passphrase", () => {
      setupStore("unlocked", keys);
      render(<SecureVault />);
      expect(screen.getByTitle("Encrypted")).toBeInTheDocument();
    });

    it("does not show lock icon for keys without passphrase", () => {
      setupStore("unlocked", [
        makeKey({
          id: "key-1",
          name: "Production Server",
          fingerprint: "aa:bb:cc:dd",
          hasPassphrase: false,
        }),
      ]);
      render(<SecureVault />);
      expect(screen.queryByTitle("Encrypted")).not.toBeInTheDocument();
    });

    it("opens the Delete dialog with the correct entry when Delete is clicked", async () => {
      setupStore("unlocked", keys);
      render(<SecureVault />);

      const deleteButtons = screen.getAllByTitle(/delete/i);
      await userEvent.click(deleteButtons[0]);

      const dialog = screen.getByRole("dialog", { name: /delete key dialog/i });
      expect(dialog).toBeInTheDocument();
      expect(dialog).toHaveTextContent("Production Server");
    });
  });

  describe("search filtering", () => {
    const keys = [
      makeKey({
        id: "key-1",
        name: "Production Server",
        fingerprint: "aa:bb:cc:dd",
      }),
      makeKey({
        id: "key-2",
        name: "Staging Server",
        fingerprint: "11:22:33:44",
      }),
    ];

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

      expect(screen.getByText(/no keys matching/i)).toBeInTheDocument();
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

  describe("auto-lock nonce — auto-open unlock dialog", () => {
    it("opens VaultUnlockDialog when autoLockNonce bumps while mounted (unlocked → locked)", () => {
      setupStore("unlocked", [], 0);
      const { rerender } = render(<SecureVault />);

      act(() => {
        getState().status = "locked";
        getState().autoLockNonce = 1;
      });
      rerender(<SecureVault />);

      expect(
        screen.getByRole("dialog", { name: /unlock vault/i }),
      ).toBeInTheDocument();
    });

    it("does NOT auto-open unlock dialog on manual lock (nonce unchanged)", () => {
      setupStore("unlocked", [], 0);
      const { rerender } = render(<SecureVault />);

      act(() => {
        getState().status = "locked";
      });
      rerender(<SecureVault />);

      expect(
        screen.getByRole("heading", { name: /your vault is locked/i }),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("dialog", { name: /unlock vault/i }),
      ).not.toBeInTheDocument();
    });

    it("does NOT auto-open unlock dialog for a stale nonce after remount", () => {
      setupStore("locked", [], 1);
      render(<SecureVault />);

      expect(
        screen.getByRole("heading", { name: /your vault is locked/i }),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("dialog", { name: /unlock vault/i }),
      ).not.toBeInTheDocument();
    });

    it("hook count is stable across unlocked→locked transition (no hook-order crash)", () => {
      setupStore("unlocked", [], 0);
      const { rerender } = render(<SecureVault />);

      act(() => {
        getState().status = "locked";
        getState().autoLockNonce = 1;
      });

      expect(() => rerender(<SecureVault />)).not.toThrow();
    });
  });

  describe("ConnectDrawer — auto-lock degrade", () => {
    const vaultKey = makeKey({
      id: "vault-key-1",
      name: "My Key",
      fingerprint: "ff:ee:dd:cc",
    });

    function setupConnectStore(
      status: "uninitialized" | "locked" | "unlocked",
      keys: VaultKeyEntry[] = [],
    ) {
      (useVaultStore as unknown as { _state: Record<string, unknown> })._state =
        {
          status,
          keys,
          refreshStatus: mockRefreshStatus,
          autoLockNonce: 0,
        };
    }

    function renderConnectDrawer() {
      return render(
        <ConnectDrawer
          open
          onClose={vi.fn()}
          deviceUid="dev-1"
          deviceName="my-device"
          sshid="my-device.ns@localhost"
        />,
        { wrapper: createTestWrapper() },
      );
    }

    it("shows locked UI and disables Connect when vault auto-locks while drawer is open", () => {
      setupConnectStore("unlocked", [vaultKey]);
      const { rerender } = renderConnectDrawer();

      act(() => {
        setupConnectStore("locked", []);
      });
      rerender(
        <ConnectDrawer
          open
          onClose={vi.fn()}
          deviceUid="dev-1"
          deviceName="my-device"
          sshid="my-device.ns@localhost"
        />,
      );

      const connectBtn = screen.getByRole("button", { name: /connect/i });
      expect(connectBtn).toBeDisabled();
    });

    it("hides key-selection UI when vault is locked", () => {
      setupConnectStore("unlocked", [vaultKey]);
      const { rerender } = renderConnectDrawer();

      act(() => {
        setupConnectStore("locked", []);
      });
      rerender(
        <ConnectDrawer
          open
          onClose={vi.fn()}
          deviceUid="dev-1"
          deviceName="my-device"
          sshid="my-device.ns@localhost"
        />,
      );

      expect(
        screen.queryByTestId("radio-segment-vault"),
      ).not.toBeInTheDocument();
      expect(
        screen.queryByTestId("radio-segment-manual"),
      ).not.toBeInTheDocument();
    });

    it("does not crash when vault auto-locks while key auth mode is active", () => {
      setupConnectStore("unlocked", [vaultKey]);
      const { rerender } = renderConnectDrawer();

      act(() => {
        setupConnectStore("locked", []);
      });

      expect(() =>
        rerender(
          <ConnectDrawer
            open
            onClose={vi.fn()}
            deviceUid="dev-1"
            deviceName="my-device"
            sshid="my-device.ns@localhost"
          />,
        ),
      ).not.toThrow();
    });
  });
});
