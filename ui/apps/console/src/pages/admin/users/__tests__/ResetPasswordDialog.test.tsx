import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import ResetPasswordDialog from "../ResetPasswordDialog";

vi.mock("@/components/common/BaseDialog", async () => ({
  default: (await import("@/tests/mocks")).MockBaseDialog,
}));

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
}));

const Wrapper = createTestWrapper();

const resetSpy = vi.fn();

function setResetResponse(password: string) {
  server.use(
    http.patch("*/admin/api/users/:id/password/reset", ({ params }) => {
      resetSpy({ path: { id: params.id } });
      return HttpResponse.json({ password });
    }),
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  resetSpy.mockReset();
  setResetResponse("default-pw");
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
    ...render(<ResetPasswordDialog {...props} />, { wrapper: Wrapper }),
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

    it("does not call adminResetUserPassword when Cancel is clicked", async () => {
      renderDialog();
      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(resetSpy).not.toHaveBeenCalled();
    });
  });

  describe("enable flow — success", () => {
    it("calls adminResetUserPassword with the correct userId when Enable is clicked", async () => {
      setResetResponse("gen-pass-123");
      renderDialog({ userId: "user-abc" });

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() =>
        expect(resetSpy).toHaveBeenCalledWith(
          expect.objectContaining({
            path: { id: "user-abc" },
          }),
        ),
      );
    });

    it("transitions to the result step after successful reset", async () => {
      setResetResponse("gen-pass-123");
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => {
        expect(screen.getByText("Password Generated")).toBeInTheDocument();
      });
    });

    it("displays the generated password in an input field", async () => {
      setResetResponse("s3cr3t-pw");
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => {
        expect(screen.getByDisplayValue("s3cr3t-pw")).toBeInTheDocument();
      });
    });

    it("renders the 'Generated password' labelled input", async () => {
      setResetResponse("abc");
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => {
        expect(
          screen.getByLabelText(/generated password/i),
        ).toBeInTheDocument();
      });
    });

    it("renders a Copy button on the result step", async () => {
      setResetResponse("abc");
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => {
        expect(
          screen.getByRole("button", { name: /copy/i }),
        ).toBeInTheDocument();
      });
    });

    it("renders a Close button on the result step", async () => {
      setResetResponse("abc");
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => {
        expect(
          screen.getByRole("button", { name: "Close" }),
        ).toBeInTheDocument();
      });
    });

    it("calls onClose when Close is clicked on result step", async () => {
      setResetResponse("abc");
      const { onClose } = renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));
      await waitFor(() => screen.getByText("Password Generated"));
      await userEvent.click(screen.getByRole("button", { name: "Close" }));

      expect(onClose).toHaveBeenCalled();
    });
  });

  describe("enable flow — error states", () => {
    it("shows specific error message for status 400 (user already has password)", async () => {
      server.use(
        http.patch("*/admin/api/users/:id/password/reset", () =>
          HttpResponse.json({}, { status: 400 }),
        ),
      );
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => {
        expect(
          screen.getByText(/already has a local password/i),
        ).toBeInTheDocument();
      });
    });

    it("shows generic error message for non-400 errors", async () => {
      server.use(
        http.patch("*/admin/api/users/:id/password/reset", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => {
        expect(screen.getByText(/failed to set password/i)).toBeInTheDocument();
      });
    });

    it("shows generic error for non-SDK errors", async () => {
      server.use(
        http.patch("*/admin/api/users/:id/password/reset", () =>
          HttpResponse.error(),
        ),
      );
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => {
        expect(screen.getByText(/failed to set password/i)).toBeInTheDocument();
      });
    });

    it("renders error with role='alert'", async () => {
      server.use(
        http.patch("*/admin/api/users/:id/password/reset", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
    });

    it("stays on the confirm step when there is an error", async () => {
      server.use(
        http.patch("*/admin/api/users/:id/password/reset", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => {
        expect(screen.getByText(/failed to set password/i)).toBeInTheDocument();
      });
      expect(screen.queryByText("Password Generated")).not.toBeInTheDocument();
    });

    it("clears error and stays on confirm step — Enable button is still visible", async () => {
      server.use(
        http.patch("*/admin/api/users/:id/password/reset", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() => screen.getByRole("alert"));
      expect(
        screen.getByRole("button", { name: /enable/i }),
      ).toBeInTheDocument();
    });
  });

  describe("state reset on reopen", () => {
    it("resets to confirm step when dialog is closed then reopened", async () => {
      setResetResponse("pw");
      const { rerender } = renderDialog({ userId: "u1" });

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));
      await waitFor(() => screen.getByText("Password Generated"));

      rerender(
        <ResetPasswordDialog open={false} onClose={vi.fn()} userId="u1" />,
      );
      rerender(
        <ResetPasswordDialog open={true} onClose={vi.fn()} userId="u1" />,
      );

      expect(
        screen.getByText("Enable Local Authentication"),
      ).toBeInTheDocument();
      expect(screen.queryByText("Password Generated")).not.toBeInTheDocument();
    });
  });
});
