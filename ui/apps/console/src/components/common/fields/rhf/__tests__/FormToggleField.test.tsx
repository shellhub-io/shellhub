import { describe, it, expect } from "vitest";
import { useEffect } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useForm, type FieldValues } from "react-hook-form";
import FormToggleField from "@/components/common/fields/rhf/FormToggleField";

interface ToggleFormValues extends FieldValues {
  enabled: boolean;
}

function ToggleForm({
  defaultValue = false,
  label = "Enable feature",
  error,
  fieldError,
  activeLabel,
  inactiveLabel,
}: {
  defaultValue?: boolean;
  label?: string;
  error?: string;
  fieldError?: string;
  activeLabel?: string;
  inactiveLabel?: string;
}) {
  const { control, setError } = useForm<ToggleFormValues>({
    defaultValues: { enabled: defaultValue },
  });

  useEffect(() => {
    if (fieldError) setError("enabled", { message: fieldError });
  }, [fieldError, setError]);

  return (
    <FormToggleField<ToggleFormValues>
      name="enabled"
      control={control}
      label={label}
      id="enabled-toggle"
      error={error}
      activeLabel={activeLabel}
      inactiveLabel={inactiveLabel}
    />
  );
}

describe("FormToggleField (RHF adapter)", () => {
  it("renders with a switch role and the given label", () => {
    render(<ToggleForm label="Enable feature" />);

    expect(
      screen.getByRole("switch", { name: "Enable feature" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Enable feature")).toBeInTheDocument();
  });

  it.each([false, true])(
    "reflects aria-checked=%s from the form value",
    (value) => {
      render(<ToggleForm defaultValue={value} />);

      expect(screen.getByRole("switch")).toHaveAttribute(
        "aria-checked",
        String(value),
      );
    },
  );

  it.each([
    { initial: false, expected: "true" },
    { initial: true, expected: "false" },
  ])(
    "toggles aria-checked to $expected when clicking while $initial",
    async ({ initial, expected }) => {
      const user = userEvent.setup();
      render(<ToggleForm defaultValue={initial} />);

      await user.click(screen.getByRole("switch"));

      expect(screen.getByRole("switch")).toHaveAttribute(
        "aria-checked",
        expected,
      );
    },
  );

  it.each([
    { value: true, visible: "Active", absent: "Inactive" },
    { value: false, visible: "Inactive", absent: "Active" },
  ])(
    "shows $visible text when value is $value",
    ({ value, visible, absent }) => {
      render(<ToggleForm defaultValue={value} />);

      expect(screen.getByText(visible)).toBeInTheDocument();
      expect(screen.queryByText(absent)).not.toBeInTheDocument();
    },
  );

  it("supports custom active/inactive labels", () => {
    render(
      <ToggleForm defaultValue={true} activeLabel="On" inactiveLabel="Off" />,
    );

    expect(screen.getByText("On")).toBeInTheDocument();
  });

  it("renders the field-level validation error and links it via aria-describedby", () => {
    render(<ToggleForm fieldError="Toggle is required" />);

    const message = screen.getByText("Toggle is required");
    expect(message).toBeInTheDocument();

    const toggle = screen.getByRole("switch");
    expect(toggle).toHaveAttribute("aria-invalid", "true");
    expect(toggle).toHaveAttribute("aria-describedby", message.id);
  });

  it("prefers an explicit error override over the field error", () => {
    render(<ToggleForm error="Override error" fieldError="Field error" />);

    expect(screen.getByText("Override error")).toBeInTheDocument();
    expect(screen.queryByText("Field error")).not.toBeInTheDocument();
  });
});
