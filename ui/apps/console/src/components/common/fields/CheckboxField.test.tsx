import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import CheckboxField from "./CheckboxField";

function renderField(
  overrides: Partial<{
    checked: boolean;
    onChange: (checked: boolean) => void;
    disabled: boolean;
    label: string;
    description: string;
    hint: string;
    error: string;
  }> = {},
) {
  const props = {
    id: "test-cb",
    label: "Accept terms",
    checked: false,
    onChange: vi.fn(),
    ...overrides,
  };
  render(<CheckboxField {...props} />);
  return props;
}

describe("CheckboxField", () => {
  it("renders the label as accessible name for the checkbox", () => {
    renderField();
    expect(
      screen.getByRole("checkbox", { name: "Accept terms" }),
    ).toBeInTheDocument();
  });

  it("reflects the checked prop", () => {
    renderField({ checked: true });
    expect(screen.getByRole("checkbox")).toBeChecked();
  });

  it("calls onChange when clicked", async () => {
    const user = userEvent.setup();
    const { onChange } = renderField();

    await user.click(screen.getByRole("checkbox"));

    expect(onChange).toHaveBeenCalledWith(true);
  });

  it("exposes the disabled state", () => {
    renderField({ disabled: true });
    expect(screen.getByRole("checkbox")).toBeDisabled();
  });

  it("renders the description text", () => {
    renderField({ description: "You must agree to continue" });
    expect(screen.getByText("You must agree to continue")).toBeInTheDocument();
  });

  it("shows the error message and marks the input as invalid", () => {
    renderField({ error: "Required field" });

    expect(screen.getByText("Required field")).toBeInTheDocument();
    expect(screen.getByRole("checkbox")).toHaveAttribute(
      "aria-invalid",
      "true",
    );
  });

  it("shows the hint when there is no error", () => {
    renderField({ hint: "Optional" });
    expect(screen.getByText("Optional")).toBeInTheDocument();
  });
});
