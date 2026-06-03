import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, beforeEach, vi } from "vitest";
import CommandPaletteTrigger from "../CommandPaletteTrigger";
import { useCommandPaletteStore } from "@/stores/commandPaletteStore";

describe("CommandPaletteTrigger", () => {
  beforeEach(() => {
    useCommandPaletteStore.setState({ open: false });
  });

  it("renders an input-style trigger with label and shortcut when expanded", () => {
    render(<CommandPaletteTrigger expanded />);

    const button = screen.getByRole("button", { name: "Open command palette" });
    expect(button).toHaveAttribute("aria-haspopup", "dialog");
    expect(screen.getByText("Quick connect…")).toBeInTheDocument();
    expect(screen.getByText("⌘K")).toBeInTheDocument();
  });

  it("renders an icon button with a title when collapsed", () => {
    render(<CommandPaletteTrigger expanded={false} />);

    const button = screen.getByRole("button", { name: "Open command palette" });
    expect(button).toHaveAttribute("title", "Quick connect (⌘K)");
    expect(screen.queryByText("Quick connect…")).not.toBeInTheDocument();
  });

  it("opens the palette on click", async () => {
    const user = userEvent.setup();
    render(<CommandPaletteTrigger expanded />);

    expect(useCommandPaletteStore.getState().open).toBe(false);
    await user.click(
      screen.getByRole("button", { name: "Open command palette" }),
    );
    expect(useCommandPaletteStore.getState().open).toBe(true);
  });

  it("runs onActivate alongside opening (e.g. to close the mobile drawer)", async () => {
    const user = userEvent.setup();
    const onActivate = vi.fn();
    render(<CommandPaletteTrigger expanded onActivate={onActivate} />);

    await user.click(
      screen.getByRole("button", { name: "Open command palette" }),
    );
    expect(useCommandPaletteStore.getState().open).toBe(true);
    expect(onActivate).toHaveBeenCalledOnce();
  });
});
