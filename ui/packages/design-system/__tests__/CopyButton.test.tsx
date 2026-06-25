import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, act } from "@testing-library/react";
import { CopyButton } from "../components/CopyButton";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function stubClipboard(resolves: boolean) {
  const writeText = resolves
    ? vi.fn().mockResolvedValue(undefined)
    : vi.fn().mockRejectedValue(new Error("clipboard denied"));

  Object.defineProperty(navigator, "clipboard", {
    value: { writeText },
    writable: true,
    configurable: true,
  });

  return writeText;
}

// ---------------------------------------------------------------------------
// (a) Clipboard icon at rest → check after click → reverts after 1500ms
// ---------------------------------------------------------------------------
describe("CopyButton — copy lifecycle (secure context)", () => {
  beforeEach(() => {
    Object.defineProperty(globalThis, "isSecureContext", {
      value: true,
      writable: true,
      configurable: true,
    });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("renders clipboard icon at rest, swaps to check after click, reverts after 1500ms", async () => {
    vi.useFakeTimers();
    const writeText = stubClipboard(true);

    render(<CopyButton text="hello" />);

    // At rest: clipboard icon present (no check icon)
    const btn = screen.getByTitle("Copy");
    const clipboardPath = btn.querySelector("path[d*='M15.666']");
    expect(clipboardPath).not.toBeNull();

    // Click — triggers copy; wrap in act so the resolved promise + state update flush
    await act(async () => {
      fireEvent.click(btn);
    });
    expect(writeText).toHaveBeenCalledWith("hello");

    // After click: check icon appears
    const checkPath = btn.querySelector("path[d*='m4.5 12.75']");
    expect(checkPath).not.toBeNull();

    // After 1500ms: reverts back to clipboard icon
    act(() => {
      vi.advanceTimersByTime(1500);
    });
    const revertedClipboardPath = btn.querySelector("path[d*='M15.666']");
    expect(revertedClipboardPath).not.toBeNull();
  });
});

// ---------------------------------------------------------------------------
// (b) onError invoked when writeText rejects
// ---------------------------------------------------------------------------
describe("CopyButton — onError on writeText rejection", () => {
  beforeEach(() => {
    Object.defineProperty(globalThis, "isSecureContext", {
      value: true,
      writable: true,
      configurable: true,
    });
  });

  it("calls onError when navigator.clipboard.writeText rejects", async () => {
    stubClipboard(false);
    const onError = vi.fn();

    render(<CopyButton text="hello" onError={onError} />);
    fireEvent.click(screen.getByTitle("Copy"));

    // Flush the rejected promise
    await vi.waitFor(() => expect(onError).toHaveBeenCalledOnce());
  });
});

// ---------------------------------------------------------------------------
// (c) onError invoked when isSecureContext is false
// ---------------------------------------------------------------------------
describe("CopyButton — onError when not in secure context", () => {
  beforeEach(() => {
    Object.defineProperty(globalThis, "isSecureContext", {
      value: false,
      writable: true,
      configurable: true,
    });
  });

  it("calls onError immediately when isSecureContext is false", () => {
    const onError = vi.fn();

    render(<CopyButton text="hello" onError={onError} />);
    fireEvent.click(screen.getByTitle("Copy"));

    expect(onError).toHaveBeenCalledOnce();
  });
});
