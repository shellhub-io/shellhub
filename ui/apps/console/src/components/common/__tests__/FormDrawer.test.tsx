import { describe, it, expect, vi } from "vitest";
import { useState } from "react";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { z } from "zod";
import FormDrawer from "@/components/common/FormDrawer";
import FormInputField from "@/components/common/fields/rhf/FormInputField";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import { UseFormReturn } from "react-hook-form";

const schema = z.object({
  name: z.string().min(1, "Name is required"),
});

type Values = z.infer<typeof schema>;

function Harness({
  onSubmit,
  initialName = "",
}: {
  onSubmit: (v: Values, form: UseFormReturn<Values>) => void | Promise<void>;
  initialName?: string;
}) {
  const [open, setOpen] = useState(true);
  const form = useDrawerForm(open, schema, { name: initialName });

  return (
    <>
      <button type="button" onClick={() => setOpen(false)}>
        external-close
      </button>
      <button type="button" onClick={() => setOpen(true)}>
        external-open
      </button>
      <FormDrawer
        form={form}
        onSubmit={(values) => onSubmit(values, form)}
        open={open}
        onClose={() => setOpen(false)}
        title="Test Form"
        submitLabel="Save"
      >
        <FormInputField name="name" control={form.control} id="name" label="Name" />
      </FormDrawer>
    </>
  );
}

describe("FormDrawer + useDrawerForm", () => {
  it("keeps submit disabled while the form is invalid and enables it once valid", async () => {
    const user = userEvent.setup();
    render(<Harness onSubmit={vi.fn()} />);

    const submit = screen.getByRole("button", { name: "Save" });
    expect(submit).toBeDisabled();

    await user.type(screen.getByLabelText("Name"), "hello");

    await waitFor(() => expect(submit).toBeEnabled());
  });

  it("submits the current values through handleSubmit once valid", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<Harness onSubmit={onSubmit} />);

    await user.type(screen.getByLabelText("Name"), "hello");

    const submit = screen.getByRole("button", { name: "Save" });
    await waitFor(() => expect(submit).toBeEnabled());
    await user.click(submit);

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit.mock.calls[0][0]).toEqual({ name: "hello" });
  });

  it("surfaces a root error set by the submit handler", async () => {
    const user = userEvent.setup();
    render(
      <Harness
        onSubmit={(_, form) => {
          form.setError("root", { message: "Server rejected it" });
        }}
      />,
    );

    await user.type(screen.getByLabelText("Name"), "value");

    const submit = screen.getByRole("button", { name: "Save" });
    await waitFor(() => expect(submit).toBeEnabled());
    await user.click(submit);

    await waitFor(() =>
      expect(screen.getByRole("alert")).toHaveTextContent("Server rejected it"),
    );
  });

  it("resets the form to defaults when the drawer is reopened", async () => {
    const user = userEvent.setup();
    render(<Harness onSubmit={vi.fn()} />);

    const input = screen.getByLabelText("Name");
    await user.type(input, "dirty");
    expect(input).toHaveValue("dirty");

    await user.click(screen.getByRole("button", { name: "external-close" }));
    await user.click(screen.getByRole("button", { name: "external-open" }));

    await waitFor(() => expect(screen.getByLabelText("Name")).toHaveValue(""));
  });
});
