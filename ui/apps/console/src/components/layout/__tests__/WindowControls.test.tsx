import { afterEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import WindowControls from "../WindowControls";

function runInsideDesktopApp() {
  const win = {
    close: vi.fn(() => Promise.resolve()),
    minimize: vi.fn(() => Promise.resolve()),
    toggleMaximize: vi.fn(() => Promise.resolve()),
  };
  window.__TAURI__ = { window: { getCurrentWindow: () => win } };
  return win;
}

afterEach(() => {
  delete window.__TAURI__;
});

describe("WindowControls", () => {
  it("leaves the window buttons to the browser", () => {
    render(<WindowControls />);

    expect(
      screen.queryByRole("button", { name: "Close window" }),
    ).not.toBeInTheDocument();
  });

  it.each([
    ["Minimize window", "minimize"],
    ["Maximize window", "toggleMaximize"],
    ["Close window", "close"],
  ] as const)("%s drives the desktop window", async (name, method) => {
    const user = userEvent.setup();
    const win = runInsideDesktopApp();
    render(<WindowControls />);

    await user.click(screen.getByRole("button", { name }));

    expect(win[method]).toHaveBeenCalledOnce();
  });
});
