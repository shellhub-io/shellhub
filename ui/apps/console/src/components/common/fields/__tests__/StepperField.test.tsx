import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import StepperField from "@/components/common/fields/StepperField";

type Overrides = Partial<React.ComponentProps<typeof StepperField>>;

function renderField(overrides: Overrides = {}) {
  const props = {
    id: "test-stepper",
    label: "Count",
    value: 5,
    onChange: vi.fn(),
    min: 1,
    ...overrides,
  };
  render(<StepperField {...props} />);
  return props;
}

describe("StepperField", () => {
  it("renders the current value", () => {
    renderField({ value: 7 });
    expect(screen.getByRole("textbox", { name: "Count" })).toHaveValue("7");
  });

  it("increments the value when + is clicked", async () => {
    const user = userEvent.setup();
    const { onChange } = renderField();

    await user.click(screen.getByRole("button", { name: "Increase" }));

    expect(onChange).toHaveBeenCalledWith(6);
  });

  it("decrements the value when − is clicked", async () => {
    const user = userEvent.setup();
    const { onChange } = renderField();

    await user.click(screen.getByRole("button", { name: "Decrease" }));

    expect(onChange).toHaveBeenCalledWith(4);
  });

  it("disables the decrease button at min", () => {
    renderField({ value: 1, min: 1 });
    expect(screen.getByRole("button", { name: "Decrease" })).toBeDisabled();
  });

  it("disables the increase button at max", () => {
    renderField({ value: 10, max: 10 });
    expect(screen.getByRole("button", { name: "Increase" })).toBeDisabled();
  });

  it("does not disable the increase button when no max is set", () => {
    renderField({ value: 999 });
    expect(
      screen.getByRole("button", { name: "Increase" }),
    ).not.toBeDisabled();
  });

  it("calls onChange when a valid in-range number is typed", async () => {
    const user = userEvent.setup();
    const { onChange } = renderField({ min: 1, max: 100 });

    const input = screen.getByRole("textbox", { name: "Count" });
    await user.clear(input);
    await user.type(input, "42");

    expect(onChange).toHaveBeenCalledWith(42);
  });

  it("reverts out-of-range input on blur", async () => {
    const user = userEvent.setup();
    renderField({ min: 2 });

    const input = screen.getByRole("textbox", { name: "Count" });
    await user.clear(input);
    await user.type(input, "1");
    await user.tab();

    expect(input).toHaveValue("5");
  });

  it("resets empty input to min on blur", async () => {
    const user = userEvent.setup();
    const { onChange } = renderField({ min: 3 });

    const input = screen.getByRole("textbox", { name: "Count" });
    await user.clear(input);
    await user.tab();

    expect(onChange).toHaveBeenLastCalledWith(3);
  });

  it("prevents typing in readOnly mode", () => {
    renderField({ readOnly: true });
    expect(screen.getByRole("textbox", { name: "Count" })).toHaveAttribute(
      "readonly",
    );
  });

  it("still allows stepping in readOnly mode", async () => {
    const user = userEvent.setup();
    const { onChange } = renderField({ readOnly: true });

    await user.click(screen.getByRole("button", { name: "Increase" }));

    expect(onChange).toHaveBeenCalledWith(6);
  });
});
