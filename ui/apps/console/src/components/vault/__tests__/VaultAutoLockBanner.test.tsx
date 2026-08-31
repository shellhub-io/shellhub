import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, act } from "@testing-library/react";
import { useVaultStore } from "@/stores/vaultStore";
import VaultAutoLockBanner from "../VaultAutoLockBanner";

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

beforeEach(() => {
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

afterEach(() => {
  vi.useRealTimers();
});

describe("VaultAutoLockBanner", () => {
  describe("initial render (pre-existing nonce)", () => {
    it("does NOT show the toast on mount when autoLockNonce is 0", () => {
      useVaultStore.setState({ autoLockNonce: 0 });
      render(<VaultAutoLockBanner />);
      expect(
        screen.queryByText(/vault locked due to inactivity/i),
      ).not.toBeInTheDocument();
    });

    it("does NOT show the toast on mount even when autoLockNonce is already > 0 (pre-existing)", () => {
      useVaultStore.setState({ autoLockNonce: 3 });
      render(<VaultAutoLockBanner />);
      expect(
        screen.queryByText(/vault locked due to inactivity/i),
      ).not.toBeInTheDocument();
    });
  });

  describe("nonce bump after mount", () => {
    it("shows the toast when autoLockNonce is bumped after mount", () => {
      useVaultStore.setState({ autoLockNonce: 0 });
      render(<VaultAutoLockBanner />);

      act(() => {
        useVaultStore.setState({ autoLockNonce: 1 });
      });

      expect(
        screen.getByText(/vault locked due to inactivity/i),
      ).toBeInTheDocument();
    });

    it("shows the toast on a second nonce bump (multiple lock events)", () => {
      vi.useFakeTimers();

      useVaultStore.setState({ autoLockNonce: 0 });
      render(<VaultAutoLockBanner />);

      act(() => {
        useVaultStore.setState({ autoLockNonce: 1 });
      });

      expect(
        screen.getByText(/vault locked due to inactivity/i),
      ).toBeInTheDocument();

      act(() => {
        vi.advanceTimersByTime(6500);
      });

      expect(
        screen.queryByText(/vault locked due to inactivity/i),
      ).not.toBeInTheDocument();

      act(() => {
        useVaultStore.setState({ autoLockNonce: 2 });
      });

      expect(
        screen.getByText(/vault locked due to inactivity/i),
      ).toBeInTheDocument();
    });
  });

  describe("dismiss behavior", () => {
    it("hides the toast when the dismiss button is clicked", () => {
      useVaultStore.setState({ autoLockNonce: 0 });
      render(<VaultAutoLockBanner />);

      act(() => {
        useVaultStore.setState({ autoLockNonce: 1 });
      });

      expect(
        screen.getByText(/vault locked due to inactivity/i),
      ).toBeInTheDocument();

      act(() => {
        screen.getByRole("button", { name: /dismiss/i }).click();
      });

      expect(
        screen.queryByText(/vault locked due to inactivity/i),
      ).not.toBeInTheDocument();
    });
  });

  describe("auto-dismiss", () => {
    it("hides the toast automatically after ~6 seconds", () => {
      vi.useFakeTimers();

      useVaultStore.setState({ autoLockNonce: 0 });
      render(<VaultAutoLockBanner />);

      act(() => {
        useVaultStore.setState({ autoLockNonce: 1 });
      });

      expect(
        screen.getByText(/vault locked due to inactivity/i),
      ).toBeInTheDocument();

      act(() => {
        vi.advanceTimersByTime(6100);
      });

      expect(
        screen.queryByText(/vault locked due to inactivity/i),
      ).not.toBeInTheDocument();
    });

    it("does NOT hide the toast before ~6 seconds elapse", () => {
      vi.useFakeTimers();

      useVaultStore.setState({ autoLockNonce: 0 });
      render(<VaultAutoLockBanner />);

      act(() => {
        useVaultStore.setState({ autoLockNonce: 1 });
      });

      act(() => {
        vi.advanceTimersByTime(5000);
      });

      expect(
        screen.getByText(/vault locked due to inactivity/i),
      ).toBeInTheDocument();
    });

    it("clears the auto-dismiss timer on unmount (no state-update after unmount)", () => {
      vi.useFakeTimers();

      useVaultStore.setState({ autoLockNonce: 0 });
      const { unmount } = render(<VaultAutoLockBanner />);

      act(() => {
        useVaultStore.setState({ autoLockNonce: 1 });
      });

      unmount();

      expect(() => {
        act(() => {
          vi.advanceTimersByTime(7000);
        });
      }).not.toThrow();
    });
  });

  describe("positioning and z-index", () => {
    it("renders inside a fixed container with correct positioning classes when visible", () => {
      useVaultStore.setState({ autoLockNonce: 0 });
      const { container } = render(<VaultAutoLockBanner />);

      act(() => {
        useVaultStore.setState({ autoLockNonce: 1 });
      });

      const wrapper = container.firstChild as HTMLElement;
      expect(wrapper).toHaveClass("fixed");
      expect(wrapper).toHaveClass("bottom-4");
      expect(wrapper).toHaveClass("right-4");
      expect(wrapper).toHaveClass("z-toast");
    });
  });
});
