import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import ConfirmDialog from "../ConfirmDialog";
function renderDialog(
  open: boolean,
  {
    onClose = vi.fn(),
    onConfirm = vi.fn(),
    title = "Delete item",
    description = "Are you sure?",
    variant,
    confirmDisabled,
  }: {
    onClose?: () => void;
    onConfirm?: () => Promise<void> | void;
    title?: string;
    description?: React.ReactNode;
    variant?: "primary" | "danger" | "success" | "warning";
    confirmDisabled?: boolean;
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
        icon={<svg />}
        title={title}
        description={description}
        variant={variant}
        confirmDisabled={confirmDisabled}
      />,
    ),
  };
}

describe("ConfirmDialog", () => {
  describe("rendering", () => {
    it("renders nothing when open=false", () => {
      renderDialog(false);
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });

  describe("ARIA", () => {
    it("is described by its description", () => {
      renderDialog(true, { description: "The key stops working." });
      expect(screen.getByRole("dialog")).toHaveAccessibleDescription(
        "The key stops working.",
      );
    });

    it("dialog has aria-labelledby wired to the title element", () => {
      renderDialog(true, { title: "My title" });
      const dialog = screen.getByRole("dialog");
      const labelId = dialog.getAttribute("aria-labelledby");
      expect(labelId).toBeTruthy();
      const titleEl = document.getElementById(labelId!);
      expect(titleEl).toHaveTextContent("My title");
    });
  });

  describe("cancel", () => {
    it("calls onClose when Cancel button is clicked", async () => {
      const user = userEvent.setup();
      const { onClose } = renderDialog(true);

      await user.click(screen.getByRole("button", { name: "Cancel" }));

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
