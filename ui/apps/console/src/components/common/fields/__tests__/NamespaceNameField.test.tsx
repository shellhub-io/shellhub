import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import NamespaceNameField from "@/components/common/fields/NamespaceNameField";
import { NAMESPACE_NAME_HINT } from "@/utils/validation";

afterEach(cleanup);

function renderField(
  props: Partial<React.ComponentProps<typeof NamespaceNameField>> = {},
) {
  const onChange = vi.fn();
  const utils = render(
    <NamespaceNameField id="ns-name" value="" onChange={onChange} {...props} />,
  );
  return { onChange, ...utils };
}

describe("NamespaceNameField", () => {
  it("renders with the canonical label and placeholder", () => {
    renderField();
    expect(screen.getByLabelText("Namespace Name")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("my-namespace")).toBeInTheDocument();
  });

  it("renders the canonical hint when no error is present", () => {
    renderField();
    expect(screen.getByText(NAMESPACE_NAME_HINT)).toBeInTheDocument();
  });

  it("lowercases input before calling onChange", async () => {
    const user = userEvent.setup();
    const { onChange } = renderField();
    await user.type(screen.getByLabelText("Namespace Name"), "Ab");
    // userEvent.type fires one onChange per keystroke; both should be lowercased.
    expect(onChange).toHaveBeenCalledWith("a");
    expect(onChange).toHaveBeenCalledWith("b");
  });

  it("enforces the 30-character maxLength on the underlying input", () => {
    renderField();
    expect(screen.getByLabelText("Namespace Name")).toHaveAttribute(
      "maxLength",
      "30",
    );
  });

  it("hides the hint when error is set", () => {
    renderField({ error: "Name must be at least 3 characters" });
    expect(screen.queryByText(NAMESPACE_NAME_HINT)).not.toBeInTheDocument();
  });

  it("marks the input as aria-invalid when error is set", () => {
    renderField({ error: "bad" });
    expect(screen.getByLabelText("Namespace Name")).toHaveAttribute(
      "aria-invalid",
      "true",
    );
  });

  it("treats null error the same as no error", () => {
    renderField({ error: null });
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.getByText(NAMESPACE_NAME_HINT)).toBeInTheDocument();
  });
});
