import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import FormCheckboxField from "@/components/common/fields/rhf/FormCheckboxField";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

interface FormValues {
  agree: boolean;
}

/**
 * Minimal harness that wires a FormCheckboxField into a real RHF form so we
 * can assert two-way binding, validation errors, and callbacks.
 */
function TestForm({
  defaultValue = false,
  error,
  onValueChange,
}: {
  defaultValue?: boolean;
  error?: string;
  onValueChange?: (v: boolean) => void;
}) {
  const { control } = useForm<FormValues>({
    defaultValues: { agree: defaultValue },
  });

  return (
    <FormCheckboxField
      name="agree"
      control={control}
      id="agree"
      label="I agree"
      error={error}
      onValueChange={onValueChange}
    />
  );
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("FormCheckboxField", () => {
  it("binds the checked state two-way from form default value (false)", () => {
    render(<TestForm defaultValue={false} />);

    expect(screen.getByLabelText("I agree")).not.toBeChecked();
  });

  it("binds the checked state two-way from form default value (true)", () => {
    render(<TestForm defaultValue={true} />);

    expect(screen.getByLabelText("I agree")).toBeChecked();
  });

  it("toggles the checkbox on click (two-way binding)", async () => {
    const user = userEvent.setup();
    render(<TestForm defaultValue={false} />);

    const checkbox = screen.getByLabelText("I agree");
    expect(checkbox).not.toBeChecked();

    await user.click(checkbox);
    expect(checkbox).toBeChecked();

    await user.click(checkbox);
    expect(checkbox).not.toBeChecked();
  });

  it("renders fieldState.error.message via CheckboxField error prop", async () => {
    function FormWithError() {
      const { control, setError } = useForm<FormValues>({
        defaultValues: { agree: false },
      });

      useEffect(() => {
        setError("agree", { message: "You must agree to continue" });
      }, [setError]);

      return (
        <FormCheckboxField
          name="agree"
          control={control}
          id="agree"
          label="I agree"
        />
      );
    }

    render(<FormWithError />);

    expect(
      await screen.findByText("You must agree to continue"),
    ).toBeInTheDocument();
  });

  it("error override prop takes precedence over fieldState error", async () => {
    function FormWithBothErrors() {
      const { control, setError } = useForm<FormValues>({
        defaultValues: { agree: false },
      });

      useEffect(() => {
        setError("agree", { message: "Field-state error" });
      }, [setError]);

      return (
        <FormCheckboxField
          name="agree"
          control={control}
          id="agree"
          label="I agree"
          error="Override error"
        />
      );
    }

    render(<FormWithBothErrors />);

    expect(screen.getByText("Override error")).toBeInTheDocument();
    expect(screen.queryByText("Field-state error")).not.toBeInTheDocument();
  });

  it("calls onValueChange with the new boolean on each toggle", async () => {
    const onValueChange = vi.fn();
    const user = userEvent.setup();

    render(<TestForm onValueChange={onValueChange} />);

    const checkbox = screen.getByLabelText("I agree");

    await user.click(checkbox);
    expect(onValueChange).toHaveBeenCalledTimes(1);
    expect(onValueChange).toHaveBeenNthCalledWith(1, true);

    await user.click(checkbox);
    expect(onValueChange).toHaveBeenCalledTimes(2);
    expect(onValueChange).toHaveBeenNthCalledWith(2, false);
  });
});
