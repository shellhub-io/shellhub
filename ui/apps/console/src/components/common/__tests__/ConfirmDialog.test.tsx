import { describe, it, expect, vi } from "vitest";
import {
  render,
  screen,
  fireEvent,
  waitFor,
  act,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import ConfirmDialog from "../ConfirmDialog";
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

    it("renders the title, the description and any children", () => {
      renderDialog(true, {
        title: "Confirm deletion",
        description: "This cannot be undone.",
        children: <span data-testid="extra-content">extra</span>,
      });
      expect(
        screen.getByRole("heading", { name: "Confirm deletion" }),
      ).toBeInTheDocument();
      expect(screen.getByText("This cannot be undone.")).toBeInTheDocument();
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
  });

  describe("buttons", () => {
    it("renders custom confirm and cancel labels", () => {
      renderDialog(true, { confirmLabel: "Delete", cancelLabel: "Go back" });
      expect(
        screen.getByRole("button", { name: "Delete" }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: "Go back" }),
      ).toBeInTheDocument();
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

    it("disables the confirm button while onConfirm is pending", async () => {
      const user = userEvent.setup();
      let resolve!: () => void;
      const onConfirm = vi.fn(
        () =>
          new Promise<void>((res) => {
            resolve = res;
          }),
      );
      renderDialog(true, { onConfirm });

      await user.click(screen.getByRole("button", { name: "Confirm" }));

      expect(screen.getByRole("button", { name: "Confirm" })).toBeDisabled();

      act(() => {
        resolve();
      });
    });

    it("re-enables the confirm button after onConfirm resolves", async () => {
      const user = userEvent.setup();
      const onConfirm = vi.fn().mockResolvedValue(undefined);
      renderDialog(true, { onConfirm });

      await user.click(screen.getByRole("button", { name: "Confirm" }));

      await waitFor(() =>
        expect(
          screen.getByRole("button", { name: "Confirm" }),
        ).not.toBeDisabled(),
      );
    });

    it("re-enables the confirm button after onConfirm rejects", async () => {
      const user = userEvent.setup();
      const onConfirm = vi.fn().mockRejectedValue(new Error("fail"));
      renderDialog(true, { onConfirm });

      await user.click(screen.getByRole("button", { name: "Confirm" }));

      await waitFor(() =>
        expect(
          screen.getByRole("button", { name: "Confirm" }),
        ).not.toBeDisabled(),
      );
    });

    it("disables the confirm button when confirmDisabled=true", () => {
      renderDialog(true, { confirmDisabled: true });
      expect(screen.getByRole("button", { name: "Confirm" })).toBeDisabled();
    });
  });
});
