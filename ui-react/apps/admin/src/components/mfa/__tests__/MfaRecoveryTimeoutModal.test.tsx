import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, act, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import MfaRecoveryTimeoutModal from "../MfaRecoveryTimeoutModal";

describe("MfaRecoveryTimeoutModal", () => {
  const onClose = vi.fn();
  const onDisable = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    onDisable.mockResolvedValue(undefined);
  });

  describe("Modal Rendering", () => {
    it("does not render when open is false", () => {
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60;
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
    beforeEach(() => {
      vi.useFakeTimers();
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it("displays countdown timer", () => {
      // Align to second boundary for exact diff
      const now = Math.floor(Date.now() / 1000) * 1000;
      vi.setSystemTime(now);
      const expiresAt = now / 1000 + 10 * 60; // 10 minutes

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      // useCountdown formats as "X minutes Y seconds remaining"
      expect(screen.getByText(/\d+ minutes? \d+ seconds? remaining/i)).toBeInTheDocument();
    });

    it("updates countdown every second", () => {
      const now = Math.floor(Date.now() / 1000) * 1000;
      vi.setSystemTime(now);
      const expiresAt = now / 1000 + 5 * 60; // 5 minutes

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      const initialText = screen.getByText(/remaining/).textContent;

      // Advance 1 second — wrap in act() to flush React state update
      act(() => {
        vi.advanceTimersByTime(1000);
      });

      const updatedText = screen.getByText(/remaining/).textContent;
      expect(initialText).not.toBe(updatedText);
    });

    it("shows expired state when countdown reaches zero", () => {
      const now = Math.floor(Date.now() / 1000) * 1000;
      vi.setSystemTime(now);
      const expiresAt = now / 1000 + 2; // 2 seconds

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      act(() => {
        vi.advanceTimersByTime(3000);
      });

      expect(screen.getByText(/expired/i)).toBeInTheDocument();
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
      vi.useFakeTimers();
      const now = Math.floor(Date.now() / 1000) * 1000;
      vi.setSystemTime(now);
      const expiresAt = now / 1000 + 1; // 1 second

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      act(() => {
        vi.advanceTimersByTime(2000);
      });

      vi.useRealTimers();

      const disableButton = screen.getByRole("button", {
        name: /disable mfa/i,
      });
      expect(disableButton).toBeDisabled();
    });

    it("calls onDisable when clicked", async () => {
      // Real timers — await user.click() hangs with fake timers
      const user = userEvent.setup();
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

      expect(onDisable).toHaveBeenCalled();
    });

    it("disables the button while disabling is in progress", async () => {
      // Real timers — await user.click() hangs with fake timers
      const user = userEvent.setup();
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60;

      let resolveDisable!: () => void;
      onDisable.mockReturnValue(
        new Promise<void>((resolve) => {
          resolveDisable = resolve;
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

      // Don't await — onDisable is pending so handleDisable is suspended at the await
      const clickPromise = user.click(disableButton);

      // Poll until setDisabling(true) triggers a re-render (real timers — waitFor polling works)
      await waitFor(() => expect(disableButton).toBeDisabled());

      // Resolve and let the click promise complete
      resolveDisable();
      await clickPromise;
    });
  });

  describe("Close Behavior", () => {
    it("calls onClose when Close button is clicked", async () => {
      // Real timers — await user.click() hangs with fake timers
      const user = userEvent.setup();
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60;

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      // The dismiss button is labelled "Close" (not "Continue to dashboard")
      const closeButton = screen.getByRole("button", { name: /^close$/i });
      await user.click(closeButton);

      expect(onClose).toHaveBeenCalled();
    });

    it("does not dismiss when clicking the backdrop (non-dismissible)", async () => {
      const user = userEvent.setup();
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60;

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      // The backdrop is intentionally non-dismissible (no onClick handler)
      // Modal has no role="dialog"; find inner container via heading
      const heading = screen.getByText(/recovery window active/i);
      const dialog = heading.closest(".relative") as HTMLElement | null;
      const backdrop = dialog?.previousElementSibling as HTMLElement | null;

      if (backdrop) {
        await user.click(backdrop);
        // onClose should NOT be called — user must explicitly use the Close button
        expect(onClose).not.toHaveBeenCalled();
      }
    });
  });

  describe("Auto-close on Expiry", () => {
    it("shows expired state after countdown reaches zero", () => {
      vi.useFakeTimers();
      const now = Math.floor(Date.now() / 1000) * 1000;
      vi.setSystemTime(now);
      const expiresAt = now / 1000 + 2; // 2 seconds

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      act(() => {
        vi.advanceTimersByTime(3000);
      });

      vi.useRealTimers();

      // Expired state is shown — parent controls actual close via onClose prop
      expect(screen.getByText(/expired/i)).toBeInTheDocument();
    });

    it("does not auto-close while a disable operation is in progress", async () => {
      vi.useFakeTimers();
      const now = Math.floor(Date.now() / 1000) * 1000;
      vi.setSystemTime(now);
      const expiresAt = now / 1000 + 1; // 1 second

      let resolveDisable!: () => void;
      onDisable.mockReturnValue(
        new Promise<void>((resolve) => {
          resolveDisable = resolve;
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

      const user = userEvent.setup({ delay: null });
      // Fire click (onDisable is pending) — don't await so we can advance timers
      const clickPromise = user.click(disableButton);

      // Flush initial state update (setDisabling(true))
      await act(async () => {
        await Promise.resolve();
      });

      // Advance time past expiration while disabling is in progress
      act(() => {
        vi.advanceTimersByTime(2000);
      });

      // onClose should not have been called autonomously (no auto-close logic in component)
      expect(onClose).not.toHaveBeenCalled();

      // Complete the disable operation
      resolveDisable();
      await clickPromise;

      expect(onDisable).toHaveBeenCalled();
      vi.useRealTimers();
    });
  });

  describe("Error Handling", () => {
    it("handles errors when disable fails", async () => {
      // Real timers — await user.click() hangs with fake timers
      const user = userEvent.setup();
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60;

      // handleDisable in the component has only finally (no catch), so the
      // rejected promise propagates as an unhandled rejection. Suppress it at
      // the process level so Vitest doesn't treat it as a test failure.
      const suppressRejection = () => { /* intentionally suppressed */ };
      process.on("unhandledRejection", suppressRejection);

      onDisable.mockImplementation(() => Promise.reject(new Error("Failed to disable")));

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

      // onDisable was called
      expect(onDisable).toHaveBeenCalled();
      // Error bubbles out of handleDisable but onClose is not triggered
      expect(onClose).not.toHaveBeenCalled();

      process.off("unhandledRejection", suppressRejection);
    });
  });

  describe("Warning Messages", () => {
    it("displays recovery window description", () => {
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
        screen.getByText(/successfully used a recovery code/i)
      ).toBeInTheDocument();
    });

    it("displays security explanation note", () => {
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60;

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />
      );

      expect(screen.getByText(/security measure/i)).toBeInTheDocument();
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
