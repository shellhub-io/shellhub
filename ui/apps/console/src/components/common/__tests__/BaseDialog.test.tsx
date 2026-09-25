import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
vi.unmock("@/hooks/useFocusTrap");
import { render, screen, fireEvent } from "@testing-library/react";
import { createRef } from "react";
import BaseDialog from "../BaseDialog";
function renderDialog(
  open: boolean,
  {
    onClose = vi.fn(),
    canClose,
    size,
    ariaLabelledBy,
    ariaDescribedBy,
    ariaLabel,
    className,
  }: {
    onClose?: () => void;
    canClose?: () => boolean;
    size?: "sm" | "md" | "lg" | "xl" | "full";
    ariaLabelledBy?: string;
    ariaDescribedBy?: string;
    ariaLabel?: string;
    className?: string;
  } = {},
) {
  return {
    onClose,
    ...render(
      <BaseDialog
        open={open}
        onClose={onClose}
        canClose={canClose}
        size={size}
        aria-labelledby={ariaLabelledBy}
        aria-describedby={ariaDescribedBy}
        aria-label={ariaLabel}
        className={className}
      >
        <p>dialog content</p>
      </BaseDialog>,
    ),
  };
}

describe("BaseDialog", () => {
  describe("rendering", () => {
    it("renders nothing when open=false", () => {
      renderDialog(false);
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });

    it("renders children inside the dialog", () => {
      renderDialog(true);
      expect(screen.getByText("dialog content")).toBeInTheDocument();
    });
  });

  describe("showModal / close lifecycle", () => {
    it("calls showModal() when open=true", () => {
      const showModalSpy = vi.spyOn(HTMLDialogElement.prototype, "showModal");
      renderDialog(true);
      expect(showModalSpy).toHaveBeenCalled();
    });

    it("removes the dialog from the DOM when open transitions to false", () => {
      const { rerender } = render(
        <BaseDialog open={true} onClose={vi.fn()}>
          <p>content</p>
        </BaseDialog>,
      );

      expect(screen.getByRole("dialog")).toBeInTheDocument();

      rerender(
        <BaseDialog open={false} onClose={vi.fn()}>
          <p>content</p>
        </BaseDialog>,
      );

      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });

  describe("ESC / cancel event", () => {
    it("calls onClose when the native cancel event fires", () => {
      const { onClose } = renderDialog(true);

      fireEvent(
        screen.getByRole("dialog"),
        new Event("cancel", { cancelable: true }),
      );

      expect(onClose).toHaveBeenCalledOnce();
    });

    it("calls onClose when a dialog mounted closed is opened later", () => {
      const onClose = vi.fn();
      const { rerender } = render(
        <BaseDialog open={false} onClose={onClose}>
          <p>content</p>
        </BaseDialog>,
      );
      rerender(
        <BaseDialog open={true} onClose={onClose}>
          <p>content</p>
        </BaseDialog>,
      );

      fireEvent(
        screen.getByRole("dialog"),
        new Event("cancel", { cancelable: true }),
      );

      expect(onClose).toHaveBeenCalledOnce();
    });

    it("does NOT call onClose when canClose returns false", () => {
      const { onClose } = renderDialog(true, {
        canClose: () => false,
      });

      fireEvent(
        screen.getByRole("dialog"),
        new Event("cancel", { cancelable: true }),
      );

      expect(onClose).not.toHaveBeenCalled();
    });
  });

  describe("closed by the browser", () => {
    function closeNatively(dialog: HTMLDialogElement) {
      fireEvent(dialog, new Event("cancel"));
      dialog.close();
      fireEvent(dialog, new Event("close"));
    }

    it("reopens and stays open when canClose refuses", () => {
      const { onClose } = renderDialog(true, { canClose: () => false });
      const dialog = screen.getByRole<HTMLDialogElement>("dialog");

      closeNatively(dialog);

      expect(dialog).toHaveAttribute("open");
      expect(onClose).not.toHaveBeenCalled();
    });

    it("reopens before asking its owner to close, so a prompt the owner opens lands on top", () => {
      const onClose = vi.fn(() => {
        expect(screen.getByRole("dialog")).toHaveAttribute("open");
      });
      renderDialog(true, { onClose });

      closeNatively(screen.getByRole<HTMLDialogElement>("dialog"));

      expect(onClose).toHaveBeenCalledOnce();
    });
  });

  describe("backdrop click", () => {
    it("calls onClose when backdrop is clicked", () => {
      const { onClose } = renderDialog(true);

      const dialog = document.querySelector("dialog") as HTMLElement;
      fireEvent.mouseDown(dialog);
      fireEvent.click(dialog);

      expect(onClose).toHaveBeenCalledOnce();
    });

    it("does NOT call onClose when backdrop is clicked with canClose=false", () => {
      const { onClose } = renderDialog(true, { canClose: () => false });

      const dialog = document.querySelector("dialog") as HTMLElement;
      fireEvent.mouseDown(dialog);
      fireEvent.click(dialog);

      expect(onClose).not.toHaveBeenCalled();
    });
  });

  describe("ARIA attributes", () => {
    it("forwards the aria-* props to the dialog element", () => {
      renderDialog(true, {
        ariaLabelledBy: "my-title",
        ariaDescribedBy: "my-description",
        ariaLabel: "My accessible dialog",
      });

      const dialog = screen.getByRole("dialog");
      expect(dialog).toHaveAttribute("aria-labelledby", "my-title");
      expect(dialog).toHaveAttribute("aria-describedby", "my-description");
      expect(dialog).toHaveAttribute("aria-label", "My accessible dialog");
    });
  });

  describe("dialogRef prop", () => {
    it("forwards dialogRef to the underlying <dialog> element", () => {
      const ref = createRef<HTMLDialogElement>();
      render(
        <BaseDialog open={true} onClose={vi.fn()} dialogRef={ref}>
          <p>content</p>
        </BaseDialog>,
      );
      expect(ref.current).toBeInstanceOf(HTMLDialogElement);
    });
  });
  describe("focus on open", () => {
    const narrow = (matches: boolean) => vi.fn().mockReturnValue({ matches });

    beforeEach(() => {
      vi.spyOn(window, "requestAnimationFrame").mockImplementation((cb) => {
        cb(0);
        return 0;
      });
    });

    afterEach(() => {
      vi.restoreAllMocks();
      Reflect.deleteProperty(window, "matchMedia");
      Reflect.deleteProperty(window, "__TAURI__");
    });

    function renderForm() {
      render(
        <BaseDialog open={true} onClose={vi.fn()} aria-label="Form">
          <input aria-label="Name" />
        </BaseDialog>,
      );
    }

    it("moves focus to the first field on a wide screen", () => {
      window.matchMedia = narrow(false);
      renderForm();
      expect(screen.getByLabelText("Name")).toHaveFocus();
    });

    it("focuses the dialog itself on a phone, so the keyboard does not cover it", () => {
      window.matchMedia = narrow(true);
      renderForm();
      expect(screen.getByRole("dialog")).toHaveFocus();
    });

    it("keeps focusing the first field in a narrow desktop app window", () => {
      window.matchMedia = narrow(true);
      window.__TAURI__ = {
        window: {
          getCurrentWindow: () => ({
            close: vi.fn(),
            minimize: vi.fn(),
            toggleMaximize: vi.fn(),
          }),
        },
      };
      renderForm();
      expect(screen.getByLabelText("Name")).toHaveFocus();
    });
  });
});
