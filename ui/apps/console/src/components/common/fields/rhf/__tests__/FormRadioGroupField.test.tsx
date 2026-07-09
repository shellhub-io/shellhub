import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useEffect } from "react";
import { useForm, type FieldValues } from "react-hook-form";
import RadioCard from "@/components/common/fields/RadioCard";
import FormRadioGroupField from "@/components/common/fields/rhf/FormRadioGroupField";

type Plan = "free" | "pro";

interface PlanFormValues extends FieldValues {
  plan: Plan;
}

const PLAN_OPTIONS = [
  { value: "free" as Plan, label: "Free", description: "Basic features" },
  { value: "pro" as Plan, label: "Pro", description: "Advanced features" },
] as const;

function PlanForm({
  defaultValue = "free",
  labelledBy,
  errorMessage,
}: {
  defaultValue?: Plan;
  labelledBy?: string;
  errorMessage?: string;
}) {
  const { control, setError } = useForm<PlanFormValues>({ defaultValues: { plan: defaultValue } });

  useEffect(() => {
    if (errorMessage) setError("plan", { message: errorMessage });
  }, [setError, errorMessage]);

  const labelProps = labelledBy ? { labelledBy } : { label: "Choose a plan" };

  return (
    <FormRadioGroupField<PlanFormValues, Plan>
      name="plan"
      control={control}
      {...labelProps}
    >
      {PLAN_OPTIONS.map((opt) => (
        <RadioCard
          key={opt.value}
          value={opt.value}
          label={opt.label}
          description={opt.description}
          icon={<span />}
        />
      ))}
    </FormRadioGroupField>
  );
}

describe("FormRadioGroupField (RHF adapter)", () => {
  it("renders the label and all radio options", () => {
    render(<PlanForm />);

    expect(screen.getByText("Choose a plan")).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /free/i })).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /pro/i })).toBeInTheDocument();
  });

  it("reflects the initial form value by checking the matching radio", () => {
    render(<PlanForm defaultValue="pro" />);

    expect(screen.getByRole("radio", { name: /pro/i })).toBeChecked();
    expect(screen.getByRole("radio", { name: /free/i })).not.toBeChecked();
  });

  it("updates the field value when the user selects a different RadioCard", async () => {
    const user = userEvent.setup();
    render(<PlanForm defaultValue="free" />);

    const proRadio = screen.getByRole("radio", { name: /pro/i });
    await user.click(proRadio);

    expect(proRadio).toBeChecked();
    expect(screen.getByRole("radio", { name: /free/i })).not.toBeChecked();
  });

  it("surfaces fieldState.error.message below the group", async () => {
    render(<PlanForm errorMessage="A plan selection is required" />);

    expect(
      await screen.findByText("A plan selection is required"),
    ).toBeInTheDocument();
  });

  it("accepts labelledBy instead of label (discriminated union)", () => {
    render(
      <>
        <span id="external-label">Billing plan</span>
        <PlanForm labelledBy="external-label" />
      </>,
    );

    const group = screen.getByRole("radiogroup");
    expect(group).toHaveAttribute("aria-labelledby", "external-label");
  });
});
