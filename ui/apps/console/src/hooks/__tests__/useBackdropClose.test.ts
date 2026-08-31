import { describe, it, expect, vi } from "vitest";
import { renderHook } from "@testing-library/react";
import { useRef } from "react";
import { useBackdropClose } from "../useBackdropClose";
function makeDialog(): HTMLDialogElement {
  return document.createElement("dialog");
}

function makeChild(dialog: HTMLDialogElement): HTMLDivElement {
  const child = document.createElement("div");
  dialog.appendChild(child);
  return child;
}

function setup(
  dialog: HTMLDialogElement,
  options: { enabled?: () => boolean } = {},
) {
  const onClose = vi.fn();

  const { result } = renderHook(() => {
    const ref = useRef<HTMLDialogElement>(dialog);
    return useBackdropClose(ref, onClose, options.enabled);
  });

  return { handlers: result.current, onClose };
}

function fireSequence(
  handlers: {
    onMouseDown: React.MouseEventHandler<HTMLDialogElement>;
    onClick: React.MouseEventHandler<HTMLDialogElement>;
  },
  _dialog: HTMLDialogElement,
  mouseDownTarget: EventTarget,
  clickTarget: EventTarget,
) {
  handlers.onMouseDown({
    target: mouseDownTarget,
  } as unknown as React.MouseEvent<HTMLDialogElement>);

  handlers.onClick({
    target: clickTarget,
  } as unknown as React.MouseEvent<HTMLDialogElement>);
}

describe("useBackdropClose", () => {
  describe("backdrop click — closes dialog", () => {
    it("calls onClose when both mousedown and click land on the dialog element", () => {
      const dialog = makeDialog();
      const { handlers, onClose } = setup(dialog);

      fireSequence(handlers, dialog, dialog, dialog);

      expect(onClose).toHaveBeenCalledOnce();
    });
  });

  describe("content click — does NOT close dialog", () => {
    it("does not call onClose when click lands on a child element", () => {
      const dialog = makeDialog();
      const child = makeChild(dialog);
      const { handlers, onClose } = setup(dialog);

      fireSequence(handlers, dialog, dialog, child);

      expect(onClose).not.toHaveBeenCalled();
    });

    it("does not call onClose when mousedown starts on a child but click ends on dialog", () => {
      const dialog = makeDialog();
      const child = makeChild(dialog);
      const { handlers, onClose } = setup(dialog);

      fireSequence(handlers, dialog, child, dialog);

      expect(onClose).not.toHaveBeenCalled();
    });

    it("does not call onClose when both mousedown and click land on a child element", () => {
      const dialog = makeDialog();
      const child = makeChild(dialog);
      const { handlers, onClose } = setup(dialog);

      fireSequence(handlers, dialog, child, child);

      expect(onClose).not.toHaveBeenCalled();
    });
  });

  describe("enabled guard", () => {
    it("does not call onClose when enabled() returns false", () => {
      const dialog = makeDialog();
      const { handlers, onClose } = setup(dialog, { enabled: () => false });

      fireSequence(handlers, dialog, dialog, dialog);

      expect(onClose).not.toHaveBeenCalled();
    });

    it("calls onClose when enabled() returns true", () => {
      const dialog = makeDialog();
      const { handlers, onClose } = setup(dialog, { enabled: () => true });

      fireSequence(handlers, dialog, dialog, dialog);

      expect(onClose).toHaveBeenCalledOnce();
    });

    it("uses the default enabled guard (always true) when no guard is provided", () => {
      const dialog = makeDialog();
      const { handlers, onClose } = setup(dialog);

      fireSequence(handlers, dialog, dialog, dialog);

      expect(onClose).toHaveBeenCalledOnce();
    });
  });

  describe("mousedown tracking is reset between clicks", () => {
    it("does not close on a second click that has no preceding mousedown on the dialog", () => {
      const dialog = makeDialog();
      const child = makeChild(dialog);
      const { handlers, onClose } = setup(dialog);

      fireSequence(handlers, dialog, child, dialog);
      expect(onClose).not.toHaveBeenCalled();

      handlers.onClick({
        target: dialog,
      } as unknown as React.MouseEvent<HTMLDialogElement>);

      expect(onClose).not.toHaveBeenCalled();
    });

    it("closes correctly on a subsequent full backdrop click after a missed one", () => {
      const dialog = makeDialog();
      const child = makeChild(dialog);
      const { handlers, onClose } = setup(dialog);

      fireSequence(handlers, dialog, child, dialog);
      expect(onClose).not.toHaveBeenCalled();

      fireSequence(handlers, dialog, dialog, dialog);
      expect(onClose).toHaveBeenCalledOnce();
    });
  });
});
