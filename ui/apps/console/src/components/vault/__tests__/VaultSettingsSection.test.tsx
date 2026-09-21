import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useVaultStore } from "@/stores/vaultStore";
import { isVaultServerEnabled } from "@/utils/vault-backend-factory";
import VaultSettingsSection from "../VaultSettingsSection";

vi.mock("@/utils/vault-crypto", () => ({
  createVaultMeta: vi.fn(),
  verifyPassword: vi.fn(),
  encrypt: vi.fn(),
  decrypt: vi.fn(),
  setSessionKey: vi.fn(),
  getSessionKey: vi.fn(),
  clearSessionKey: vi.fn(),
}));

vi.mock("@/utils/vault-backend-factory", () => ({
  getVaultBackend: vi.fn(() => ({
    loadMeta: vi.fn(() => null),
    loadData: vi.fn(() => null),
    loadSettings: vi.fn(() => ({
      autoLockTimeoutMinutes: 15,
      lockOnHidden: false,
    })),
    saveMeta: vi.fn(),
    saveData: vi.fn(),
    saveSettings: vi.fn(),
    loadLegacyKeys: vi.fn(() => []),
    clearLegacyKeys: vi.fn(),
    clear: vi.fn(),
  })),
  isVaultServerEnabled: vi.fn(() => false),
}));

vi.mock("@/stores/authStore", () => {
  const state = { user: "testuser", tenant: "test-tenant" };
  const useAuthStore = Object.assign(
    vi.fn((selector: (s: typeof state) => unknown) => selector(state)),
    { getState: vi.fn(() => state) },
  );
  return { useAuthStore };
});

vi.mock("@/utils/vault-activity-tracker", () => ({
  start: vi.fn(),
  stop: vi.fn(),
}));

vi.mock("@/utils/vault-migrate", () => ({
  serverVaultExists: vi.fn(() => Promise.resolve(false)),
  migrateLocalToServer: vi.fn(),
  migrateServerToLocal: vi.fn(),
  adoptServerVault: vi.fn(),
  localVaultExists: vi.fn(() => false),
}));

function renderSection() {
  return render(<VaultSettingsSection />);
}

function setUnlocked(
  overrides: { autoLockTimeoutMinutes?: number; lockOnHidden?: boolean } = {},
) {
  useVaultStore.setState({
    status: "unlocked",
    autoLockTimeoutMinutes: overrides.autoLockTimeoutMinutes ?? 15,
    lockOnHidden: overrides.lockOnHidden ?? false,
  });
}

beforeEach(() => {
  useVaultStore.setState({
    status: "locked",
    keys: [],
    loading: false,
    error: null,
    autoLockTimeoutMinutes: 15,
    lockOnHidden: false,
    storageMode: "local",
    autoLockNonce: 0,
  });
  vi.clearAllMocks();
  vi.mocked(isVaultServerEnabled).mockReturnValue(false);
});

describe("VaultSettingsSection", () => {
  it.each(["locked", "uninitialized"] as const)(
    "renders nothing when %s",
    (status) => {
      useVaultStore.setState({ status });
      const { container } = renderSection();
      expect(container).toBeEmptyDOMElement();
    },
  );

  describe("Auto-lock timeout menu", () => {
    it("renders all 5 timeout options when opened", async () => {
      setUnlocked();
      renderSection();

      await userEvent.click(
        screen.getByRole("button", { name: /auto-lock timeout/i }),
      );

      const items = screen.getAllByRole("menuitem");
      expect(items).toHaveLength(5);
      expect(items.map((i) => i.textContent)).toEqual([
        "Never",
        "5 minutes",
        "15 minutes",
        "30 minutes",
        "60 minutes",
      ]);
    });

    it.each([
      ["30 minutes", 30],
      [/never/i, 0],
    ] as const)("selecting %s persists a timeout of %i", async (
      option,
      expected,
    ) => {
      setUnlocked({ autoLockTimeoutMinutes: 15 });
      const updateAutoLockSettings = vi.fn();
      useVaultStore.setState({ updateAutoLockSettings });

      renderSection();

      await userEvent.click(
        screen.getByRole("button", { name: /auto-lock timeout/i }),
      );
      await userEvent.click(screen.getByRole("menuitem", { name: option }));

      expect(updateAutoLockSettings).toHaveBeenCalledWith({
        autoLockTimeoutMinutes: expected,
      });
    });

    it.each([
      [15, "15 minutes"],
      [0, "Never"],
    ])("a persisted timeout of %i paints as '%s'", (minutes, label) => {
      setUnlocked({ autoLockTimeoutMinutes: minutes });
      renderSection();

      expect(
        screen.getByRole("button", { name: /auto-lock timeout/i }),
      ).toHaveTextContent(label);
    });
  });

  describe("Lock-when-tab-hidden checkbox", () => {
    function hiddenCheckbox() {
      return screen.getByRole("checkbox", { name: /lock when hidden/i });
    }

    it.each([false, true])(
      "toggling from lockOnHidden=%s persists the opposite",
      async (lockOnHidden) => {
        setUnlocked({ lockOnHidden });
        const updateAutoLockSettings = vi.fn();
        useVaultStore.setState({ updateAutoLockSettings });

        renderSection();

        await userEvent.click(hiddenCheckbox());

        expect(updateAutoLockSettings).toHaveBeenCalledWith({
          lockOnHidden: !lockOnHidden,
        });
      },
    );

    it.each([true, false])(
      "a persisted lockOnHidden=%s paints the checkbox accordingly",
      (lockOnHidden) => {
        setUnlocked({ lockOnHidden });
        renderSection();

        if (lockOnHidden) expect(hiddenCheckbox()).toBeChecked();
        else expect(hiddenCheckbox()).not.toBeChecked();
      },
    );
  });

  describe("Change master password", () => {
    it("opens ChangePasswordDrawer when the Change button is clicked", async () => {
      setUnlocked();
      renderSection();

      const heading = screen.getByRole("heading", {
        name: /change master password/i,
        hidden: true,
      });
      expect(heading.closest("[inert]")).not.toBeNull();

      const changeBtn = screen.getByRole("button", {
        name: "Change master password",
      });
      await userEvent.click(changeBtn);

      expect(
        screen.getByRole("heading", { name: /change master password/i }),
      ).toBeInTheDocument();
      expect(heading.closest("[inert]")).toBeNull();
    });
  });

  describe("Lock vault", () => {
    it("calls the lock() store action when the Lock button is clicked", async () => {
      setUnlocked();
      const lock = vi.fn();
      useVaultStore.setState({ lock });

      renderSection();

      const lockBtn = screen.getByRole("button", { name: "Lock vault" });
      await userEvent.click(lockBtn);

      expect(lock).toHaveBeenCalledTimes(1);
    });
  });

  describe("Reset vault", () => {
    it("opens ConfirmDialog with title 'Reset Secure Vault' when Reset button is clicked", async () => {
      setUnlocked();
      renderSection();

      const resetBtn = screen.getByRole("button", { name: "Reset vault" });
      await userEvent.click(resetBtn);

      expect(
        screen.getByRole("heading", { name: /reset secure vault/i }),
      ).toBeInTheDocument();
    });
  });

  describe("Storage row", () => {
    it.each([true, false])(
      "isVaultServerEnabled()=%s decides whether the row renders",
      (enabled) => {
        vi.mocked(isVaultServerEnabled).mockReturnValue(enabled);
        useVaultStore.setState({ storageMode: "local" });
        setUnlocked();
        renderSection();

        const row = screen.queryByText(/^storage$/i);
        if (enabled) expect(row).toBeInTheDocument();
        else expect(row).not.toBeInTheDocument();
      },
    );
  });

  describe("Storage row copy", () => {
    beforeEach(() => {
      vi.mocked(isVaultServerEnabled).mockReturnValue(true);
    });

    it.each([
      [
        "server",
        "Move",
        "Synced with the ShellHub server. Click to move it to this device.",
      ],
      [
        "local",
        "Sync",
        "Stored in this browser only. Click to sync it to the ShellHub server.",
      ],
    ] as const)("a %s vault offers '%s'", (storageMode, action, copy) => {
      useVaultStore.setState({ storageMode });
      setUnlocked();
      renderSection();

      expect(
        screen.getByRole("button", { name: "Change vault storage location" }),
      ).toHaveTextContent(action);
      expect(screen.getByText(copy)).toBeInTheDocument();
    });

    it("opens VaultSyncDialog when storage button is clicked", async () => {
      useVaultStore.setState({ storageMode: "local" });
      setUnlocked();
      renderSection();

      await userEvent.click(
        screen.getByRole("button", { name: "Change vault storage location" }),
      );

      expect(
        screen.getByRole("heading", { name: /sync/i }),
      ).toBeInTheDocument();
    });
  });
});
