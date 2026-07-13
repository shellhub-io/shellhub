import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import Drawer from "@/components/common/Drawer";

function renderDrawer(open = true) {
  return render(
    <Drawer open={open} onClose={vi.fn()} title="Edit device">
      <p>Body</p>
    </Drawer>,
  );
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
