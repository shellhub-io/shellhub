import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useEffect, type ReactElement } from "react";
import { useForm, type Control, type FieldValues } from "react-hook-form";
import FormInputField from "@/components/common/fields/rhf/FormInputField";
import FormPasswordField from "@/components/common/fields/rhf/FormPasswordField";
import FormCheckboxField from "@/components/common/fields/rhf/FormCheckboxField";

// ---------------------------------------------------------------------------
// Shared adapter contract — text-like fields (FormInputField, FormPasswordField)
//
// FormInputField and FormPasswordField are near-identical react-hook-form
// adapters over a controlled text input. Rather than copy the same suite into
// two files, we describe the contract once and run it against each wrapper. The
// only thing that differs is which component renders the input, so each case
// provides a `render` closure that keeps full generic typing (no casts).
// ---------------------------------------------------------------------------

interface TextFormValues extends FieldValues {
  value: string;
}

type TextFieldRenderer = (props: {
  control: Control<TextFormValues>;
  label: string;
  error?: string;
  onValueChange?: (v: string) => void;
}) => ReactElement;

const TEXT_CASES: { name: string; label: string; render: TextFieldRenderer }[] = [
  {
    name: "FormInputField",
    label: "Username",
    render: ({ control, label, error, onValueChange }) => (
      <FormInputField<TextFormValues>
        name="value"
        control={control}
        id="value"
        label={label}
        error={error}
        onValueChange={onValueChange}
      />
    ),
  },
  {
    name: "FormPasswordField",
    label: "Password",
    render: ({ control, label, error, onValueChange }) => (
      <FormPasswordField<TextFormValues>
        name="value"
        control={control}
        id="value"
        label={label}
        error={error}
        onValueChange={onValueChange}
      />
    ),
  },
];

/** Wires a text field into a real RHF form so we can assert against real usage. */
function TextForm({
  render: renderField,
  label,
  defaultValue = "",
  onValueChange,
}: {
  render: TextFieldRenderer;
  label: string;
  defaultValue?: string;
  onValueChange?: (v: string) => void;
}) {
  const { control } = useForm<TextFormValues>({
    defaultValues: { value: defaultValue },
  });

  return renderField({ control, label, onValueChange });
}

/** Same harness, but seeds a fieldState error (and optionally an override). */
function TextFormWithFieldError({
  render: renderField,
  label,
  message,
  error,
}: {
  render: TextFieldRenderer;
  label: string;
  message: string;
  error?: string;
}) {
  const { control, setError } = useForm<TextFormValues>({
    defaultValues: { value: "" },
  });

  useEffect(() => {
    setError("value", { message });
  }, [setError, message]);

  return renderField({ control, label, error });
}

describe.each(TEXT_CASES)("$name (RHF adapter contract)", ({ label, render: renderField }) => {
  it("binds the initial value from the form state into the input", () => {
    render(<TextForm render={renderField} label={label} defaultValue="preset" />);

    expect(screen.getByLabelText(label)).toHaveValue("preset");
  });

  it("updates the input value as the user types (two-way binding)", async () => {
    const user = userEvent.setup();
    render(<TextForm render={renderField} label={label} />);

    const input = screen.getByLabelText(label);
    await user.type(input, "bob");

    expect(input).toHaveValue("bob");
  });

  it("surfaces fieldState.error.message via the underlying field", async () => {
    render(
      <TextFormWithFieldError
        render={renderField}
        label={label}
        message="This field is required"
      />,
    );

    expect(
      await screen.findByText("This field is required"),
    ).toBeInTheDocument();
  });

  it("error override prop takes precedence over fieldState error", () => {
    render(
      <TextFormWithFieldError
        render={renderField}
        label={label}
        message="Field-state error"
        error="Override error"
      />,
    );

    expect(screen.getByText("Override error")).toBeInTheDocument();
    expect(screen.queryByText("Field-state error")).not.toBeInTheDocument();
  });

  it("calls onValueChange on each edit, forwarding the current field value", async () => {
    const onValueChange = vi.fn();
    const user = userEvent.setup();

    render(
      <TextForm render={renderField} label={label} onValueChange={onValueChange} />,
    );

    await user.type(screen.getByLabelText(label), "hi");

    // RHF's controlled onChange fires per keystroke with the accumulated value.
    expect(onValueChange).toHaveBeenCalledTimes(2);
    expect(onValueChange).toHaveBeenNthCalledWith(1, "h");
    expect(onValueChange).toHaveBeenNthCalledWith(2, "hi");
  });
});

// ---------------------------------------------------------------------------
// FormCheckboxField — same contract, but boolean/checked semantics differ
// enough (toggle via click, checked instead of value) to warrant its own block.
// ---------------------------------------------------------------------------

interface CheckboxFormValues extends FieldValues {
  agree: boolean;
}

function CheckboxForm({
  defaultValue = false,
  error,
  onValueChange,
}: {
  defaultValue?: boolean;
  error?: string;
  onValueChange?: (v: boolean) => void;
}) {
  const { control } = useForm<CheckboxFormValues>({
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

function CheckboxFormWithFieldError({
  message,
  error,
}: {
  message: string;
  error?: string;
}) {
  const { control, setError } = useForm<CheckboxFormValues>({
    defaultValues: { agree: false },
  });

  useEffect(() => {
    setError("agree", { message });
  }, [setError, message]);

  return (
    <FormCheckboxField
      name="agree"
      control={control}
      id="agree"
      label="I agree"
      error={error}
    />
  );
}

describe("FormCheckboxField (RHF adapter contract)", () => {
  it("binds the checked state from the form default value (false)", () => {
    render(<CheckboxForm defaultValue={false} />);

    expect(screen.getByLabelText("I agree")).not.toBeChecked();
  });

  it("binds the checked state from the form default value (true)", () => {
    render(<CheckboxForm defaultValue />);

    expect(screen.getByLabelText("I agree")).toBeChecked();
  });

  it("toggles the checkbox on click (two-way binding)", async () => {
    const user = userEvent.setup();
    render(<CheckboxForm defaultValue={false} />);

    const checkbox = screen.getByLabelText("I agree");
    expect(checkbox).not.toBeChecked();

    await user.click(checkbox);
    expect(checkbox).toBeChecked();

    await user.click(checkbox);
    expect(checkbox).not.toBeChecked();
  });

  it("surfaces fieldState.error.message via the underlying field", async () => {
    render(<CheckboxFormWithFieldError message="You must agree to continue" />);

    expect(
      await screen.findByText("You must agree to continue"),
    ).toBeInTheDocument();
  });

  it("error override prop takes precedence over fieldState error", () => {
    render(
      <CheckboxFormWithFieldError
        message="Field-state error"
        error="Override error"
      />,
    );

    expect(screen.getByText("Override error")).toBeInTheDocument();
    expect(screen.queryByText("Field-state error")).not.toBeInTheDocument();
  });

  it("calls onValueChange with the new boolean on each toggle", async () => {
    const onValueChange = vi.fn();
    const user = userEvent.setup();

    render(<CheckboxForm onValueChange={onValueChange} />);

    const checkbox = screen.getByLabelText("I agree");

    await user.click(checkbox);
    expect(onValueChange).toHaveBeenCalledTimes(1);
    expect(onValueChange).toHaveBeenNthCalledWith(1, true);

    await user.click(checkbox);
    expect(onValueChange).toHaveBeenCalledTimes(2);
    expect(onValueChange).toHaveBeenNthCalledWith(2, false);
  });
});
