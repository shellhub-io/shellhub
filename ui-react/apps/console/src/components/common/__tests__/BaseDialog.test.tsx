import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup, fireEvent } from "@testing-library/react";
import { createRef } from "react";
import "./helpers/setup-dialog";

// Mock useFocusTrap to avoid jsdom focus-management side effects.
vi.mock("@/hooks/useFocusTrap", () => ({
  useFocusTrap: vi.fn(),
}));

import BaseDialog from "../BaseDialog";

afterEach(cleanup);

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

    it("renders a <dialog> element when open=true", () => {
      renderDialog(true);
      expect(screen.getByRole("dialog")).toBeInTheDocument();
    });

    it("renders children inside the dialog", () => {
      renderDialog(true);
      expect(screen.getByText("dialog content")).toBeInTheDocument();
    });
  });

  describe("showModal / close lifecycle", () => {
    it("calls showModal() when open=true", () => {
      renderDialog(true);
      expect(HTMLDialogElement.prototype.showModal).toHaveBeenCalled();
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

    it("calls onClose when canClose returns true", () => {
      const { onClose } = renderDialog(true, {
        canClose: () => true,
      });

      fireEvent(screen.getByRole("dialog"), new Event("cancel"));

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
    it("passes aria-labelledby to the dialog element", () => {
      renderDialog(true, { ariaLabelledBy: "my-title" });
      expect(screen.getByRole("dialog")).toHaveAttribute(
        "aria-labelledby",
        "my-title",
      );
    });

    it("passes aria-describedby to the dialog element", () => {
      renderDialog(true, { ariaDescribedBy: "my-description" });
      expect(screen.getByRole("dialog")).toHaveAttribute(
        "aria-describedby",
        "my-description",
      );
    });

    it("passes aria-label to the dialog element", () => {
      renderDialog(true, { ariaLabel: "My accessible dialog" });
      expect(screen.getByRole("dialog")).toHaveAttribute(
        "aria-label",
        "My accessible dialog",
      );
    });
  });

  describe("data-custom-backdrop", () => {
    it("has data-custom-backdrop attribute for native ::backdrop CSS styling", () => {
      renderDialog(true);
      expect(screen.getByRole("dialog")).toHaveAttribute("data-custom-backdrop");
    });
  });

  describe("size prop", () => {
    it("applies sm:max-w-sm by default (size omitted)", () => {
      renderDialog(true);
      expect(screen.getByRole("dialog").className).toContain("sm:max-w-sm");
    });

    it("applies sm:max-w-sm when size='sm'", () => {
      renderDialog(true, { size: "sm" });
      expect(screen.getByRole("dialog").className).toContain("sm:max-w-sm");
    });

    it("applies sm:max-w-md when size='md'", () => {
      renderDialog(true, { size: "md" });
      expect(screen.getByRole("dialog").className).toContain("sm:max-w-md");
    });

    it("applies sm:max-w-lg when size='lg'", () => {
      renderDialog(true, { size: "lg" });
      expect(screen.getByRole("dialog").className).toContain("sm:max-w-lg");
    });

    it("applies sm:max-w-xl when size='xl'", () => {
      renderDialog(true, { size: "xl" });
      expect(screen.getByRole("dialog").className).toContain("sm:max-w-xl");
    });

    it("does not apply any max-w class when size='full'", () => {
      renderDialog(true, { size: "full" });
      expect(screen.getByRole("dialog").className).not.toContain("sm:max-w");
    });
  });

  describe("className prop", () => {
    it("appends extra classes to the dialog panel", () => {
      renderDialog(true, { className: "sm:max-h-[85vh]" });
      expect(screen.getByRole("dialog").className).toContain("sm:max-h-[85vh]");
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
