import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import FormPasswordField from "@/components/common/fields/rhf/FormPasswordField";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

interface FormValues {
  password: string;
}

/**
 * Minimal harness that wires a FormPasswordField into a real RHF form so we can
 * assert two-way binding, validation errors, and callbacks.
 */
function TestForm({
  defaultValue = "",
  error,
  onValueChange,
}: {
  defaultValue?: string;
  error?: string;
  onValueChange?: (v: string) => void;
}) {
  const { control } = useForm<FormValues>({
    defaultValues: { password: defaultValue },
  });

  return (
    <FormPasswordField
      name="password"
      control={control}
      id="password"
      label="Password"
      error={error}
      onValueChange={onValueChange}
    />
  );
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("FormPasswordField", () => {
  it("binds the initial value from the form state into the input", () => {
    render(<TestForm defaultValue="secret" />);

    expect(screen.getByLabelText("Password")).toHaveValue("secret");
  });

  it("updates the input value as the user types (two-way binding)", async () => {
    const user = userEvent.setup();
    render(<TestForm defaultValue="" />);

    const input = screen.getByLabelText("Password");
    await user.type(input, "abc");

    expect(input).toHaveValue("abc");
  });

  it("renders fieldState.error.message via PasswordField error prop", async () => {
    function FormWithError() {
      const { control, setError } = useForm<FormValues>({
        defaultValues: { password: "" },
      });

      useEffect(() => {
        setError("password", { message: "Password is required" });
      }, [setError]);

      return (
        <FormPasswordField
          name="password"
          control={control}
          id="password"
          label="Password"
        />
      );
    }

    render(<FormWithError />);

    expect(await screen.findByText("Password is required")).toBeInTheDocument();
  });

  it("error override prop takes precedence over fieldState error", async () => {
    function FormWithBothErrors() {
      const { control, setError } = useForm<FormValues>({
        defaultValues: { password: "" },
      });

      useEffect(() => {
        setError("password", { message: "Field-state error" });
      }, [setError]);

      return (
        <FormPasswordField
          name="password"
          control={control}
          id="password"
          label="Password"
          error="Override error"
        />
      );
    }

    render(<FormWithBothErrors />);

    // Override is visible immediately; field-state error never appears.
    expect(screen.getByText("Override error")).toBeInTheDocument();
    expect(screen.queryByText("Field-state error")).not.toBeInTheDocument();
  });

  it("calls onValueChange on each edit, forwarding the current field value", async () => {
    const onValueChange = vi.fn();
    const user = userEvent.setup();

    render(<TestForm onValueChange={onValueChange} />);

    await user.type(screen.getByLabelText("Password"), "hi");

    // RHF's controlled onChange fires per keystroke with the accumulated value.
    expect(onValueChange).toHaveBeenCalledTimes(2);
    expect(onValueChange).toHaveBeenNthCalledWith(1, "h");
    expect(onValueChange).toHaveBeenNthCalledWith(2, "hi");
  });
});
