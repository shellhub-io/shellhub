import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useVaultStore } from "@/stores/vaultStore";
import { isVaultServerEnabled } from "@/utils/vault-backend-factory";
import VaultSettingsSection from "../VaultSettingsSection";

// Prevent real vault-crypto / backend calls from running in tests
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
  // Reset to a known base state before each test
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
  // clearAllMocks only clears call history — it does NOT restore mockReturnValue.
  // Reset the implementation explicitly so every test starts from the documented default.
  vi.mocked(isVaultServerEnabled).mockReturnValue(false);
});

describe("VaultSettingsSection", () => {
  describe("when vault is not unlocked", () => {
    it("renders nothing when locked", () => {
      useVaultStore.setState({ status: "locked" });
      const { container } = renderSection();
      expect(container).toBeEmptyDOMElement();
    });

    it("renders nothing when uninitialized", () => {
      useVaultStore.setState({ status: "uninitialized" });
      const { container } = renderSection();
      expect(container).toBeEmptyDOMElement();
    });
  });

  describe("Auto-lock timeout dropdown", () => {
    it("renders all 5 timeout options when dropdown is opened", async () => {
      setUnlocked();
      renderSection();

      // Open the dropdown
      const trigger = screen.getByRole("button", {
        name: /auto-lock timeout/i,
      });
      await userEvent.click(trigger);

      // All 5 options should be visible — use exact strings to avoid
      // substring collisions (e.g. "5 minutes" inside "15 minutes")
      expect(
        screen.getByRole("option", { name: "5 minutes" }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("option", { name: "15 minutes" }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("option", { name: "30 minutes" }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("option", { name: "60 minutes" }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("option", { name: /never/i }),
      ).toBeInTheDocument();
    });

    it("calls updateAutoLockSettings with correct timeout when an option is selected", async () => {
      setUnlocked({ autoLockTimeoutMinutes: 15 });
      const updateAutoLockSettings = vi.fn();
      useVaultStore.setState({ updateAutoLockSettings });

      renderSection();

      const trigger = screen.getByRole("button", {
        name: /auto-lock timeout/i,
      });
      await userEvent.click(trigger);

      const option30 = screen.getByRole("option", { name: /30 minutes/i });
      await userEvent.click(option30);

      expect(updateAutoLockSettings).toHaveBeenCalledWith({
        autoLockTimeoutMinutes: 30,
      });
    });

    it("calls updateAutoLockSettings with 0 when Never is selected", async () => {
      setUnlocked({ autoLockTimeoutMinutes: 15 });
      const updateAutoLockSettings = vi.fn();
      useVaultStore.setState({ updateAutoLockSettings });

      renderSection();

      const trigger = screen.getByRole("button", {
        name: /auto-lock timeout/i,
      });
      await userEvent.click(trigger);

      const neverOption = screen.getByRole("option", { name: /never/i });
      await userEvent.click(neverOption);

      expect(updateAutoLockSettings).toHaveBeenCalledWith({
        autoLockTimeoutMinutes: 0,
      });
    });

    it("reflects the persisted timeout value on first paint (15 min)", () => {
      setUnlocked({ autoLockTimeoutMinutes: 15 });
      renderSection();

      // The trigger button should display the current setting
      expect(
        screen.getByRole("button", { name: /auto-lock timeout/i }),
      ).toHaveTextContent("15 minutes");
    });

    it("reflects the persisted timeout value on first paint (Never)", () => {
      setUnlocked({ autoLockTimeoutMinutes: 0 });
      renderSection();

      expect(
        screen.getByRole("button", { name: /auto-lock timeout/i }),
      ).toHaveTextContent("Never");
    });

    it("closes the dropdown after selecting an option", async () => {
      setUnlocked();
      const updateAutoLockSettings = vi.fn();
      useVaultStore.setState({ updateAutoLockSettings });

      renderSection();

      await userEvent.click(
        screen.getByRole("button", { name: /auto-lock timeout/i }),
      );
      // Use exact string to avoid substring collision with "15 minutes"
      await userEvent.click(screen.getByRole("option", { name: "5 minutes" }));

      // Options should no longer be visible
      expect(
        screen.queryByRole("option", { name: "5 minutes" }),
      ).not.toBeInTheDocument();
    });
  });

  describe("Lock-when-tab-hidden checkbox", () => {
    it("renders the lock-when-hidden checkbox with a description", () => {
      setUnlocked();
      renderSection();

      const checkbox = screen.getByRole("checkbox", {
        name: /lock when hidden/i,
      });
      expect(checkbox).toBeInTheDocument();
    });

    it("renders the description text about switching away", () => {
      setUnlocked();
      renderSection();

      expect(
        screen.getByText(
          /locks the vault about a minute after you switch away or minimize/i,
        ),
      ).toBeInTheDocument();
    });

    it("calls updateAutoLockSettings with lockOnHidden:true when checked", async () => {
      setUnlocked({ lockOnHidden: false });
      const updateAutoLockSettings = vi.fn();
      useVaultStore.setState({ updateAutoLockSettings });

      renderSection();

      const checkbox = screen.getByRole("checkbox", {
        name: /lock when hidden/i,
      });
      await userEvent.click(checkbox);

      expect(updateAutoLockSettings).toHaveBeenCalledWith({
        lockOnHidden: true,
      });
    });

    it("calls updateAutoLockSettings with lockOnHidden:false when unchecked", async () => {
      setUnlocked({ lockOnHidden: true });
      const updateAutoLockSettings = vi.fn();
      useVaultStore.setState({ updateAutoLockSettings });

      renderSection();

      const checkbox = screen.getByRole("checkbox", {
        name: /lock when hidden/i,
      });
      await userEvent.click(checkbox);

      expect(updateAutoLockSettings).toHaveBeenCalledWith({
        lockOnHidden: false,
      });
    });

    it("reflects persisted lockOnHidden=true on first paint (checked)", () => {
      setUnlocked({ lockOnHidden: true });
      renderSection();

      const checkbox = screen.getByRole("checkbox", {
        name: /lock when hidden/i,
      });
      expect(checkbox).toBeChecked();
    });

    it("reflects persisted lockOnHidden=false on first paint (unchecked)", () => {
      setUnlocked({ lockOnHidden: false });
      renderSection();

      const checkbox = screen.getByRole("checkbox", {
        name: /lock when hidden/i,
      });
      expect(checkbox).not.toBeChecked();
    });
  });

  describe("Change master password", () => {
    it("opens ChangePasswordDrawer when the Change button is clicked", async () => {
      setUnlocked();
      renderSection();

      // The Drawer always renders its panel in the DOM (CSS-only hide via
      // translate-x-full). When closed the panel carries inert so its
      // descendants are removed from the tab order and the accessibility tree.
      const heading = screen.getByRole("heading", {
        name: /change master password/i,
        hidden: true,
      });
      expect(heading.closest("[inert]")).not.toBeNull();

      // Requires aria-label="Change master password" (lowercase 'm') on the button
      const changeBtn = screen.getByRole("button", {
        name: "Change master password",
      });
      await userEvent.click(changeBtn);

      // After opening, inert is removed — the heading becomes accessible and
      // the password inputs are queryable without hidden:true.
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

      // Requires aria-label="Lock vault" (lowercase 'v') on the button
      const lockBtn = screen.getByRole("button", { name: "Lock vault" });
      await userEvent.click(lockBtn);

      expect(lock).toHaveBeenCalledTimes(1);
    });
  });

  describe("Reset vault", () => {
    it("opens ConfirmDialog with title 'Reset Secure Vault' when Reset button is clicked", async () => {
      setUnlocked();
      renderSection();

      // Requires aria-label="Reset vault" (lowercase 'v') on the button
      const resetBtn = screen.getByRole("button", { name: "Reset vault" });
      await userEvent.click(resetBtn);

      expect(
        screen.getByRole("heading", { name: /reset secure vault/i }),
      ).toBeInTheDocument();
    });
  });

  describe("Storage row", () => {
    it("is visible when isVaultServerEnabled() returns true", () => {
      vi.mocked(isVaultServerEnabled).mockReturnValue(true);
      useVaultStore.setState({ storageMode: "local" });
      setUnlocked();
      renderSection();

      expect(screen.getByText(/^storage$/i)).toBeInTheDocument();
    });

    it("is absent when isVaultServerEnabled() returns false", () => {
      vi.mocked(isVaultServerEnabled).mockReturnValue(false);
      useVaultStore.setState({ storageMode: "local" });
      setUnlocked();
      renderSection();

      expect(screen.queryByText(/^storage$/i)).not.toBeInTheDocument();
    });
  });

  describe("SettingsCard layout (unit 5)", () => {
    describe("Vault Settings card", () => {
      it("renders a heading 'Vault Settings'", () => {
        setUnlocked();
        renderSection();
        expect(
          screen.getByRole("heading", { name: "Vault Settings" }),
        ).toBeInTheDocument();
      });

      it("renders row title 'Change Master Password'", () => {
        setUnlocked();
        renderSection();
        // Drawer also renders this title, so use getAllByText and check at least one match
        expect(
          screen.getAllByText("Change Master Password").length,
        ).toBeGreaterThanOrEqual(1);
      });

      it("renders row description for Change Master Password", () => {
        setUnlocked();
        renderSection();
        expect(
          screen.getByText("Re-encrypt all keys with a new password."),
        ).toBeInTheDocument();
      });

      it("renders a 'Change' button with aria-label 'Change master password'", () => {
        setUnlocked();
        renderSection();
        const btn = screen.getByRole("button", {
          name: "Change master password",
        });
        expect(btn).toHaveTextContent("Change");
      });

      it("renders row title 'Auto-lock Timeout'", () => {
        setUnlocked();
        renderSection();
        expect(screen.getByText("Auto-lock Timeout")).toBeInTheDocument();
      });

      it("renders row description for Auto-lock Timeout", () => {
        setUnlocked();
        renderSection();
        expect(
          screen.getByText(
            "Automatically lock the vault after this period of inactivity.",
          ),
        ).toBeInTheDocument();
      });

      it("renders row title 'Lock when hidden'", () => {
        setUnlocked();
        renderSection();
        expect(
          screen.getAllByText("Lock when hidden").length,
        ).toBeGreaterThanOrEqual(1);
      });

      it("renders a 'Lock' button with aria-label 'Lock vault'", () => {
        setUnlocked();
        renderSection();
        const btn = screen.getByRole("button", { name: "Lock vault" });
        expect(btn).toHaveTextContent("Lock");
      });

      it("renders row title 'Lock Vault'", () => {
        setUnlocked();
        renderSection();
        expect(screen.getByText("Lock Vault")).toBeInTheDocument();
      });

      it("renders row description for Lock Vault", () => {
        setUnlocked();
        renderSection();
        expect(
          screen.getByText("Clear decrypted keys from memory."),
        ).toBeInTheDocument();
      });
    });

    describe("Storage row (unit 5)", () => {
      it("renders 'Move' button label when storageMode is server", () => {
        vi.mocked(isVaultServerEnabled).mockReturnValue(true);
        useVaultStore.setState({ storageMode: "server" });
        setUnlocked();
        renderSection();

        const btn = screen.getByRole("button", {
          name: "Change vault storage location",
        });
        expect(btn).toHaveTextContent("Move");
      });

      it("renders 'Sync' button label when storageMode is local", () => {
        vi.mocked(isVaultServerEnabled).mockReturnValue(true);
        useVaultStore.setState({ storageMode: "local" });
        setUnlocked();
        renderSection();

        const btn = screen.getByRole("button", {
          name: "Change vault storage location",
        });
        expect(btn).toHaveTextContent("Sync");
      });

      it("renders description for server storageMode", () => {
        vi.mocked(isVaultServerEnabled).mockReturnValue(true);
        useVaultStore.setState({ storageMode: "server" });
        setUnlocked();
        renderSection();

        expect(
          screen.getByText(
            "Synced with the ShellHub server. Click to move it to this device.",
          ),
        ).toBeInTheDocument();
      });

      it("renders description for local storageMode", () => {
        vi.mocked(isVaultServerEnabled).mockReturnValue(true);
        useVaultStore.setState({ storageMode: "local" });
        setUnlocked();
        renderSection();

        expect(
          screen.getByText(
            "Stored in this browser only. Click to sync it to the ShellHub server.",
          ),
        ).toBeInTheDocument();
      });

      it("opens VaultSyncDialog when storage button is clicked", async () => {
        vi.mocked(isVaultServerEnabled).mockReturnValue(true);
        useVaultStore.setState({ storageMode: "local" });
        setUnlocked();
        renderSection();

        const btn = screen.getByRole("button", {
          name: "Change vault storage location",
        });
        await userEvent.click(btn);

        // VaultSyncDialog renders a heading or dialog when open
        expect(
          screen.getByRole("heading", { name: /sync/i }),
        ).toBeInTheDocument();
      });
    });

    describe("Danger Zone card (unit 5)", () => {
      it("renders a heading 'Danger Zone'", () => {
        setUnlocked();
        renderSection();
        expect(
          screen.getByRole("heading", { name: "Danger Zone" }),
        ).toBeInTheDocument();
      });

      it("renders row title 'Reset Vault'", () => {
        setUnlocked();
        renderSection();
        expect(screen.getByText("Reset Vault")).toBeInTheDocument();
      });

      it("renders row description for Reset Vault", () => {
        setUnlocked();
        renderSection();
        expect(
          screen.getByText(
            "Permanently delete all stored keys. This cannot be undone.",
          ),
        ).toBeInTheDocument();
      });

      it("renders a 'Reset' button with aria-label 'Reset vault'", () => {
        setUnlocked();
        renderSection();
        const btn = screen.getByRole("button", { name: "Reset vault" });
        expect(btn).toHaveTextContent("Reset");
      });
    });
  });
});
