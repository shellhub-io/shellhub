import React from "react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import ResetPasswordDialog from "../ResetPasswordDialog";
import { useResetUserPassword } from "../../../../hooks/useAdminUserMutations";

vi.mock("../../../../hooks/useAdminUserMutations", () => ({
  useResetUserPassword: vi.fn(),
}));

// BaseDialog renders open/close state; we flatten it to a simple div for test isolation.
vi.mock("../../../../components/common/BaseDialog", () => ({
  default: ({
    open,
    onClose,
    children,
  }: {
    open: boolean;
    onClose: () => void;
    children: React.ReactNode;
  }) => {
    if (!open) return null;
    return (
      <div role="dialog">
        <button onClick={onClose}>Close Dialog</button>
        {children}
      </div>
    );
  },
}));

vi.mock("../../../../components/common/CopyButton", () => ({
  default: ({ text }: { text: string }) => (
    <button type="button" aria-label="Copy">
      {text}
    </button>
  ),
}));

const mockMutateAsync = vi.fn();

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(useResetUserPassword).mockReturnValue({
    mutateAsync: mockMutateAsync,
  } as never);
});

function renderDialog(
  overrides: Partial<{
    open: boolean;
    onClose: () => void;
    userId: string;
  }> = {},
) {
  const defaults = { open: true, onClose: vi.fn(), userId: "user-123" };
  const props = { ...defaults, ...overrides };
  return {
    onClose: props.onClose,
    ...render(<ResetPasswordDialog {...props} />),
  };
}

describe("ResetPasswordDialog", () => {
  describe("rendering — closed", () => {
    it("renders nothing when open is false", () => {
      renderDialog({ open: false });
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });

  describe("rendering — confirm step (initial)", () => {
    it("renders the dialog when open is true", () => {
      renderDialog();
      expect(screen.getByRole("dialog")).toBeInTheDocument();
    });

    it("renders the 'Enable Local Authentication' heading", () => {
      renderDialog();
      expect(
        screen.getByText("Enable Local Authentication"),
      ).toBeInTheDocument();
    });

    it("renders the explanatory description text", () => {
      renderDialog();
      expect(screen.getByText(/temporary password/i)).toBeInTheDocument();
    });

    it("renders the Enable button", () => {
      renderDialog();
      expect(
        screen.getByRole("button", { name: /enable/i }),
      ).toBeInTheDocument();
    });

    it("renders the Cancel button", () => {
      renderDialog();
      expect(
        screen.getByRole("button", { name: /cancel/i }),
      ).toBeInTheDocument();
    });

    it("does not render the password result step content initially", () => {
      renderDialog();
      expect(screen.queryByText("Password Generated")).not.toBeInTheDocument();
    });
  });

  describe("cancel", () => {
    it("calls onClose when Cancel is clicked", async () => {
      const { onClose } = renderDialog();
      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(onClose).toHaveBeenCalledTimes(1);
    });

    it("does not call mutateAsync when Cancel is clicked", async () => {
      renderDialog();
      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(mockMutateAsync).not.toHaveBeenCalled();
    });
  });

  describe("enable flow — success", () => {
    it("calls mutateAsync with the correct userId when Enable is clicked", async () => {
      mockMutateAsync.mockResolvedValue({ password: "gen-pass-123" });
      renderDialog({ userId: "user-abc" });

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() =>
        expect(mockMutateAsync).toHaveBeenCalledWith({
          path: { id: "user-abc" },
        }),
      );
    });

    it("transitions to the result step after successful reset", async () => {
      mockMutateAsync.mockResolvedValue({ password: "gen-pass-123" });
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => {
        expect(screen.getByText("Password Generated")).toBeInTheDocument();
      });
    });

    it("displays the generated password in an input field", async () => {
      mockMutateAsync.mockResolvedValue({ password: "s3cr3t-pw" });
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => {
        expect(screen.getByDisplayValue("s3cr3t-pw")).toBeInTheDocument();
      });
    });

    it("renders the 'Generated password' labelled input", async () => {
      mockMutateAsync.mockResolvedValue({ password: "abc" });
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => {
        expect(
          screen.getByLabelText(/generated password/i),
        ).toBeInTheDocument();
      });
    });

    it("renders a Copy button on the result step", async () => {
      mockMutateAsync.mockResolvedValue({ password: "abc" });
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => {
        expect(
          screen.getByRole("button", { name: /copy/i }),
        ).toBeInTheDocument();
      });
    });

    it("renders a Close button on the result step", async () => {
      mockMutateAsync.mockResolvedValue({ password: "abc" });
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => {
        expect(
          screen.getByRole("button", { name: "Close" }),
        ).toBeInTheDocument();
      });
    });

    it("calls onClose when Close is clicked on result step", async () => {
      mockMutateAsync.mockResolvedValue({ password: "abc" });
      const { onClose } = renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));
      await waitFor(() => screen.getByText("Password Generated"));
      await userEvent.click(screen.getByRole("button", { name: "Close" }));

      expect(onClose).toHaveBeenCalled();
    });
  });

  describe("enable flow — error states", () => {
    it("shows specific error message for status 400 (user already has password)", async () => {
      mockMutateAsync.mockRejectedValue({ status: 400 });
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => {
        expect(
          screen.getByText(/already has a local password/i),
        ).toBeInTheDocument();
      });
    });

    it("shows generic error message for non-400 errors", async () => {
      mockMutateAsync.mockRejectedValue({ status: 500 });
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => {
        expect(screen.getByText(/failed to set password/i)).toBeInTheDocument();
      });
    });

    it("shows generic error for non-SDK errors", async () => {
      mockMutateAsync.mockRejectedValue(new Error("network error"));
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => {
        expect(screen.getByText(/failed to set password/i)).toBeInTheDocument();
      });
    });

    it("renders error with role='alert'", async () => {
      mockMutateAsync.mockRejectedValue({ status: 500 });
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
    });

    it("stays on the confirm step when there is an error", async () => {
      mockMutateAsync.mockRejectedValue({ status: 500 });
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => {
        expect(screen.getByText(/failed to set password/i)).toBeInTheDocument();
      });
      expect(screen.queryByText("Password Generated")).not.toBeInTheDocument();
    });

    it("clears error and stays on confirm step — Enable button is still visible", async () => {
      mockMutateAsync.mockRejectedValue({ status: 500 });
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => screen.getByRole("alert"));
      // The Enable button should still be present so the user can retry
      expect(
        screen.getByRole("button", { name: /enable/i }),
      ).toBeInTheDocument();
    });
  });

  describe("state reset on reopen", () => {
    it("resets to confirm step when dialog is closed then reopened", async () => {
      mockMutateAsync.mockResolvedValue({ password: "pw" });
      const { rerender } = renderDialog({ userId: "u1" });

      // Move to result step
      await userEvent.click(screen.getByRole("button", { name: /enable/i }));
      await waitFor(() => screen.getByText("Password Generated"));

      // Close and reopen
      rerender(
        <ResetPasswordDialog open={false} onClose={vi.fn()} userId="u1" />,
      );
      rerender(
        <ResetPasswordDialog open={true} onClose={vi.fn()} userId="u1" />,
      );

      // Should be back on confirm step
      expect(
        screen.getByText("Enable Local Authentication"),
      ).toBeInTheDocument();
      expect(screen.queryByText("Password Generated")).not.toBeInTheDocument();
    });
  });
});
