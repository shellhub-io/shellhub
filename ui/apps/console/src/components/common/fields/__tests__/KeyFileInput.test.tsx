import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import KeyFileInput from "@/components/common/fields/KeyFileInput";
const noop = () => {};
const alwaysValid = () => true;

interface Props {
  value?: string;
  onChange?: (v: string) => void;
  validate?: (t: string) => boolean;
  disabled?: boolean;
  error?: string | null;
  label?: string;
  id?: string;
  hint?: string;
  disabledHint?: string;
  loadedLabel?: string;
  emptyLabel?: string;
  onFileName?: (name: string) => void;
}

function renderComponent(props: Props = {}) {
  const {
    value = "",
    onChange = noop,
    validate = alwaysValid,
    label = "Public Key",
    id = "key-file-input",
    ...rest
  } = props;
  return render(
    <KeyFileInput
      id={id}
      label={label}
      value={value}
      onChange={onChange}
      validate={validate}
      {...rest}
    />,
  );
}

function mockFileReader(content: string) {
  const original = globalThis.FileReader;

  class MockFileReader extends EventTarget {
    result: string | null = null;
    onload: (() => void) | null = null;

    readAsText() {
      void Promise.resolve().then(() => {
        this.result = content;
        if (this.onload) this.onload();
      });
    }
  }

  // @ts-expect-error - partial mock
  globalThis.FileReader = MockFileReader;
  return () => {
    globalThis.FileReader = original;
  };
}

describe("KeyFileInput", () => {
  describe("mode toggle", () => {
    it("hides mode toggle buttons when disabled", () => {
      renderComponent({ disabled: true });
      expect(
        screen.queryByRole("button", { name: "File" }),
      ).not.toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: "Text" }),
      ).not.toBeInTheDocument();
    });

    it.each([
      ["by default", [] as string[]],
      ["after switching to Text and back to File", ["Text", "File"]],
    ])("shows the drop zone %s", async (_label, clicks) => {
      renderComponent();
      for (const name of clicks) {
        await userEvent.click(screen.getByRole("button", { name }));
      }
      expect(
        screen.getByText("Drop key file, paste, or browse"),
      ).toBeInTheDocument();
      expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
    });

    it("switches to textarea when Text button is clicked", async () => {
      renderComponent();
      await userEvent.click(screen.getByRole("button", { name: "Text" }));
      expect(screen.getByRole("textbox")).toBeInTheDocument();
    });
  });

  describe("drop zone", () => {
    it("renders a custom emptyLabel", () => {
      renderComponent({ emptyLabel: "Upload your public key" });
      expect(screen.getByText("Upload your public key")).toBeInTheDocument();
    });

    it.each([
      [undefined, "Key loaded"],
      ["Key ready", "Key ready"],
    ])("loadedLabel=%s renders as '%s'", (loadedLabel, expected) => {
      renderComponent({ value: "ssh-rsa AAAA", loadedLabel });
      expect(screen.getByText(expected)).toBeInTheDocument();
    });

    it("calls onChange('') when Clear is clicked", async () => {
      const onChange = vi.fn();
      renderComponent({ value: "ssh-rsa AAAA", onChange });
      await userEvent.click(screen.getByRole("button", { name: "Clear" }));
      expect(onChange).toHaveBeenCalledWith("");
    });

    it("renders the error message", () => {
      renderComponent({ error: "Invalid key format" });
      expect(screen.getByText("Invalid key format")).toBeInTheDocument();
    });
  });

  describe("drag and drop", () => {
    it("calls onChange with file content after a valid drop", async () => {
      const restore = mockFileReader("ssh-rsa AAAAB3NzaC1");
      const onChange = vi.fn();
      const { container } = renderComponent({ onChange });
      const dropZone = container.querySelector(".border-dashed") as HTMLElement;

      const file = new File(["ssh-rsa AAAAB3NzaC1"], "id_rsa.pub", {
        type: "text/plain",
      });
      const dataTransfer = { files: [file] };
      fireEvent.drop(dropZone, { dataTransfer });

      await waitFor(() =>
        expect(onChange).toHaveBeenCalledWith("ssh-rsa AAAAB3NzaC1"),
      );
      restore();
    });

    it("does not call onChange for files over 512 KB", async () => {
      const onChange = vi.fn();
      const { container } = renderComponent({ onChange });
      const dropZone = container.querySelector(".border-dashed") as HTMLElement;

      const bigContent = new Uint8Array(513 * 1024);
      const bigFile = new File([bigContent], "big.pem", { type: "text/plain" });
      fireEvent.drop(dropZone, { dataTransfer: { files: [bigFile] } });

      await new Promise((r) => setTimeout(r, 20));
      expect(onChange).not.toHaveBeenCalled();
    });
  });

  describe("file input via browse", () => {
    it("clicking the browse button triggers the hidden file input", async () => {
      const { container } = renderComponent();
      const fileInput = container.querySelector(
        'input[type="file"]',
      ) as HTMLInputElement;
      const clickSpy = vi.spyOn(fileInput, "click");
      await userEvent.click(
        screen.getByRole("button", {
          name: /drop key file, paste, or browse/i,
        }),
      );
      expect(clickSpy).toHaveBeenCalled();
    });

    it("calls onChange after selecting a file through the file input", async () => {
      const restore = mockFileReader("key text");
      const onChange = vi.fn();
      const { container } = renderComponent({ onChange });
      const fileInput = container.querySelector(
        'input[type="file"]',
      ) as HTMLInputElement;

      const file = new File(["key text"], "k.pem", { type: "text/plain" });
      fireEvent.change(fileInput, { target: { files: [file] } });

      await waitFor(() => expect(onChange).toHaveBeenCalledWith("key text"));
      restore();
    });
  });

  describe("disabled state", () => {
    it("renders a disabled textarea instead of the drop zone", () => {
      renderComponent({ disabled: true });
      expect(screen.getByRole("textbox")).toBeDisabled();
      expect(
        screen.queryByText("Drop key file, paste, or browse"),
      ).not.toBeInTheDocument();
    });

    it.each([
      [true, "Cannot edit now", "Upload a key"],
      [false, "Upload a key", "Cannot edit now"],
    ])("disabled=%s shows '%s' and not '%s'", (disabled, shown, hidden) => {
      renderComponent({
        disabled,
        hint: "Upload a key",
        disabledHint: "Cannot edit now",
      });
      expect(screen.getByText(shown)).toBeInTheDocument();
      expect(screen.queryByText(hidden)).not.toBeInTheDocument();
    });
  });

  describe("accessibility", () => {
    it("associates the label with the textarea via id in text mode", async () => {
      renderComponent({ id: "pub-key" });
      await userEvent.click(screen.getByRole("button", { name: "Text" }));
      expect(screen.getByRole("textbox")).toHaveAttribute("id", "pub-key");
      expect(screen.getByLabelText("Public Key")).toBeInTheDocument();
    });

    it("marks the textarea invalid and links the error paragraph in text mode", async () => {
      renderComponent({ error: "Bad key", id: "pub-key" });
      await userEvent.click(screen.getByRole("button", { name: "Text" }));
      const textarea = screen.getByRole("textbox");
      expect(textarea).toHaveAttribute("aria-invalid", "true");
      expect(textarea).toHaveAttribute("aria-describedby", "pub-key-error");
    });
  });

  describe("text mode textarea", () => {
    it("reflects the current value", async () => {
      renderComponent({ value: "ssh-rsa ABC" });
      await userEvent.click(screen.getByRole("button", { name: "Text" }));
      expect(screen.getByRole("textbox")).toHaveValue("ssh-rsa ABC");
    });

    it("calls onChange when the user types", async () => {
      const onChange = vi.fn();
      renderComponent({ onChange });
      await userEvent.click(screen.getByRole("button", { name: "Text" }));
      await userEvent.type(screen.getByRole("textbox"), "a");
      expect(onChange).toHaveBeenCalled();
    });
  });
});
