import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import DialogHeader from "../DialogHeader";

function renderInDialog(onClose?: () => void) {
  render(
    <div
      role="dialog"
      aria-labelledby="header-title"
      aria-describedby="header-description"
    >
      <DialogHeader
        icon={<svg />}
        title="Delete key"
        description="Integrations using it stop working."
        titleId="header-title"
        descriptionId="header-description"
        onClose={onClose}
      />
    </div>,
  );
}

describe("DialogHeader", () => {
  it("names and describes the dialog through the ids it is given", () => {
    renderInDialog();

    const dialog = screen.getByRole("dialog", { name: "Delete key" });
    expect(dialog).toHaveAccessibleDescription(
      "Integrations using it stop working.",
    );
  });

  it("offers a close button that calls onClose", async () => {
    const onClose = vi.fn();
    renderInDialog(onClose);

    await userEvent.click(screen.getByRole("button", { name: "Close" }));

    expect(onClose).toHaveBeenCalledOnce();
  });

  it("has no close button without onClose", () => {
    renderInDialog();

    expect(
      screen.queryByRole("button", { name: "Close" }),
    ).not.toBeInTheDocument();
  });
});
