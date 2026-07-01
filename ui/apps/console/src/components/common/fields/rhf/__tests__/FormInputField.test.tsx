import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import FormInputField from "@/components/common/fields/rhf/FormInputField";
import {
  FormInputField as BarrelFormInputField,
  FormPasswordField,
  FormCheckboxField,
  SignUpFormValues,
} from "@/components/common/fields/rhf";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

interface FormValues {
  username: string;
}

/**
 * Minimal harness that wires a FormInputField into a real RHF form so we can
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
    defaultValues: { username: defaultValue },
  });

  return (
    <FormInputField
      name="username"
      control={control}
      id="username"
      label="Username"
      error={error}
      onValueChange={onValueChange}
    />
  );
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("FormInputField", () => {
  it("binds the initial value from the form state into the input", () => {
    render(<TestForm defaultValue="alice" />);

    expect(screen.getByLabelText("Username")).toHaveValue("alice");
  });

  it("updates the input value as the user types (two-way binding)", async () => {
    const user = userEvent.setup();
    render(<TestForm defaultValue="" />);

    const input = screen.getByLabelText("Username");
    await user.type(input, "bob");

    expect(input).toHaveValue("bob");
  });

  it("renders fieldState.error.message via InputField error prop", async () => {
    function FormWithError() {
      const { control, setError } = useForm<FormValues>({
        defaultValues: { username: "" },
      });

      useEffect(() => {
        setError("username", { message: "Username is required" });
      }, [setError]);

      return (
        <FormInputField
          name="username"
          control={control}
          id="username"
          label="Username"
        />
      );
    }

    render(<FormWithError />);

    expect(await screen.findByText("Username is required")).toBeInTheDocument();
  });

  it("error override prop takes precedence over fieldState error", async () => {
    function FormWithBothErrors() {
      const { control, setError } = useForm<FormValues>({
        defaultValues: { username: "" },
      });

      useEffect(() => {
        setError("username", { message: "Field-state error" });
      }, [setError]);

      return (
        <FormInputField
          name="username"
          control={control}
          id="username"
          label="Username"
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

    await user.type(screen.getByLabelText("Username"), "hi");

    // RHF's controlled onChange fires per keystroke with the accumulated value.
    expect(onValueChange).toHaveBeenCalledTimes(2);
    expect(onValueChange).toHaveBeenNthCalledWith(1, "h");
    expect(onValueChange).toHaveBeenNthCalledWith(2, "hi");
  });
});

// ---------------------------------------------------------------------------
// Barrel index re-exports
// ---------------------------------------------------------------------------

describe("rhf/index barrel", () => {
  it("re-exports FormInputField as a function", () => {
    expect(typeof BarrelFormInputField).toBe("function");
  });

  it("re-exports FormPasswordField as a function", () => {
    expect(typeof FormPasswordField).toBe("function");
  });

  it("re-exports FormCheckboxField as a function", () => {
    expect(typeof FormCheckboxField).toBe("function");
  });

  it("re-exports SignUpFormValues as a type (checked by import resolution)", () => {
    // SignUpFormValues is a type-only export; we verify it compiles by using it
    // as a type annotation here without a runtime assertion needed.
    const _typeCheck: SignUpFormValues = {
      name: "Alice",
      username: "alice",
      email: "alice@example.com",
      password: "Secret123!",
      confirmPassword: "Secret123!",
      acceptPrivacyPolicy: true,
    };

    expect(_typeCheck.username).toBe("alice");
  });
});
