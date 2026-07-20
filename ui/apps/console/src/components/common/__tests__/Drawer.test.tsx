import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import Drawer from "@/components/common/Drawer";

function renderDrawer(props: { open?: boolean; onClose?: () => void } = {}) {
  const onClose = props.onClose ?? vi.fn();
  return {
    onClose,
    ...render(
      <Drawer open={props.open ?? true} onClose={onClose} title="Edit device">
        <p>Body</p>
      </Drawer>,
    ),
  };
}

describe("Drawer accessibility", () => {
  it("exposes the panel as a modal dialog labelled by its title", () => {
    renderDrawer();

    const dialog = screen.getByRole("dialog", { name: "Edit device" });
    expect(dialog).toHaveAttribute("aria-modal", "true");
  });

  it("associates aria-labelledby with the heading element via a matching id", () => {
    renderDrawer();

    const dialog = screen.getByRole("dialog");
    const heading = screen.getByRole("heading", { name: "Edit device" });

    expect(dialog.getAttribute("aria-labelledby")).toBe(heading.id);
    expect(heading.id).not.toBe("");
  });
});

describe("Drawer Escape key", () => {
  it("closes when Escape is pressed and the panel has focus", async () => {
    const user = userEvent.setup();
    const { onClose } = renderDrawer();

    screen.getByLabelText("Close").focus();
    await user.keyboard("{Escape}");

    expect(onClose).toHaveBeenCalledOnce();
  });

  it("does not close when focus is inside a sibling dialog", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();

    render(
      <>
        <Drawer open onClose={onClose} title="Edit device">
          <p>Body</p>
        </Drawer>
        <dialog ref={(el) => el?.showModal()}>
          <button type="button">Inside dialog</button>
        </dialog>
      </>,
    );

    const inner = screen.getByText("Inside dialog");
    inner.focus();

    await user.keyboard("{Escape}");

    expect(onClose).not.toHaveBeenCalled();
  });
});
