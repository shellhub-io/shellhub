import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { useEffect } from "react";
import { useForm, type FieldValues } from "react-hook-form";
import FormNumericInput from "@/components/common/fields/rhf/FormNumericInput";

interface NumericFormValues extends FieldValues {
  count: string;
}

function NumericForm({
  defaultValue = "",
  error,
  fieldErrorMessage,
}: {
  defaultValue?: string;
  error?: string;
  fieldErrorMessage?: string;
}) {
  const { control, setError } = useForm<NumericFormValues>({
    defaultValues: { count: defaultValue },
  });

  useEffect(() => {
    if (fieldErrorMessage) setError("count", { message: fieldErrorMessage });
  }, [setError, fieldErrorMessage]);

  return (
    <FormNumericInput<NumericFormValues>
      name="count"
      control={control}
      id="count"
      label="Count"
      error={error}
    />
  );
}

describe("FormNumericInput (RHF adapter contract)", () => {
  it("reflects the control value in the input", () => {
    render(<NumericForm defaultValue="42" />);

    expect(screen.getByLabelText("Count")).toHaveValue("42");
  });

  it("surfaces fieldState.error.message as visible error text", async () => {
    render(<NumericForm fieldErrorMessage="Must be a number" />);

    expect(await screen.findByText("Must be a number")).toBeInTheDocument();
  });

  it("error override prop takes precedence over fieldState error", () => {
    const fieldErrorMessage = "Field-state error";

    render(
      <NumericForm
        fieldErrorMessage={fieldErrorMessage}
        error="Override error"
      />,
    );

    expect(screen.getByText("Override error")).toBeInTheDocument();
    expect(screen.queryByText(fieldErrorMessage)).not.toBeInTheDocument();
  });
});
