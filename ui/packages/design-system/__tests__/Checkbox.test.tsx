import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Checkbox } from "../primitives/Checkbox";

function renderCheckbox(
  overrides: Partial<{
    checked: boolean;
    onChange: (checked: boolean) => void;
    disabled: boolean;
    id: string;
  }> = {},
) {
  const props = {
    checked: false,
    onChange: vi.fn(),
    ...overrides,
  };
  render(<Checkbox {...props} />);
  return props;
}

describe("Checkbox", () => {
  it("reflects the checked prop on the input", () => {
    renderCheckbox({ checked: true });
    expect(screen.getByRole("checkbox")).toBeChecked();
  });

  it("reflects unchecked state on the input", () => {
    renderCheckbox({ checked: false });
    expect(screen.getByRole("checkbox")).not.toBeChecked();
  });

  it("calls onChange with true when clicking an unchecked checkbox", async () => {
    const user = userEvent.setup();
    const { onChange } = renderCheckbox({ checked: false });

    await user.click(screen.getByRole("checkbox"));

    expect(onChange).toHaveBeenCalledWith(true);
  });

  it("calls onChange with false when clicking a checked checkbox", async () => {
    const user = userEvent.setup();
    const { onChange } = renderCheckbox({ checked: true });

    await user.click(screen.getByRole("checkbox"));

    expect(onChange).toHaveBeenCalledWith(false);
  });

  it("does not call onChange when disabled", async () => {
    const user = userEvent.setup();
    const { onChange } = renderCheckbox({ disabled: true });

    await user.click(screen.getByRole("checkbox"));

    expect(onChange).not.toHaveBeenCalled();
  });

  it("exposes the disabled state to assistive tech", () => {
    renderCheckbox({ disabled: true });
    expect(screen.getByRole("checkbox")).toBeDisabled();
  });

  it("forwards the id to the underlying input", () => {
    renderCheckbox({ id: "agree" });
    expect(screen.getByRole("checkbox")).toHaveAttribute("id", "agree");
  });

  it("forwards aria-required to the input", () => {
    render(
      <Checkbox checked={false} onChange={() => {}} aria-required={true} />,
    );
    expect(screen.getByRole("checkbox")).toHaveAttribute(
      "aria-required",
      "true",
    );
  });
});
