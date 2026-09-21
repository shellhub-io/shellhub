import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { getConfig, defaultConfig } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import type { SdkHttpError } from "@/api/errors";
import { createTestWrapper } from "@/tests/wrapper";
import ActionDialog from "../ActionDialog";
import type { Action } from "@/hooks/useActionDialog";

function makeSdkError(status: number): Error & SdkHttpError {
  return Object.assign(new Error("Request failed"), {
    status,
    headers: new Headers(),
  });
}

vi.mock("../ConfirmDialog", async () => ({
  default: (await import("@/tests/mocks")).MockConfirmDialog,
}));

vi.mock("../BaseDialog", async () => ({
  default: (await import("@/tests/mocks")).MockBaseDialog,
}));

const mockNavigate = vi.fn();
vi.mock("react-router-dom", async () => {
  const actual =
    await vi.importActual<typeof import("react-router-dom")>(
      "react-router-dom",
    );
  return { ...actual, useNavigate: () => mockNavigate };
});

const mockGetConfig = vi.mocked(getConfig);

const entity = { uid: "uid-1", name: "my-device" };
const acceptAction: Action = { entity, operation: "accept" };
const rejectAction: Action = { entity, operation: "reject" };
const removeAction: Action = { entity, operation: "remove" };

const Wrapper = createTestWrapper({ initialEntries: ["/"] });

function renderDialog(
  overrides: Partial<{
    action: Action;
    entityType: "device" | "container";
    runAction: () => Promise<void>;
    onClose: () => void;
    onSuccess: () => void;
  }> = {},
) {
  const props = {
    action: acceptAction,
    entityType: "device" as const,
    runAction: vi.fn<() => Promise<void>>().mockResolvedValue(undefined),
    onClose: vi.fn(),
    onSuccess: vi.fn(),
    ...overrides,
  };
  render(
    <Wrapper>
      <ActionDialog {...props} />
    </Wrapper>,
  );
  return props;
}

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig });
  useAuthStore.setState({ role: "owner" });
});

describe("ActionDialog", () => {
  describe("title and confirm label", () => {
    it.each([
      [acceptAction, "device", "Accept Device", "Accept"],
      [rejectAction, "device", "Reject Device", "Reject"],
      [removeAction, "device", "Remove Device", "Remove"],
      [acceptAction, "container", "Accept Container", "Accept"],
    ] as const)(
      "renders '%s' as '%s' with a '%s' button",
      (action, entityType, title, confirmLabel) => {
        renderDialog({ action, entityType });
        expect(screen.getByText(title)).toBeInTheDocument();
        expect(
          screen.getByRole("button", { name: confirmLabel }),
        ).toBeInTheDocument();
      },
    );
  });

  describe("error handling — accept", () => {
    it.each([
      [402, /license/i],
      [403, /permission/i],
      [409, /already exists/i],
    ])("shows the %i message when accept fails", async (status, message) => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
      const runAction = vi.fn().mockRejectedValue(makeSdkError(status));
      renderDialog({ runAction });
      await userEvent.click(screen.getByRole("button", { name: "Accept" }));
      await waitFor(() =>
        expect(screen.getByRole("alert")).toHaveTextContent(message),
      );
    });
  });

  describe("error handling — reject/remove", () => {
    it.each([
      [rejectAction, "device", "Reject", /failed to reject device/i],
      [removeAction, "device", "Remove", /failed to remove device/i],
      [removeAction, "container", "Remove", /failed to remove container/i],
    ] as const)(
      "shows the generic error naming the %s and %s",
      async (action, entityType, confirmLabel, message) => {
        const runAction = vi.fn().mockRejectedValue(makeSdkError(500));
        renderDialog({ action, entityType, runAction });
        await userEvent.click(
          screen.getByRole("button", { name: confirmLabel }),
        );
        await waitFor(() =>
          expect(screen.getByRole("alert")).toHaveTextContent(message),
        );
      },
    );
  });

  describe("billing dialog — cloud 402 on accept", () => {
    beforeEach(() => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
    });

    it.each([
      ["device", "Device limit reached"],
      ["container", "Container limit reached"],
    ] as const)(
      "shows the %s billing ConfirmDialog for owners",
      async (entityType, title) => {
        useAuthStore.setState({ role: "owner" });
        const runAction = vi.fn().mockRejectedValue(makeSdkError(402));
        renderDialog({ entityType, runAction });
        await userEvent.click(screen.getByRole("button", { name: "Accept" }));
        await waitFor(() =>
          expect(screen.getByText(title)).toBeInTheDocument(),
        );
        expect(
          screen.getByRole("button", { name: "Go to billing" }),
        ).toBeInTheDocument();
        expect(
          screen.getByRole("button", { name: "Not now" }),
        ).toBeInTheDocument();
      },
    );

    it("shows BaseDialog with single Close button for non-owners", async () => {
      useAuthStore.setState({ role: "observer" });
      const runAction = vi.fn().mockRejectedValue(makeSdkError(402));
      renderDialog({ runAction });
      await userEvent.click(screen.getByRole("button", { name: "Accept" }));
      await waitFor(() =>
        expect(screen.getByText("Device limit reached")).toBeInTheDocument(),
      );
      expect(
        screen.getByRole("button", { name: /^Close$/ }),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: "Go to billing" }),
      ).not.toBeInTheDocument();
    });

    it("navigates to billing on owner confirm", async () => {
      useAuthStore.setState({ role: "owner" });
      const runAction = vi.fn().mockRejectedValue(makeSdkError(402));
      renderDialog({ runAction });
      await userEvent.click(screen.getByRole("button", { name: "Accept" }));
      await waitFor(() => screen.getByText("Device limit reached"));
      await userEvent.click(
        screen.getByRole("button", { name: "Go to billing" }),
      );
      expect(mockNavigate).toHaveBeenCalledWith("/settings#billing");
    });

    it("does not trigger billing dialog for reject 402", async () => {
      const runAction = vi.fn().mockRejectedValue(makeSdkError(402));
      renderDialog({ action: rejectAction, runAction });
      await userEvent.click(screen.getByRole("button", { name: "Reject" }));
      await waitFor(() =>
        expect(screen.getByRole("alert")).toBeInTheDocument(),
      );
      expect(
        screen.queryByText("Device limit reached"),
      ).not.toBeInTheDocument();
    });
  });
});
