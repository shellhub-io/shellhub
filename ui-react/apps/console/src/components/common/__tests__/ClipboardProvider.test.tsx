import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "./helpers/setup-dialog";

vi.mock("@/hooks/useFocusTrap", () => ({
  useFocusTrap: vi.fn(),
}));

import { ClipboardProvider } from "../ClipboardProvider";
import { useCopy } from "@/hooks/useCopy";

// ─── clipboard setup ──────────────────────────────────────────────────────────

const clipboardWriteText = vi.fn<() => Promise<void>>();

beforeEach(() => {
  Object.defineProperty(navigator, "clipboard", {
    configurable: true,
    get: () => ({ writeText: clipboardWriteText }),
  });
});

afterEach(() => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  delete (navigator as any).clipboard;
  cleanup();
  vi.clearAllMocks();
});

// ─── helpers ──────────────────────────────────────────────────────────────────

function CopyConsumer({ text = "test-text" }: { text?: string }) {
  const { copy, copied } = useCopy();
  return (
    <button data-testid="copy-btn" onClick={() => copy(text)}>
      {copied ? "Copied!" : "Copy"}
    </button>
  );
}

function renderWithProvider(text?: string) {
  return render(
    <ClipboardProvider>
      <CopyConsumer text={text} />
    </ClipboardProvider>,
  );
}

// ─── useCopy ─────────────────────────────────────────────────────────────────

describe("ClipboardProvider / useCopy", () => {
  describe("throws outside provider", () => {
    it("throws when used outside ClipboardProvider", () => {
      const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      expect(() => render(<CopyConsumer />)).toThrow(
        "useCopy must be used within <ClipboardProvider>",
      );
      consoleSpy.mockRestore();
    });
  });

  // ─── secure context ────────────────────────────────────────────────────────

  describe("secure context", () => {
    beforeEach(() => {
      Object.defineProperty(globalThis, "isSecureContext", {
        configurable: true,
        writable: true,
        value: true,
      });
      clipboardWriteText.mockResolvedValue(undefined);
    });

    it("calls navigator.clipboard.writeText with the given text", () => {
      renderWithProvider("hello");
      fireEvent.click(screen.getByTestId("copy-btn"));
      expect(clipboardWriteText).toHaveBeenCalledWith("hello");
    });

    it("sets copied=true after a successful copy", async () => {
      renderWithProvider();
      fireEvent.click(screen.getByTestId("copy-btn"));
      await waitFor(() =>
        expect(screen.getByTestId("copy-btn")).toHaveTextContent("Copied!"),
      );
    });

    it("resets copied to false after 1500 ms", async () => {
      vi.useFakeTimers({ shouldAdvanceTime: true });
      renderWithProvider();

      fireEvent.click(screen.getByTestId("copy-btn"));
      await waitFor(() =>
        expect(screen.getByTestId("copy-btn")).toHaveTextContent("Copied!"),
      );

      await act(() => vi.advanceTimersByTime(1500));
      await waitFor(() =>
        expect(screen.getByTestId("copy-btn")).toHaveTextContent("Copy"),
      );

      vi.useRealTimers();
    });

    it("does not open the dialog on successful copy", async () => {
      renderWithProvider();
      fireEvent.click(screen.getByTestId("copy-btn"));
      await waitFor(() => expect(clipboardWriteText).toHaveBeenCalled());
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });

  // ─── insecure context ──────────────────────────────────────────────────────

  describe("insecure context", () => {
    beforeEach(() => {
      Object.defineProperty(globalThis, "isSecureContext", {
        configurable: true,
        writable: true,
        value: false,
      });
    });

    it("does not call clipboard.writeText", () => {
      renderWithProvider();
      fireEvent.click(screen.getByTestId("copy-btn"));
      expect(clipboardWriteText).not.toHaveBeenCalled();
    });

    it("opens the warning dialog", () => {
      renderWithProvider();
      fireEvent.click(screen.getByTestId("copy-btn"));
      expect(screen.getByRole("dialog")).toBeInTheDocument();
      expect(screen.getByText("Copying is not allowed")).toBeInTheDocument();
    });

    it("closes the dialog when OK is clicked", async () => {
      const user = userEvent.setup();
      renderWithProvider();
      fireEvent.click(screen.getByTestId("copy-btn"));
      expect(screen.getByRole("dialog")).toBeInTheDocument();

      await user.click(screen.getByTestId("copy-warning-ok-btn"));
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });

    it("only opens the dialog once even after multiple copy calls", () => {
      renderWithProvider();
      fireEvent.click(screen.getByTestId("copy-btn"));
      fireEvent.click(screen.getByTestId("copy-btn"));
      expect(screen.getAllByRole("dialog")).toHaveLength(1);
    });
  });

  // ─── clipboard API error ──────────────────────────────────────────────────

  describe("clipboard API error", () => {
    beforeEach(() => {
      Object.defineProperty(globalThis, "isSecureContext", {
        configurable: true,
        writable: true,
        value: true,
      });
      clipboardWriteText.mockRejectedValue(new Error("permission denied"));
    });

    it("opens the warning dialog when writeText rejects", async () => {
      renderWithProvider();
      fireEvent.click(screen.getByTestId("copy-btn"));
      await waitFor(() =>
        expect(screen.getByRole("dialog")).toBeInTheDocument(),
      );
    });
  });

  // ─── shared dialog (multiple consumers) ───────────────────────────────────

  describe("shared dialog across multiple consumers", () => {
    beforeEach(() => {
      Object.defineProperty(globalThis, "isSecureContext", {
        configurable: true,
        writable: true,
        value: false,
      });
    });

    it("uses a single dialog instance for multiple useCopy consumers", () => {
      render(
        <ClipboardProvider>
          <CopyConsumer />
          <CopyConsumer />
        </ClipboardProvider>,
      );

      const [btn1] = screen.getAllByTestId("copy-btn");
      fireEvent.click(btn1);

      expect(screen.getAllByRole("dialog")).toHaveLength(1);
    });
  });
});
