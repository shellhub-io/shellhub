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
  it("resets the named user's password and moves to the result step", async () => {
    setResetResponse("gen-pass-123");
    renderDialog({ userId: "user-abc" });

    await userEvent.click(screen.getByRole("button", { name: /enable/i }));

    await waitFor(() =>
      expect(resetSpy).toHaveBeenCalledWith(
        expect.objectContaining({ path: { id: "user-abc" } }),
      ),
    );
    expect(screen.getByText("Password Generated")).toBeInTheDocument();
  });

  describe("enable flow — error states", () => {
    it.each([
      [
        "a 400 (user already has a password)",
        () => HttpResponse.json({}, { status: 400 }),
        /already has a local password/i,
      ],
      [
        "a non-400 status",
        () => HttpResponse.json({}, { status: 500 }),
        /failed to set password/i,
      ],
      [
        "a network failure",
        () => HttpResponse.error(),
        /failed to set password/i,
      ],
    ] as const)("%s reports '%s' and stays on the confirm step", async (
      _label,
      resolver,
      message,
    ) => {
      server.use(
        http.patch("*/admin/api/users/:id/password/reset", resolver),
      );
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /enable/i }));

      await waitFor(() =>
        expect(screen.getByText(message)).toBeInTheDocument(),
      );
      expect(screen.queryByText("Password Generated")).not.toBeInTheDocument();
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
