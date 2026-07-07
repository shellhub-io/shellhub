import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useEffect } from "react";
import { useForm, type FieldValues } from "react-hook-form";
import FormTextareaField from "@/components/common/fields/rhf/FormTextareaField";

interface TextareaFormValues extends FieldValues {
  notes: string;
}

function TextareaForm({
  defaultValue = "",
  onValueChange,
}: {
  defaultValue?: string;
  onValueChange?: (v: string) => void;
}) {
  const { control } = useForm<TextareaFormValues>({
    defaultValues: { notes: defaultValue },
  });

  return (
    <FormTextareaField<TextareaFormValues>
      name="notes"
      control={control}
      id="notes"
      label="Notes"
      onValueChange={onValueChange}
    />
  );
}

function TextareaFormWithFieldError({
  message,
  error,
}: {
  message: string;
  error?: string;
}) {
  const { control, setError } = useForm<TextareaFormValues>({
    defaultValues: { notes: "" },
  });

  useEffect(() => {
    setError("notes", { message });
  }, [setError, message]);

  return (
    <FormTextareaField<TextareaFormValues>
      name="notes"
      control={control}
      id="notes"
      label="Notes"
      error={error}
    />
  );
}

describe("FormTextareaField (RHF adapter contract)", () => {
  it("binds the initial value from the form state into the textarea", () => {
    render(<TextareaForm defaultValue="pre-filled content" />);

    expect(screen.getByLabelText("Notes")).toHaveValue("pre-filled content");
  });

  it("updates the textarea value as the user types (two-way binding)", async () => {
    const user = userEvent.setup();
    render(<TextareaForm />);

    const textarea = screen.getByLabelText("Notes");
    await user.type(textarea, "hello");

    expect(textarea).toHaveValue("hello");
  });

  it("calls onValueChange on each edit, forwarding the current field value", async () => {
    const onValueChange = vi.fn();
    const user = userEvent.setup();

    render(<TextareaForm onValueChange={onValueChange} />);

    await user.type(screen.getByLabelText("Notes"), "hi");

    expect(onValueChange).toHaveBeenCalledTimes(2);
    expect(onValueChange).toHaveBeenNthCalledWith(1, "h");
    expect(onValueChange).toHaveBeenNthCalledWith(2, "hi");
  });

  it("error override prop takes precedence over fieldState error", () => {
    render(
      <TextareaFormWithFieldError
        message="Field-state error"
        error="Override error"
      />,
    );

    expect(screen.getByText("Override error")).toBeInTheDocument();
    expect(screen.queryByText("Field-state error")).not.toBeInTheDocument();
  });

  it("sets aria-invalid to true when an error is present", async () => {
    render(<TextareaFormWithFieldError message="Required" />);

    const textarea = await screen.findByLabelText("Notes");

    expect(textarea).toHaveAttribute("aria-invalid", "true");
  });

  it("omits aria-invalid when there is no error", () => {
    render(<TextareaForm />);

    expect(screen.getByLabelText("Notes")).not.toHaveAttribute("aria-invalid");
  });

  it("sets aria-describedby to the error element id when an error is present", async () => {
    render(<TextareaFormWithFieldError message="Required" />);

    const textarea = await screen.findByLabelText("Notes");

    expect(textarea).toHaveAttribute("aria-describedby", "notes-error");
  });
});
