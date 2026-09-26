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
      <Modal
        open={props.open ?? true}
        onClose={onClose}
        icon={<svg />}
        title="Edit device"
        description="Change how the device is named."
        footerStart={<a href="/docs">Naming rules</a>}
      >
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

  it("is described by its description", () => {
    renderModal();

    expect(screen.getByRole("dialog")).toHaveAccessibleDescription(
      "Change how the device is named.",
    );
  });

  it("renders the footer's leading content", () => {
    renderModal();

    expect(
      screen.getByRole("link", { name: "Naming rules" }),
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

  it("keeps its name and description when laid out centred", () => {
    render(
      <Modal
        open
        onClose={vi.fn()}
        layout="center"
        icon={<svg />}
        title="MFA is on"
        description="Keep these recovery codes."
        footer={<button type="button">I saved them</button>}
      >
        <p>codes</p>
      </Modal>,
    );

    const dialog = screen.getByRole("dialog", { name: "MFA is on" });
    expect(dialog).toHaveAccessibleDescription("Keep these recovery codes.");
    expect(
      screen.getByRole("button", { name: "I saved them" }),
    ).toBeInTheDocument();
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
        <Modal
          open
          onClose={vi.fn()}
          icon={<svg />}
          title="Rename"
          description="Pick a new name."
        >
          <input aria-label="Name" />
        </Modal>,
      );

      expect(screen.getByLabelText("Name")).toHaveFocus();
    });

    it("lands on the close button when there is nothing else to focus", () => {
      render(
        <Modal
          open
          onClose={vi.fn()}
          icon={<svg />}
          title="Details"
          description="What the device reports."
        >
          <p>Read only</p>
        </Modal>,
      );

      expect(screen.getByRole("button", { name: "Close" })).toHaveFocus();
    });
  });
});
