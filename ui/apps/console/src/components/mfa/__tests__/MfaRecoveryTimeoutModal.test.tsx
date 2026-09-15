import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  render,
  screen,
  act,
  waitFor,
  fireEvent,
} from "@testing-library/react";
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
        />,
      );

      expect(container.firstChild).toBeNull();
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
      const now = Math.floor(Date.now() / 1000) * 1000;
      vi.setSystemTime(now);
      const expiresAt = now / 1000 + 10 * 60; // 10 minutes

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />,
      );

      expect(
        screen.getByText(/\d+ minutes? \d+ seconds? remaining/i),
      ).toBeInTheDocument();
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
        />,
      );

      const initialText = screen.getByText(/remaining/).textContent;

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
        />,
      );

      act(() => {
        vi.advanceTimersByTime(3000);
      });

      expect(screen.getByText(/expired/i)).toBeInTheDocument();
    });
  });

  describe("Disable Button", () => {
    it("is enabled while the countdown runs and disabled once it expires", () => {
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
        />,
      );

      const disableButton = screen.getByRole("button", {
        name: /disable mfa/i,
      });
      expect(disableButton).toBeEnabled();

      act(() => {
        vi.advanceTimersByTime(2000);
      });

      vi.useRealTimers();

      expect(disableButton).toBeDisabled();
    });

    it("calls onDisable when clicked", async () => {
      const user = userEvent.setup();
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60;

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />,
      );

      const disableButton = screen.getByRole("button", {
        name: /disable mfa/i,
      });
      await user.click(disableButton);

      expect(onDisable).toHaveBeenCalled();
    });

    it("disables the button while disabling is in progress", async () => {
      const user = userEvent.setup();
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60;

      let resolveDisable!: () => void;
      onDisable.mockReturnValue(
        new Promise<void>((resolve) => {
          resolveDisable = resolve;
        }),
      );

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />,
      );

      const disableButton = screen.getByRole("button", {
        name: /disable mfa/i,
      });

      const clickPromise = user.click(disableButton);

      await waitFor(() => expect(disableButton).toBeDisabled());

      resolveDisable();
      await clickPromise;
    });
  });

  describe("Close Behavior", () => {
    it("calls onClose when Close button is clicked", async () => {
      const user = userEvent.setup();
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60;

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />,
      );

      const closeButton = screen.getByRole("button", { name: /^close$/i });
      await user.click(closeButton);

      expect(onClose).toHaveBeenCalled();
    });

    it("does not dismiss when clicking the backdrop (non-dismissible)", () => {
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60;

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />,
      );

      const dialog = document.querySelector("dialog") as HTMLElement;
      fireEvent(dialog, new Event("cancel"));
      expect(onClose).not.toHaveBeenCalled();
    });
  });

  describe("Auto-close on Expiry", () => {
    it("does not auto-close while a disable operation is in progress", async () => {
      vi.useFakeTimers();
      const now = Math.floor(Date.now() / 1000) * 1000;
      vi.setSystemTime(now);
      const expiresAt = now / 1000 + 1; // 1 second

      let resolveDisable!: () => void;
      onDisable.mockReturnValue(
        new Promise<void>((resolve) => {
          resolveDisable = resolve;
        }),
      );

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />,
      );

      const disableButton = screen.getByRole("button", {
        name: /disable mfa/i,
      });

      const user = userEvent.setup({ delay: null });
      const clickPromise = user.click(disableButton);

      await act(async () => {
        await Promise.resolve();
      });

      act(() => {
        vi.advanceTimersByTime(2000);
      });

      expect(onClose).not.toHaveBeenCalled();

      resolveDisable();
      await clickPromise;

      expect(onDisable).toHaveBeenCalled();
      vi.useRealTimers();
    });
  });

  describe("Error Handling", () => {
    it("handles errors when disable fails", async () => {
      const user = userEvent.setup();
      const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60;

      const suppressRejection = () => {};
      process.on("unhandledRejection", suppressRejection);

      onDisable.mockImplementation(() =>
        Promise.reject(new Error("Failed to disable")),
      );

      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />,
      );

      const disableButton = screen.getByRole("button", {
        name: /disable mfa/i,
      });
      await user.click(disableButton);

      expect(onDisable).toHaveBeenCalled();
      expect(onClose).not.toHaveBeenCalled();

      process.off("unhandledRejection", suppressRejection);
    });
  });

  describe("Invalid Timestamp", () => {
    it.each([NaN, 0])("still renders when expiresAt is %p", (expiresAt) => {
      render(
        <MfaRecoveryTimeoutModal
          open={true}
          expiresAt={expiresAt}
          onClose={onClose}
          onDisable={onDisable}
        />,
      );

      expect(screen.getByText(/recovery window/i)).toBeInTheDocument();
    });
  });
});
