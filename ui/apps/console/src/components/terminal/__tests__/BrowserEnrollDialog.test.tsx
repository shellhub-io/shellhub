import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import BrowserEnrollDialog from "../BrowserEnrollDialog";

vi.mock("@/components/common/BaseDialog", async () => ({
  default: (await import("@/tests/mocks")).MockBaseDialog,
}));

const DEFAULT_NAME = "Chrome 149 on Linux";

function renderDialog(
  props?: Partial<Parameters<typeof BrowserEnrollDialog>[0]>,
) {
  const onClose = vi.fn();
  const onConfirm = vi
    .fn<(name: string) => Promise<void>>()
    .mockResolvedValue();
  render(
    <BrowserEnrollDialog
      open
      defaultName={DEFAULT_NAME}
      onClose={onClose}
      onConfirm={onConfirm}
      {...props}
    />,
  );
  return { onClose, onConfirm };
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("BrowserEnrollDialog", () => {
  it("registers under the prefilled name when accepted as-is", async () => {
    const user = userEvent.setup();
    const { onConfirm } = renderDialog();

    expect(screen.getByLabelText(/name/i)).toHaveValue(DEFAULT_NAME);
    await user.click(
      screen.getByRole("button", { name: /register and connect/i }),
    );

    await waitFor(() => expect(onConfirm).toHaveBeenCalledWith(DEFAULT_NAME));
  });

  it("registers under an edited name", async () => {
    const user = userEvent.setup();
    const { onConfirm } = renderDialog();

    const field = screen.getByLabelText(/name/i);
    await user.clear(field);
    await user.type(field, "My work laptop");
    await user.click(
      screen.getByRole("button", { name: /register and connect/i }),
    );

    await waitFor(() =>
      expect(onConfirm).toHaveBeenCalledWith("My work laptop"),
    );
  });

  it("falls back to the default name when the field is cleared", async () => {
    const user = userEvent.setup();
    const { onConfirm } = renderDialog();

    await user.clear(screen.getByLabelText(/name/i));
    await user.click(
      screen.getByRole("button", { name: /register and connect/i }),
    );

    await waitFor(() => expect(onConfirm).toHaveBeenCalledWith(DEFAULT_NAME));
  });

  it("cancels without registering", async () => {
    const user = userEvent.setup();
    const { onClose, onConfirm } = renderDialog();

    await user.click(screen.getByRole("button", { name: /cancel/i }));

    expect(onClose).toHaveBeenCalledOnce();
    expect(onConfirm).not.toHaveBeenCalled();
  });
});
