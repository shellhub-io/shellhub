import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useVaultStore } from "@/stores/vaultStore";
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

vi.mock("@/stores/authStore", () => ({
  useAuthStore: {
    getState: vi.fn(() => ({ user: "testuser", tenant: "test-tenant" })),
  },
}));

vi.mock("@/utils/vault-activity-tracker", () => ({
  start: vi.fn(),
  stop: vi.fn(),
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
    autoLockNonce: 0,
  });
  vi.clearAllMocks();
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
});
