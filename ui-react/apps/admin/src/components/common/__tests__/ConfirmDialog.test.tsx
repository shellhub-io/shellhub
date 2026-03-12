import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup, fireEvent, waitFor, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "./helpers/setup-dialog";

// Mock useFocusTrap to avoid jsdom focus-management side effects.
vi.mock("@/hooks/useFocusTrap", () => ({
  useFocusTrap: vi.fn(),
}));

import ConfirmDialog from "../ConfirmDialog";

afterEach(cleanup);

function renderDialog(
  open: boolean,
  {
    onClose = vi.fn(),
    onConfirm = vi.fn(),
    title = "Delete item",
    description = "Are you sure?",
    confirmLabel,
    cancelLabel,
    variant,
    confirmDisabled,
    children,
  }: {
    onClose?: () => void;
    onConfirm?: () => Promise<void> | void;
    title?: string;
    description?: React.ReactNode;
    confirmLabel?: string;
    cancelLabel?: string;
    variant?: "primary" | "danger" | "success" | "warning";
    confirmDisabled?: boolean;
    children?: React.ReactNode;
  } = {},
) {
  return {
    onClose,
    onConfirm,
    ...render(
      <ConfirmDialog
        open={open}
        onClose={onClose}
        onConfirm={onConfirm}
        title={title}
        description={description}
        confirmLabel={confirmLabel}
        cancelLabel={cancelLabel}
        variant={variant}
        confirmDisabled={confirmDisabled}
      >
        {children}
      </ConfirmDialog>,
    ),
  };
}

describe("ConfirmDialog", () => {
  describe("rendering", () => {
    it("renders nothing when open=false", () => {
      renderDialog(false);
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });

    it("renders the dialog when open=true", () => {
      renderDialog(true);
      expect(screen.getByRole("dialog")).toBeInTheDocument();
    });

    it("renders the title", () => {
      renderDialog(true, { title: "Confirm deletion" });
      expect(
        screen.getByRole("heading", { name: "Confirm deletion" }),
      ).toBeInTheDocument();
    });

    it("renders the description", () => {
      renderDialog(true, { description: "This cannot be undone." });
      expect(screen.getByText("This cannot be undone.")).toBeInTheDocument();
    });

    it("renders children between description and buttons", () => {
      renderDialog(true, {
        children: <span data-testid="extra-content">extra</span>,
      });
      expect(screen.getByTestId("extra-content")).toBeInTheDocument();
    });
  });

  describe("ARIA", () => {
    it("dialog has aria-labelledby wired to the title element", () => {
      renderDialog(true, { title: "My title" });
      const dialog = screen.getByRole("dialog");
      const labelId = dialog.getAttribute("aria-labelledby");
      expect(labelId).toBeTruthy();
      const titleEl = document.getElementById(labelId!);
      expect(titleEl).toHaveTextContent("My title");
    });

    it("title element has an id that matches dialog's aria-labelledby", () => {
      renderDialog(true);
      const dialog = screen.getByRole("dialog");
      const labelId = dialog.getAttribute("aria-labelledby")!;
      expect(screen.getByRole("heading")).toHaveAttribute("id", labelId);
    });
  });

  describe("buttons", () => {
    it("renders the confirm button with default label 'Confirm'", () => {
      renderDialog(true);
      expect(
        screen.getByRole("button", { name: "Confirm" }),
      ).toBeInTheDocument();
    });

    it("renders the cancel button with default label 'Cancel'", () => {
      renderDialog(true);
      expect(
        screen.getByRole("button", { name: "Cancel" }),
      ).toBeInTheDocument();
    });

    it("renders a custom confirmLabel", () => {
      renderDialog(true, { confirmLabel: "Delete" });
      expect(screen.getByRole("button", { name: "Delete" })).toBeInTheDocument();
    });

    it("renders a custom cancelLabel", () => {
      renderDialog(true, { cancelLabel: "Go back" });
      expect(screen.getByRole("button", { name: "Go back" })).toBeInTheDocument();
    });
  });

  describe("cancel", () => {
    it("calls onClose when Cancel button is clicked", async () => {
      const user = userEvent.setup();
      const { onClose } = renderDialog(true);

      await user.click(screen.getByRole("button", { name: "Cancel" }));

      expect(onClose).toHaveBeenCalledOnce();
    });

    it("calls onClose when the native cancel event fires (ESC key)", () => {
      const { onClose } = renderDialog(true);

      fireEvent(screen.getByRole("dialog"), new Event("cancel"));

      expect(onClose).toHaveBeenCalledOnce();
    });
  });

  describe("confirm", () => {
    it("calls onConfirm when the confirm button is clicked", async () => {
      const user = userEvent.setup();
      const { onConfirm } = renderDialog(true);

      await user.click(screen.getByRole("button", { name: "Confirm" }));

      expect(onConfirm).toHaveBeenCalledOnce();
    });

    it("shows a spinner and disables the confirm button while onConfirm is pending", async () => {
      const user = userEvent.setup();
      let resolve!: () => void;
      const onConfirm = vi.fn(
        () => new Promise<void>((res) => { resolve = res; }),
      );
      renderDialog(true, { onConfirm });

      await user.click(screen.getByRole("button", { name: "Confirm" }));

      const confirmBtn = screen.getByRole("button", { name: "Confirm" });
      expect(confirmBtn).toBeDisabled();
      expect(screen.getByTestId("confirm-spinner")).toBeInTheDocument();

      act(() => { resolve(); });
    });

    it("re-enables the confirm button after onConfirm resolves", async () => {
      const user = userEvent.setup();
      const onConfirm = vi.fn().mockResolvedValue(undefined);
      renderDialog(true, { onConfirm });

      await user.click(screen.getByRole("button", { name: "Confirm" }));

      await waitFor(() =>
        expect(screen.getByRole("button", { name: "Confirm" })).not.toBeDisabled(),
      );
    });

    it("re-enables the confirm button after onConfirm rejects", async () => {
      const user = userEvent.setup();
      const onConfirm = vi.fn().mockRejectedValue(new Error("fail"));
      renderDialog(true, { onConfirm });

      await user.click(screen.getByRole("button", { name: "Confirm" }));

      await waitFor(() =>
        expect(screen.getByRole("button", { name: "Confirm" })).not.toBeDisabled(),
      );
    });

    it("disables the confirm button when confirmDisabled=true", () => {
      renderDialog(true, { confirmDisabled: true });
      expect(screen.getByRole("button", { name: "Confirm" })).toBeDisabled();
    });
  });

  describe("variant", () => {
    it("applies danger variant classes by default", () => {
      renderDialog(true);
      const btn = screen.getByRole("button", { name: "Confirm" });
      expect(btn.className).toContain("bg-accent-red/90");
    });

    it("applies primary variant classes when variant='primary'", () => {
      renderDialog(true, { variant: "primary" });
      const btn = screen.getByRole("button", { name: "Confirm" });
      expect(btn.className).toContain("bg-primary");
      expect(btn.className).not.toContain("bg-accent-red");
    });

    it("applies success variant classes when variant='success'", () => {
      renderDialog(true, { variant: "success" });
      const btn = screen.getByRole("button", { name: "Confirm" });
      expect(btn.className).toContain("bg-accent-green");
      expect(btn.className).not.toContain("bg-accent-red");
    });

    it("applies warning variant classes when variant='warning'", () => {
      renderDialog(true, { variant: "warning" });
      const btn = screen.getByRole("button", { name: "Confirm" });
      expect(btn.className).toContain("bg-accent-yellow");
      expect(btn.className).not.toContain("bg-accent-red");
    });
  });
});
