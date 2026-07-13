import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Toggle } from "../primitives/Toggle";

function renderToggle(
  overrides: Partial<{
    enabled: boolean;
    onChange: (enabled: boolean) => void;
    disabled: boolean;
  }> = {},
) {
  const props = {
    enabled: false,
    onChange: vi.fn(),
    "aria-label": "Test toggle",
    ...overrides,
  };
  render(<Toggle {...props} />);
  return props;
}

describe("Toggle", () => {
  it("renders as a switch with aria-checked reflecting the enabled prop", () => {
    renderToggle({ enabled: true });

    const toggle = screen.getByRole("switch", { name: "Test toggle" });
    expect(toggle).toHaveAttribute("aria-checked", "true");
  });

  it("reflects enabled=false as aria-checked=false", () => {
    renderToggle({ enabled: false });

    const toggle = screen.getByRole("switch", { name: "Test toggle" });
    expect(toggle).toHaveAttribute("aria-checked", "false");
  });

  it("calls onChange with the negated value on click", async () => {
    const user = userEvent.setup();
    const { onChange } = renderToggle({ enabled: false });

    await user.click(screen.getByRole("switch"));

    expect(onChange).toHaveBeenCalledWith(true);
  });

  it("does not call onChange when disabled", async () => {
    const user = userEvent.setup();
    const { onChange } = renderToggle({ disabled: true });

    await user.click(screen.getByRole("switch"));

    expect(onChange).not.toHaveBeenCalled();
  });

  it("exposes the disabled state to assistive tech", () => {
    renderToggle({ disabled: true });
    expect(screen.getByRole("switch")).toBeDisabled();
  });

  it("forwards aria-labelledby to the button", () => {
    render(
      <Toggle enabled={false} onChange={() => {}} aria-labelledby="my-label" />,
    );
    expect(screen.getByRole("switch")).toHaveAttribute(
      "aria-labelledby",
      "my-label",
    );
  });
});
