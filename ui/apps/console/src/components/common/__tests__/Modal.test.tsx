import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
vi.unmock("@/hooks/useFocusTrap");
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import Modal from "@/components/common/Modal";

function renderModal(props: { open?: boolean; onClose?: () => void } = {}) {
  const onClose = props.onClose ?? vi.fn();
  return {
    onClose,
    ...render(
      <Modal open={props.open ?? true} onClose={onClose} title="Edit device">
        <p>Body</p>
      </Modal>,
    ),
  };
}

describe("Modal", () => {
  it("is a dialog named by its title", () => {
    renderModal();

    expect(
      screen.getByRole("dialog", { name: "Edit device" }),
    ).toBeInTheDocument();
  });

  it("closes from its close button", async () => {
    const user = userEvent.setup();
    const { onClose } = renderModal();

    await user.click(screen.getByRole("button", { name: "Close" }));

    expect(onClose).toHaveBeenCalledOnce();
  });

  it("renders nothing while closed", () => {
    renderModal({ open: false });

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  describe("focus on open", () => {
    beforeEach(() => {
      vi.spyOn(window, "requestAnimationFrame").mockImplementation((cb) => {
        cb(0);
        return 0;
      });
    });

    afterEach(() => {
      vi.restoreAllMocks();
    });

    it("lands on the first field, not the close button", () => {
      render(
        <Modal open onClose={vi.fn()} title="Rename">
          <input aria-label="Name" />
        </Modal>,
      );

      expect(screen.getByLabelText("Name")).toHaveFocus();
    });

    it("lands on the close button when there is nothing else to focus", () => {
      render(
        <Modal open onClose={vi.fn()} title="Details">
          <p>Read only</p>
        </Modal>,
      );

      expect(screen.getByRole("button", { name: "Close" })).toHaveFocus();
    });
  });
});
