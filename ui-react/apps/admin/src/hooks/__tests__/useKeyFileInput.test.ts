import React from "react";
import { describe, it, expect, vi } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useKeyFileInput } from "../useKeyFileInput";

function makeFile(name: string, content: string): File {
  return new File([content], name, { type: "text/plain" });
}

function makeOversizedFile(): File {
  // 513 KB — just above the 512 KB limit
  const blob = new Blob([new Uint8Array(513 * 1024)]);
  return new File([blob], "big.pem", { type: "text/plain" });
}

/** Simulate the FileReader reading a File synchronously via readAsText mock. */
function mockFileReader(content: string) {
  const originalFileReader = globalThis.FileReader;

  class MockFileReader extends EventTarget {
    result: string | null = null;
    onload: (() => void) | null = null;

    readAsText(_file: File) {
      // Simulate async read with a microtask
      Promise.resolve().then(() => {
        this.result = content;
        if (this.onload) this.onload();
      });
    }
  }

  // @ts-expect-error - partial mock
  globalThis.FileReader = MockFileReader;

  return () => {
    globalThis.FileReader = originalFileReader;
  };
}

function makeDropEvent(file: File): React.DragEvent<Element> {
  return {
    preventDefault: vi.fn(),
    dataTransfer: { files: [file] },
  } as unknown as React.DragEvent<Element>;
}

describe("useKeyFileInput", () => {
  describe("inputMode initialisation", () => {
    it('starts in "file" mode when not disabled', () => {
      const { result } = renderHook(() =>
        useKeyFileInput({
          validate: () => true,
          onChange: vi.fn(),
          disabled: false,
        }),
      );
      expect(result.current.inputMode).toBe("file");
    });

    it('starts in "text" mode when disabled', () => {
      const { result } = renderHook(() =>
        useKeyFileInput({
          validate: () => true,
          onChange: vi.fn(),
          disabled: true,
        }),
      );
      expect(result.current.inputMode).toBe("text");
    });
  });

  describe("setInputMode", () => {
    it('switches to "text" mode on demand', () => {
      const { result } = renderHook(() =>
        useKeyFileInput({ validate: () => true, onChange: vi.fn() }),
      );
      act(() => result.current.setInputMode("text"));
      expect(result.current.inputMode).toBe("text");
    });

    it('switches back to "file" mode on demand', () => {
      const { result } = renderHook(() =>
        useKeyFileInput({ validate: () => true, onChange: vi.fn() }),
      );
      act(() => result.current.setInputMode("text"));
      act(() => result.current.setInputMode("file"));
      expect(result.current.inputMode).toBe("file");
    });
  });

  describe("dragging state", () => {
    it("starts with dragging = false", () => {
      const { result } = renderHook(() =>
        useKeyFileInput({ validate: () => true, onChange: vi.fn() }),
      );
      expect(result.current.dragging).toBe(false);
    });

    it("setDragging(true) updates dragging state", () => {
      const { result } = renderHook(() =>
        useKeyFileInput({ validate: () => true, onChange: vi.fn() }),
      );
      act(() => result.current.setDragging(true));
      expect(result.current.dragging).toBe(true);
    });

    it("handleDrop resets dragging to false", () => {
      const restore = mockFileReader("key content");
      const onChange = vi.fn();
      const { result } = renderHook(() =>
        useKeyFileInput({ validate: () => true, onChange }),
      );
      act(() => result.current.setDragging(true));
      const file = makeFile("id_rsa.pub", "key content");
      act(() => result.current.handleDrop(makeDropEvent(file)));
      expect(result.current.dragging).toBe(false);
      restore();
    });
  });

  describe("processFile", () => {
    it("calls onChange with file text content", async () => {
      const restore = mockFileReader("ssh-rsa AAAAB3NzaC1");
      const onChange = vi.fn();
      const { result } = renderHook(() =>
        useKeyFileInput({ validate: () => true, onChange }),
      );
      await act(async () => {
        result.current.processFile(makeFile("id_rsa.pub", "ssh-rsa AAAAB3NzaC1"));
      });
      expect(onChange).toHaveBeenCalledWith("ssh-rsa AAAAB3NzaC1");
      restore();
    });

    it("calls onFileName with the base name (no extension)", async () => {
      const restore = mockFileReader("key content");
      const onChange = vi.fn();
      const onFileName = vi.fn();
      const { result } = renderHook(() =>
        useKeyFileInput({ validate: () => true, onChange, onFileName }),
      );
      await act(async () => {
        result.current.processFile(makeFile("my-key.pem", "key content"));
      });
      expect(onFileName).toHaveBeenCalledWith("my-key");
      restore();
    });

    it("does NOT call onFileName when the prop is omitted", async () => {
      const restore = mockFileReader("key content");
      const onChange = vi.fn();
      const { result } = renderHook(() =>
        useKeyFileInput({ validate: () => true, onChange }),
      );
      await act(async () => {
        result.current.processFile(makeFile("key.pem", "key content"));
      });
      // onChange still called, no error thrown
      expect(onChange).toHaveBeenCalledWith("key content");
      restore();
    });

    it("ignores files larger than 512 KB", async () => {
      const onChange = vi.fn();
      const { result } = renderHook(() =>
        useKeyFileInput({ validate: () => true, onChange }),
      );
      await act(async () => {
        result.current.processFile(makeOversizedFile());
      });
      expect(onChange).not.toHaveBeenCalled();
    });

    it("strips only the last extension from the filename", async () => {
      const restore = mockFileReader("data");
      const onFileName = vi.fn();
      const { result } = renderHook(() =>
        useKeyFileInput({ validate: () => true, onChange: vi.fn(), onFileName }),
      );
      await act(async () => {
        result.current.processFile(makeFile("my.key.pem", "data"));
      });
      expect(onFileName).toHaveBeenCalledWith("my.key");
      restore();
    });
  });

  describe("handleDrop", () => {
    it("calls preventDefault on the event", () => {
      const restore = mockFileReader("key");
      const { result } = renderHook(() =>
        useKeyFileInput({ validate: () => true, onChange: vi.fn() }),
      );
      const event = makeDropEvent(makeFile("k.pub", "key"));
      act(() => result.current.handleDrop(event));
      expect((event.preventDefault as ReturnType<typeof vi.fn>)).toHaveBeenCalled();
      restore();
    });

    it("processes the first dropped file", async () => {
      const restore = mockFileReader("dropped content");
      const onChange = vi.fn();
      const { result } = renderHook(() =>
        useKeyFileInput({ validate: () => true, onChange }),
      );
      await act(async () => {
        result.current.handleDrop(makeDropEvent(makeFile("k.pub", "dropped content")));
      });
      expect(onChange).toHaveBeenCalledWith("dropped content");
      restore();
    });

    it("does nothing when dataTransfer has no files", () => {
      const onChange = vi.fn();
      const { result } = renderHook(() =>
        useKeyFileInput({ validate: () => true, onChange }),
      );
      const emptyDropEvent = {
        preventDefault: vi.fn(),
        dataTransfer: { files: [] },
      } as unknown as React.DragEvent<Element>;
      act(() => result.current.handleDrop(emptyDropEvent));
      expect(onChange).not.toHaveBeenCalled();
    });
  });

  describe("handleFileInputChange", () => {
    it("processes the selected file", async () => {
      const restore = mockFileReader("input content");
      const onChange = vi.fn();
      const { result } = renderHook(() =>
        useKeyFileInput({ validate: () => true, onChange }),
      );
      const file = makeFile("k.pem", "input content");
      const event = {
        target: { files: [file], value: "" },
      } as unknown as React.ChangeEvent<HTMLInputElement>;
      await act(async () => {
        result.current.handleFileInputChange(event);
      });
      expect(onChange).toHaveBeenCalledWith("input content");
      restore();
    });

    it("clears the input value after processing", async () => {
      const restore = mockFileReader("x");
      const { result } = renderHook(() =>
        useKeyFileInput({ validate: () => true, onChange: vi.fn() }),
      );
      const target = { files: [makeFile("k.pem", "x")], value: "some-path" };
      const event = { target } as unknown as React.ChangeEvent<HTMLInputElement>;
      await act(async () => {
        result.current.handleFileInputChange(event);
      });
      expect(target.value).toBe("");
      restore();
    });

    it("does nothing when no file is selected", () => {
      const onChange = vi.fn();
      const { result } = renderHook(() =>
        useKeyFileInput({ validate: () => true, onChange }),
      );
      const event = {
        target: { files: [], value: "" },
      } as unknown as React.ChangeEvent<HTMLInputElement>;
      act(() => result.current.handleFileInputChange(event));
      expect(onChange).not.toHaveBeenCalled();
    });
  });

  describe("paste interception", () => {
    function firePaste(text: string) {
      const event = new Event("paste") as ClipboardEvent;
      Object.defineProperty(event, "clipboardData", {
        value: { getData: (_: string) => text },
      });
      // Make preventDefault a spy we can assert on
      event.preventDefault = vi.fn();
      document.dispatchEvent(event);
      return event;
    }

    it("calls onChange when pasted text passes validation", () => {
      const onChange = vi.fn();
      renderHook(() =>
        useKeyFileInput({ validate: (t) => t.startsWith("ssh-rsa"), onChange }),
      );
      const event = firePaste("ssh-rsa AAAAB3NzaC1");
      expect(onChange).toHaveBeenCalledWith("ssh-rsa AAAAB3NzaC1");
      expect(event.preventDefault).toHaveBeenCalled();
    });

    it("does NOT call onChange when pasted text fails validation", () => {
      const onChange = vi.fn();
      renderHook(() =>
        useKeyFileInput({ validate: (t) => t.startsWith("ssh-rsa"), onChange }),
      );
      const event = firePaste("not-a-key");
      expect(onChange).not.toHaveBeenCalled();
      expect(event.preventDefault).not.toHaveBeenCalled();
    });

    it("validates trimmed text, not raw paste text", () => {
      const validate = vi.fn().mockReturnValue(true);
      const onChange = vi.fn();
      renderHook(() => useKeyFileInput({ validate, onChange }));
      firePaste("  ssh-rsa AAAAB  ");
      expect(validate).toHaveBeenCalledWith("ssh-rsa AAAAB");
    });

    it("does not register paste listener when disabled", () => {
      const onChange = vi.fn();
      renderHook(() =>
        useKeyFileInput({
          validate: () => true,
          onChange,
          disabled: true,
        }),
      );
      firePaste("any text");
      expect(onChange).not.toHaveBeenCalled();
    });

    it("removes paste listener on unmount", () => {
      const onChange = vi.fn();
      const { unmount } = renderHook(() =>
        useKeyFileInput({ validate: () => true, onChange }),
      );
      unmount();
      firePaste("after unmount");
      expect(onChange).not.toHaveBeenCalled();
    });
  });

  describe("ref stability", () => {
    it("always calls the latest onChange without needing new processFile reference", async () => {
      const firstOnChange = vi.fn();
      const secondOnChange = vi.fn();
      const restore = mockFileReader("content");

      const { result, rerender } = renderHook(
        ({ onChange }: { onChange: (t: string) => void }) =>
          useKeyFileInput({ validate: () => true, onChange }),
        { initialProps: { onChange: firstOnChange } },
      );

      rerender({ onChange: secondOnChange });

      await act(async () => {
        result.current.processFile(makeFile("k.pem", "content"));
      });

      expect(firstOnChange).not.toHaveBeenCalled();
      expect(secondOnChange).toHaveBeenCalledWith("content");
      restore();
    });
  });
});
