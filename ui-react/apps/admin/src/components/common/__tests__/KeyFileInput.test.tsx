// @vitest-environment jsdom
import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import KeyFileInput from "../KeyFileInput";

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
    ...rest
  } = props;
  return render(
    <KeyFileInput
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
      Promise.resolve().then(() => {
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
  describe("label", () => {
    it("renders the label text", () => {
      renderComponent({ label: "Private Key" });
      expect(screen.getByText("Private Key")).toBeInTheDocument();
    });
  });

  describe("mode toggle", () => {
    it("shows File and Text buttons when not disabled", () => {
      renderComponent();
      expect(screen.getByRole("button", { name: "File" })).toBeInTheDocument();
      expect(screen.getByRole("button", { name: "Text" })).toBeInTheDocument();
    });

    it("hides mode toggle buttons when disabled", () => {
      renderComponent({ disabled: true });
      expect(screen.queryByRole("button", { name: "File" })).not.toBeInTheDocument();
      expect(screen.queryByRole("button", { name: "Text" })).not.toBeInTheDocument();
    });

    it("shows the drop zone by default (file mode)", () => {
      renderComponent();
      // Drop zone is identifiable by the empty-state label
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

    it("switches back to drop zone when File button is clicked after Text", async () => {
      renderComponent();
      await userEvent.click(screen.getByRole("button", { name: "Text" }));
      await userEvent.click(screen.getByRole("button", { name: "File" }));
      expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
      expect(
        screen.getByText("Drop key file, paste, or browse"),
      ).toBeInTheDocument();
    });
  });

  describe("drop zone — empty state", () => {
    it("renders the empty label", () => {
      renderComponent();
      expect(
        screen.getByText("Drop key file, paste, or browse"),
      ).toBeInTheDocument();
    });

    it("renders a custom emptyLabel", () => {
      renderComponent({ emptyLabel: "Upload your public key" });
      expect(screen.getByText("Upload your public key")).toBeInTheDocument();
    });
  });

  describe("drop zone — loaded state", () => {
    it("renders the loaded label when a value is present", () => {
      renderComponent({ value: "ssh-rsa AAAA" });
      expect(screen.getByText("Key loaded")).toBeInTheDocument();
    });

    it("renders a custom loadedLabel", () => {
      renderComponent({ value: "ssh-rsa AAAA", loadedLabel: "Key ready" });
      expect(screen.getByText("Key ready")).toBeInTheDocument();
    });

    it("renders a Clear button when key is loaded", () => {
      renderComponent({ value: "ssh-rsa AAAA" });
      expect(screen.getByRole("button", { name: "Clear" })).toBeInTheDocument();
    });

    it("calls onChange('') when Clear is clicked", async () => {
      const onChange = vi.fn();
      renderComponent({ value: "ssh-rsa AAAA", onChange });
      await userEvent.click(screen.getByRole("button", { name: "Clear" }));
      expect(onChange).toHaveBeenCalledWith("");
    });
  });

  describe("drop zone — dragging state", () => {
    it("sets dragging visual when dragOver fires on the drop zone", () => {
      const { container } = renderComponent();
      // Find the drop zone div by a stable child element
      const dropZone = container.querySelector(
        "[ondragover], .border-dashed",
      ) as HTMLElement | null;
      if (!dropZone) throw new Error("drop zone not found");

      fireEvent.dragOver(dropZone, { preventDefault: () => {} });
      // After dragOver the border class shifts to primary color
      expect(dropZone.className).toMatch(/border-primary/);
    });

    it("removes dragging visual on dragLeave", () => {
      const { container } = renderComponent();
      const dropZone = container.querySelector(".border-dashed") as HTMLElement;
      fireEvent.dragOver(dropZone);
      fireEvent.dragLeave(dropZone);
      expect(dropZone.className).not.toMatch(/bg-primary\/5/);
    });
  });

  describe("drop zone — error state", () => {
    it("renders the error message", () => {
      renderComponent({ error: "Invalid key format" });
      expect(screen.getByText("Invalid key format")).toBeInTheDocument();
    });

    it("does not render an error message when error is null", () => {
      renderComponent({ error: null });
      expect(screen.queryByText(/invalid/i)).not.toBeInTheDocument();
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

      await waitFor(() => expect(onChange).toHaveBeenCalledWith("ssh-rsa AAAAB3NzaC1"));
      restore();
    });

    it("does not call onChange for files over 512 KB", async () => {
      const onChange = vi.fn();
      const { container } = renderComponent({ onChange });
      const dropZone = container.querySelector(".border-dashed") as HTMLElement;

      const bigContent = new Uint8Array(513 * 1024);
      const bigFile = new File([bigContent], "big.pem", { type: "text/plain" });
      fireEvent.drop(dropZone, { dataTransfer: { files: [bigFile] } });

      // Give any async processing time to run
      await new Promise((r) => setTimeout(r, 20));
      expect(onChange).not.toHaveBeenCalled();
    });
  });

  describe("file input via browse", () => {
    it("clicking the drop zone triggers the hidden file input", async () => {
      const { container } = renderComponent();
      const fileInput = container.querySelector(
        'input[type="file"]',
      ) as HTMLInputElement;
      const clickSpy = vi.spyOn(fileInput, "click");
      const dropZone = container.querySelector(".border-dashed") as HTMLElement;
      await userEvent.click(dropZone);
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
    it("renders a textarea (not the drop zone) when disabled", () => {
      renderComponent({ disabled: true });
      expect(screen.getByRole("textbox")).toBeInTheDocument();
      expect(screen.queryByText("Drop key file, paste, or browse")).not.toBeInTheDocument();
    });

    it("textarea is disabled", () => {
      renderComponent({ disabled: true });
      expect(screen.getByRole("textbox")).toBeDisabled();
    });

    it("renders disabledHint when disabled", () => {
      renderComponent({ disabled: true, disabledHint: "Cannot edit now" });
      expect(screen.getByText("Cannot edit now")).toBeInTheDocument();
    });

    it("does not render hint when disabled", () => {
      renderComponent({ disabled: true, hint: "Upload a key" });
      expect(screen.queryByText("Upload a key")).not.toBeInTheDocument();
    });
  });

  describe("hint", () => {
    it("renders the hint when not disabled", () => {
      renderComponent({ hint: "Paste or drag your key file" });
      expect(screen.getByText("Paste or drag your key file")).toBeInTheDocument();
    });

    it("does not render disabledHint when not disabled", () => {
      renderComponent({ disabledHint: "Read only" });
      expect(screen.queryByText("Read only")).not.toBeInTheDocument();
    });
  });

  describe("accessibility", () => {
    it("associates the label with the textarea via id in text mode", async () => {
      renderComponent({ id: "pub-key" });
      await userEvent.click(screen.getByRole("button", { name: "Text" }));
      expect(screen.getByRole("textbox")).toHaveAttribute("id", "pub-key");
      expect(screen.getByLabelText("Public Key")).toBeInTheDocument();
    });

    it("marks the textarea aria-invalid when error is provided (text mode)", async () => {
      renderComponent({ error: "Bad key", id: "pub-key" });
      await userEvent.click(screen.getByRole("button", { name: "Text" }));
      const textarea = screen.getByRole("textbox");
      expect(textarea).toHaveAttribute("aria-invalid", "true");
    });

    it("links the error paragraph via aria-describedby when id is provided (text mode)", async () => {
      renderComponent({ error: "Bad key", id: "pub-key" });
      await userEvent.click(screen.getByRole("button", { name: "Text" }));
      const textarea = screen.getByRole("textbox");
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
