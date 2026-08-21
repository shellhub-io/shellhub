import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useEffect } from "react";
import { useForm, type FieldValues } from "react-hook-form";
import { createTestWrapper } from "@/tests/wrapper";
import { mockTags } from "@/tests/mockTags";
import FormTagsSelector from "@/components/common/fields/rhf/FormTagsSelector";

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return { ...actual, getTags: vi.fn() };
});

interface TagsFormValues extends FieldValues {
  tags: string[];
}

function TagsForm({
  defaultValue = [],
  error,
  fieldErrorMessage,
}: {
  defaultValue?: string[];
  error?: string;
  fieldErrorMessage?: string;
}) {
  const { control, setError } = useForm<TagsFormValues>({
    defaultValues: { tags: defaultValue },
  });

  useEffect(() => {
    if (fieldErrorMessage) setError("tags", { message: fieldErrorMessage });
  }, [setError, fieldErrorMessage]);

  return (
    <FormTagsSelector<TagsFormValues>
      name="tags"
      control={control}
      id="tags"
      label="Tags"
      error={error}
    />
  );
}

function renderForm(props: Parameters<typeof TagsForm>[0] = {}) {
  return render(<TagsForm {...props} />, { wrapper: createTestWrapper() });
}

describe("FormTagsSelector (RHF adapter contract)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockTags(["production", "staging", "dev"]);
  });

  it("renders the tags selector with the given label", () => {
    renderForm();
    expect(screen.getByPlaceholderText("Search tags...")).toBeInTheDocument();
  });

  it("selecting a tag adds it to the string[] form value", async () => {
    const user = userEvent.setup();
    renderForm();

    await user.click(screen.getByPlaceholderText("Search tags..."));
    await user.click(await screen.findByRole("option", { name: "production" }));

    expect(screen.getByText("production")).toBeInTheDocument();
  });

  it("selecting multiple tags updates the string[] value with all selected tags", async () => {
    const user = userEvent.setup();
    renderForm();

    await user.click(screen.getByPlaceholderText("Search tags..."));
    await user.click(await screen.findByRole("option", { name: "production" }));
    await user.click(screen.getByRole("option", { name: "staging" }));

    expect(screen.getByText("production")).toBeInTheDocument();
    expect(screen.getByText("staging")).toBeInTheDocument();
  });

  it("removing a tag via the Remove tag button updates the string[] value", async () => {
    const user = userEvent.setup();
    renderForm({ defaultValue: ["production", "staging"] });

    const removeButtons = screen.getAllByRole("button", { name: "Remove tag" });
    await user.click(removeButtons[0]);

    expect(screen.queryByText("production")).not.toBeInTheDocument();
    expect(screen.getByText("staging")).toBeInTheDocument();
  });

  it("surfaces fieldState.error.message via the underlying field", async () => {
    const fieldErrorMessage = "At least one tag is required";

    renderForm({ fieldErrorMessage });

    expect(await screen.findByText(fieldErrorMessage)).toBeInTheDocument();
  });

  it("error override prop takes precedence over fieldState error", () => {
    const fieldErrorMessage = "Field-state error";
    const errorOverride = "Override error";

    renderForm({ fieldErrorMessage, error: errorOverride });

    expect(screen.getByText(errorOverride)).toBeInTheDocument();
    expect(screen.queryByText(fieldErrorMessage)).not.toBeInTheDocument();
  });
});
