import { describe, it, expect, vi } from "vitest";
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

      fireEvent(screen.getByRole("dialog"), new Event("cancel"));

      expect(onClose).toHaveBeenCalledOnce();
    });

    it("does NOT call onClose when canClose returns false", () => {
      const { onClose } = renderDialog(true, {
        canClose: () => false,
      });

      fireEvent(screen.getByRole("dialog"), new Event("cancel"));

      expect(onClose).not.toHaveBeenCalled();
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
});
