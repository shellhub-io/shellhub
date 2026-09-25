import { describe, it, expect, vi } from "vitest";
import { useState } from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { z } from "zod";
import FormModal from "@/components/common/FormModal";
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
  closeOnSubmit = false,
}: {
  onSubmit: (v: Values, form: UseFormReturn<Values>) => void | Promise<void>;
  initialName?: string;
  closeOnSubmit?: boolean;
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
      <FormModal
        form={form}
        onSubmit={async (values) => {
          await onSubmit(values, form);
          if (closeOnSubmit) setOpen(false);
        }}
        open={open}
        onClose={() => setOpen(false)}
        title="Test Form"
        submitLabel="Save"
      >
        <FormInputField
          name="name"
          control={form.control}
          id="name"
          label="Name"
        />
      </FormModal>
    </>
  );
}

describe("FormModal + useDrawerForm", () => {
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

    await waitFor(() =>
      expect(onSubmit).toHaveBeenCalledWith(
        { name: "hello" },
        expect.anything(),
      ),
    );
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

  it("resets the form to defaults when the modal is reopened", async () => {
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

describe("FormModal discarding changes", () => {
  const discardPrompt = () =>
    screen.queryByRole("dialog", { name: "Discard changes?" });

  it("closes at once from Cancel when nothing was typed", async () => {
    const user = userEvent.setup();
    render(<Harness onSubmit={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(discardPrompt()).not.toBeInTheDocument();
    expect(screen.queryByLabelText("Name")).not.toBeInTheDocument();
  });

  it("asks before Cancel throws away typed input", async () => {
    const user = userEvent.setup();
    render(<Harness onSubmit={vi.fn()} />);

    await user.type(screen.getByLabelText("Name"), "draft");
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(discardPrompt()).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toHaveValue("draft");
  });

  it("asks before Escape throws away typed input", async () => {
    const user = userEvent.setup();
    render(<Harness onSubmit={vi.fn()} />);

    await user.type(screen.getByLabelText("Name"), "draft");
    fireEvent(
      screen.getByRole("dialog", { name: "Test Form" }),
      new Event("cancel", { cancelable: true }),
    );

    expect(discardPrompt()).toBeInTheDocument();
  });

  it("asks before a backdrop click throws away typed input", async () => {
    const user = userEvent.setup();
    render(<Harness onSubmit={vi.fn()} />);

    await user.type(screen.getByLabelText("Name"), "draft");
    const modal = screen.getByRole("dialog", { name: "Test Form" });
    fireEvent.mouseDown(modal);
    fireEvent.click(modal);

    expect(discardPrompt()).toBeInTheDocument();
  });

  it("asks before the close button throws away typed input", async () => {
    const user = userEvent.setup();
    render(<Harness onSubmit={vi.fn()} />);

    await user.type(screen.getByLabelText("Name"), "draft");
    await user.click(screen.getByRole("button", { name: "Close" }));

    expect(discardPrompt()).toBeInTheDocument();
  });

  it("keeps editing when the discard is declined", async () => {
    const user = userEvent.setup();
    render(<Harness onSubmit={vi.fn()} />);

    await user.type(screen.getByLabelText("Name"), "draft");
    await user.click(screen.getByRole("button", { name: "Cancel" }));
    await user.click(screen.getByRole("button", { name: "Keep editing" }));

    expect(discardPrompt()).not.toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toHaveValue("draft");
  });

  it("closes once the discard is confirmed", async () => {
    const user = userEvent.setup();
    render(<Harness onSubmit={vi.fn()} />);

    await user.type(screen.getByLabelText("Name"), "draft");
    await user.click(screen.getByRole("button", { name: "Cancel" }));
    await user.click(screen.getByRole("button", { name: "Discard" }));

    expect(discardPrompt()).not.toBeInTheDocument();
    expect(screen.queryByLabelText("Name")).not.toBeInTheDocument();
  });

  it("closes without asking when a save closes it", async () => {
    const user = userEvent.setup();
    render(
      <Harness
        onSubmit={(_, form) => form.setValue("name", "saved")}
        closeOnSubmit
      />,
    );

    await user.type(screen.getByLabelText("Name"), "draft");
    const submit = screen.getByRole("button", { name: "Save" });
    await waitFor(() => expect(submit).toBeEnabled());
    await user.click(submit);

    await waitFor(() =>
      expect(screen.queryByLabelText("Name")).not.toBeInTheDocument(),
    );
    expect(discardPrompt()).not.toBeInTheDocument();
  });
});
