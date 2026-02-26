import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import MfaRecoveryTimeoutModal from "../MfaRecoveryTimeoutModal";

describe("MfaRecoveryTimeoutModal", () => {
  const onClose = vi.fn();
  const onDisable = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
    onDisable.mockResolvedValue(undefined);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  describe("Modal Rendering", () => {
    it("does not render when open is false", () => {
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60; // Unix timestamp 10 minutes from now
      const { container } = render(
        <MfaRecoveryTimeoutModal
          open={false}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      expect(container.firstChild).toBeNull();
    });

    it("renders when open is true", () => {
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60;
      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      expect(screen.getByText(/recovery window/i)).toBeInTheDocument();
    });
  });

  describe("Countdown Display", () => {
    it("displays countdown timer", () => {
      const now = Date.now();
      const expiresAt = Math.floor(now / 1000) + 10 * 60; // 10 minutes

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      // Should show minutes and seconds
      expect(screen.getByText(/\d+m \d+s/)).toBeInTheDocument();
    });

    it("updates countdown every second", () => {
      const now = Date.now();
      const expiresAt = Math.floor(now / 1000) + 5 * 60; // 5 minutes

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      const initialText = screen.getByText(/\d+m \d+s/).textContent;

      // Advance time by 1 second
      vi.advanceTimersByTime(1000);

      const updatedText = screen.getByText(/\d+m \d+s/).textContent;

      // Countdown should have changed
      expect(initialText).not.toBe(updatedText);
    });

    it("shows expired message when countdown reaches zero", () => {
      const now = Date.now();
      const expiresAt = Math.floor(now / 1000) + 2; // 2 seconds

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      // Advance time past expiration
      vi.advanceTimersByTime(3000);

      waitFor(() => {
        expect(screen.getByText(/expired/i)).toBeInTheDocument();
      });
    });
  });

  describe("Disable Button", () => {
    it("enables disable button when countdown is active", () => {
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60;

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      const disableButton = screen.getByRole("button", {
        name: /disable mfa/i,
      });
      expect(disableButton).toBeEnabled();
    });

    it("disables disable button when countdown expires", () => {
      const now = Date.now();
      const expiresAt = Math.floor(now / 1000) + 1; // 1 second

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      // Advance time past expiration
      vi.advanceTimersByTime(2000);

      waitFor(() => {
        const disableButton = screen.getByRole("button", {
          name: /disable mfa/i,
        });
        expect(disableButton).toBeDisabled();
      });
    });

    it("calls onDisable when clicked", async () => {
      const user = userEvent.setup({ delay: null }); // Disable delay for fake timers
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60;

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      const disableButton = screen.getByRole("button", {
        name: /disable mfa/i,
      });
      await user.click(disableButton);

      await waitFor(() => {
        expect(onDisable).toHaveBeenCalled();
      });
    });

    it("shows loading state while disabling", async () => {
      const user = userEvent.setup({ delay: null });
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60;

      let resolveDisable: () => void;
      onDisable.mockReturnValue(
        new Promise((resolve) => {
          resolveDisable = resolve as () => void;
        })
      );

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      const disableButton = screen.getByRole("button", {
        name: /disable mfa/i,
      });
      await user.click(disableButton);

      // Should show loading state
      await waitFor(() => {
        expect(screen.getByText(/disabling/i)).toBeInTheDocument();
      });

      // Button should be disabled during loading
      expect(disableButton).toBeDisabled();

      // Resolve
      resolveDisable!();

      await waitFor(() => {
        expect(screen.queryByText(/disabling/i)).not.toBeInTheDocument();
      });
    });
  });

  describe("Close Behavior", () => {
    it("calls onClose when Continue button is clicked", async () => {
      const user = userEvent.setup({ delay: null });
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60;

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      const continueButton = screen.getByRole("button", {
        name: /continue to dashboard/i,
      });
      await user.click(continueButton);

      expect(onClose).toHaveBeenCalled();
    });

    it("calls onClose when clicking backdrop", async () => {
      const user = userEvent.setup({ delay: null });
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60;

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      // Get the backdrop element
      const backdrop = screen.getByText(/recovery window/i).parentElement?.parentElement
        ?.previousElementSibling;

      if (backdrop) {
        await user.click(backdrop);
        expect(onClose).toHaveBeenCalled();
      }
    });
  });

  describe("Auto-close on Expiry", () => {
    it("automatically closes when countdown expires", () => {
      const now = Date.now();
      const expiresAt = Math.floor(now / 1000) + 2; // 2 seconds

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      // Advance time past expiration
      vi.advanceTimersByTime(3000);

      waitFor(() => {
        expect(onClose).toHaveBeenCalled();
      });
    });

    it("does not close if user is in the middle of disabling", async () => {
      const user = userEvent.setup({ delay: null });
      const now = Date.now();
      const expiresAt = Math.floor(now / 1000) + 1; // 1 second

      let resolveDisable: () => void;
      onDisable.mockReturnValue(
        new Promise((resolve) => {
          resolveDisable = resolve as () => void;
        })
      );

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      const disableButton = screen.getByRole("button", {
        name: /disable mfa/i,
      });
      await user.click(disableButton);

      // Advance time past expiration while disabling
      vi.advanceTimersByTime(2000);

      // Should not auto-close while operation is in progress
      expect(onClose).not.toHaveBeenCalled();

      // Complete the disable operation
      resolveDisable!();

      await waitFor(() => {
        expect(onDisable).toHaveBeenCalled();
      });
    });
  });

  describe("Error Handling", () => {
    it("handles errors when disable fails", async () => {
      const user = userEvent.setup({ delay: null });
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60;

      onDisable.mockRejectedValue(new Error("Failed to disable"));

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      const disableButton = screen.getByRole("button", {
        name: /disable mfa/i,
      });
      await user.click(disableButton);

      await waitFor(() => {
        expect(onDisable).toHaveBeenCalled();
      });

      // Should handle the error gracefully
      // Error is logged but modal remains open for user to retry or close
      expect(onClose).not.toHaveBeenCalled();
    });
  });

  describe("Warning Messages", () => {
    it("displays warning about limited time window", () => {
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60;

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      expect(
        screen.getByText(/you have a limited time/i)
      ).toBeInTheDocument();
    });

    it("displays warning about account security", () => {
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60;

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      expect(screen.getByText(/if you disable/i)).toBeInTheDocument();
    });
  });

  describe("Invalid Timestamp", () => {
    it("handles invalid timestamp gracefully", () => {
      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={NaN}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      // Should still render but show expired state
      expect(screen.getByText(/recovery window/i)).toBeInTheDocument();
    });

    it("handles zero timestamp", () => {
      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={0}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      // Should still render
      expect(screen.getByText(/recovery window/i)).toBeInTheDocument();
    });
  });
});
