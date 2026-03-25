import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useRef } from "react";
import "./helpers/setup-dialog";

vi.mock("@/hooks/useFocusTrap", () => ({
  useFocusTrap: vi.fn(),
}));

import CopyWarning, { CopyWarningHandle } from "../CopyWarning";
import { ClipboardProvider } from "../ClipboardProvider";

// ─── clipboard setup ──────────────────────────────────────────────────────────
//
// userEvent.setup() replaces navigator.clipboard with its own Clipboard stub,
// which intercepts writeText before our mock ever sees it.  Using fireEvent.click()
// instead avoids the hijack and lets us assert directly on clipboardWriteText.
//
// jsdom does not have a Clipboard API, so we install one on Navigator.prototype
// once for the whole file.  Tests configure resolve/reject via
// clipboardWriteText.mockResolvedValue / mockRejectedValue in their beforeEach.

const clipboardWriteText = vi.fn<() => Promise<void>>();

// userEvent.setup() can install its own Clipboard stub on the navigator
// *instance*, shadowing any prototype-level definition.  Re-install our mock
// on the instance before every test so it always takes precedence, and clean
// up after so the next test starts fresh.
beforeEach(() => {
  Object.defineProperty(navigator, "clipboard", {
    configurable: true,
    get: () => ({ writeText: clipboardWriteText }),
  });
});

afterEach(() => {
  // Remove the instance-level override so there's no stale property.
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  delete (navigator as any).clipboard;
  cleanup();
  vi.clearAllMocks();
});

// ─── helpers ──────────────────────────────────────────────────────────────────

function renderCopyWarning(
  {
    macro,
    bypass,
  }: {
    macro?: string;
    bypass?: boolean;
  } = {},
) {
  return render(
    <ClipboardProvider>
      <CopyWarning macro={macro} bypass={bypass}>
        {({ copy, copied }) => (
          <button
            data-testid="copy-btn"
            onClick={() => copy("test-text")}
          >
            {copied ? "Copied!" : "Copy"}
          </button>
        )}
      </CopyWarning>
    </ClipboardProvider>,
  );
}

// ─── rendering ───────────────────────────────────────────────────────────────

describe("CopyWarning", () => {
  describe("rendering", () => {
    it("renders the slot content", () => {
      renderCopyWarning();
      expect(screen.getByTestId("copy-btn")).toBeInTheDocument();
    });

    it("does not render the dialog initially", () => {
      renderCopyWarning();
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });

  // ─── secure context ──────────────────────────────────────────────────────

  describe("copy in a secure context", () => {
    beforeEach(() => {
      Object.defineProperty(globalThis, "isSecureContext", {
        configurable: true,
        writable: true,
        value: true,
      });
      clipboardWriteText.mockResolvedValue(undefined);
    });

    it("calls navigator.clipboard.writeText with the provided text", () => {
      renderCopyWarning();
      fireEvent.click(screen.getByTestId("copy-btn"));
      expect(clipboardWriteText).toHaveBeenCalledWith("test-text");
    });

    it("sets copied=true after a successful copy", async () => {
      renderCopyWarning();
      fireEvent.click(screen.getByTestId("copy-btn"));

      await waitFor(() =>
        expect(screen.getByTestId("copy-btn")).toHaveTextContent("Copied!"),
      );
    });

    it("resets copied back to false after 1500 ms", async () => {
      vi.useFakeTimers({ shouldAdvanceTime: true });
      renderCopyWarning();

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

    it("does not show the dialog on success", async () => {
      renderCopyWarning();
      fireEvent.click(screen.getByTestId("copy-btn"));

      await waitFor(() => expect(clipboardWriteText).toHaveBeenCalled());
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });

  // ─── insecure context ────────────────────────────────────────────────────

  describe("copy in an insecure context", () => {
    beforeEach(() => {
      Object.defineProperty(globalThis, "isSecureContext", {
        configurable: true,
        writable: true,
        value: false,
      });
    });

    it("shows the warning dialog instead of copying", () => {
      renderCopyWarning();
      fireEvent.click(screen.getByTestId("copy-btn"));

      expect(screen.getByRole("dialog")).toBeInTheDocument();
      expect(screen.getByText("Copying is not allowed")).toBeInTheDocument();
    });

    it("dialog contains the security explanation", () => {
      renderCopyWarning();
      fireEvent.click(screen.getByTestId("copy-btn"));

      expect(
        screen.getByText(
          /Clipboard access is only permitted on secure \(HTTPS\) or localhost origins/,
        ),
      ).toBeInTheDocument();
    });

    it("closes the dialog when OK is clicked", async () => {
      const user = userEvent.setup();
      renderCopyWarning();
      fireEvent.click(screen.getByTestId("copy-btn"));
      expect(screen.getByRole("dialog")).toBeInTheDocument();

      await user.click(screen.getByTestId("copy-warning-ok-btn"));
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });

    it("does not call clipboard.writeText", () => {
      renderCopyWarning();
      fireEvent.click(screen.getByTestId("copy-btn"));
      expect(clipboardWriteText).not.toHaveBeenCalled();
    });
  });

  // ─── clipboard API error ──────────────────────────────────────────────────

  describe("clipboard API error handling", () => {
    beforeEach(() => {
      Object.defineProperty(globalThis, "isSecureContext", {
        configurable: true,
        writable: true,
        value: true,
      });
      clipboardWriteText.mockRejectedValue(new Error("permission denied"));
    });

    it("shows the warning dialog when clipboard.writeText rejects", async () => {
      renderCopyWarning();
      fireEvent.click(screen.getByTestId("copy-btn"));

      await waitFor(() =>
        expect(screen.getByRole("dialog")).toBeInTheDocument(),
      );
      expect(screen.getByText("Copying is not allowed")).toBeInTheDocument();
    });
  });

  // ─── bypass prop ──────────────────────────────────────────────────────────

  describe("bypass prop", () => {
    beforeEach(() => {
      Object.defineProperty(globalThis, "isSecureContext", {
        configurable: true,
        writable: true,
        value: true,
      });
      clipboardWriteText.mockResolvedValue(undefined);
    });

    it("does not call clipboard.writeText when bypass=true", () => {
      renderCopyWarning({ bypass: true });
      fireEvent.click(screen.getByTestId("copy-btn"));
      expect(clipboardWriteText).not.toHaveBeenCalled();
    });

    it("does not show the dialog when bypass=true (insecure context)", () => {
      Object.defineProperty(globalThis, "isSecureContext", {
        configurable: true,
        writable: true,
        value: false,
      });
      renderCopyWarning({ bypass: true });
      fireEvent.click(screen.getByTestId("copy-btn"));
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });

  // ─── accessibility ────────────────────────────────────────────────────────

  describe("accessibility", () => {
    beforeEach(() => {
      Object.defineProperty(globalThis, "isSecureContext", {
        configurable: true,
        writable: true,
        value: false,
      });
    });

    it("dialog has aria-labelledby pointing to the title", () => {
      renderCopyWarning();
      fireEvent.click(screen.getByTestId("copy-btn"));

      const dialog = screen.getByRole("dialog");
      const labelId = dialog.getAttribute("aria-labelledby");
      expect(labelId).toBeTruthy();
      expect(document.getElementById(labelId!)).toHaveTextContent(
        "Copying is not allowed",
      );
    });

    it("dialog has aria-describedby pointing to the description", () => {
      renderCopyWarning();
      fireEvent.click(screen.getByTestId("copy-btn"));

      const dialog = screen.getByRole("dialog");
      const descId = dialog.getAttribute("aria-describedby");
      expect(descId).toBeTruthy();
      expect(document.getElementById(descId!)).toHaveTextContent(
        /Clipboard access is only permitted/,
      );
    });
  });

  // ─── keyboard shortcut (macro) ────────────────────────────────────────────

  describe("keyboard shortcut (macro prop)", () => {
    beforeEach(() => {
      Object.defineProperty(globalThis, "isSecureContext", {
        configurable: true,
        writable: true,
        value: true,
      });
      clipboardWriteText.mockResolvedValue(undefined);
    });

    it("copies the macro text when Ctrl+C is pressed", async () => {
      renderCopyWarning({ macro: "ssh root@device" });

      act(() => {
        document.dispatchEvent(
          new KeyboardEvent("keydown", {
            key: "c",
            ctrlKey: true,
            bubbles: true,
            cancelable: true,
          }),
        );
      });

      await waitFor(() =>
        expect(clipboardWriteText).toHaveBeenCalledWith("ssh root@device"),
      );
    });

    it("calls e.preventDefault() on Ctrl+C when macro is set", async () => {
      renderCopyWarning({ macro: "ssh root@device" });

      const event = new KeyboardEvent("keydown", {
        key: "c",
        ctrlKey: true,
        bubbles: true,
        cancelable: true,
      });
      const preventDefaultSpy = vi.spyOn(event, "preventDefault");

      await act(() => document.dispatchEvent(event));

      expect(preventDefaultSpy).toHaveBeenCalled();
    });

    it("does not copy on keyup (only keydown)", () => {
      renderCopyWarning({ macro: "ssh root@device" });

      act(() => {
        document.dispatchEvent(
          new KeyboardEvent("keyup", { key: "c", ctrlKey: true, bubbles: true }),
        );
      });

      expect(clipboardWriteText).not.toHaveBeenCalled();
    });

    it("does not copy when Ctrl is not held", () => {
      renderCopyWarning({ macro: "ssh root@device" });

      act(() => {
        document.dispatchEvent(
          new KeyboardEvent("keydown", { key: "c", ctrlKey: false, bubbles: true }),
        );
      });

      expect(clipboardWriteText).not.toHaveBeenCalled();
    });

    it("does not register a keydown listener when macro is not provided", () => {
      const addSpy = vi.spyOn(document, "addEventListener");

      renderCopyWarning();

      const keydownCalls = addSpy.mock.calls.filter(([type]) => type === "keydown");
      expect(keydownCalls).toHaveLength(0);
    });

    it("skips Ctrl+C copy when bypass=true", () => {
      renderCopyWarning({ macro: "ssh root@device", bypass: true });

      act(() => {
        document.dispatchEvent(
          new KeyboardEvent("keydown", { key: "c", ctrlKey: true, bubbles: true }),
        );
      });

      expect(clipboardWriteText).not.toHaveBeenCalled();
    });
  });

  // ─── imperative handle (ref) ──────────────────────────────────────────────

  describe("imperative handle (ref)", () => {
    beforeEach(() => {
      Object.defineProperty(globalThis, "isSecureContext", {
        configurable: true,
        writable: true,
        value: true,
      });
      clipboardWriteText.mockResolvedValue(undefined);
    });

    it("exposes copyFn that triggers clipboard write", async () => {
      const handleRef = { current: null as CopyWarningHandle | null };

      function Wrapper() {
        const ref = useRef<CopyWarningHandle>(null);
        return (
          <ClipboardProvider>
            <CopyWarning ref={ref}>
              {() => (
                <button
                  data-testid="setup-btn"
                  onClick={() => { handleRef.current = ref.current; }}
                >
                  setup
                </button>
              )}
            </CopyWarning>
          </ClipboardProvider>
        );
      }

      render(<Wrapper />);
      // Use fireEvent to populate the ref without clipboard interference.
      fireEvent.click(screen.getByTestId("setup-btn"));
      expect(handleRef.current).not.toBeNull();

      act(() => { handleRef.current?.copyFn("imperative-text"); });

      await waitFor(() =>
        expect(clipboardWriteText).toHaveBeenCalledWith("imperative-text"),
      );
    });
  });
});
