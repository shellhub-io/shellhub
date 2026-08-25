import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { getConfig, defaultConfig } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import { makeSdkError } from "@/tests/sdk";
import { createTestWrapper } from "@/tests/wrapper";
import ActionDialog from "../ActionDialog";
import type { Action } from "@/hooks/useActionDialog";

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
  describe("confirm flow", () => {
    it("renders the correct title and confirm label for accept", () => {
      renderDialog({ action: acceptAction });
      expect(screen.getByText("Accept Device")).toBeInTheDocument();
      expect(screen.getByRole("button", { name: "Accept" })).toBeInTheDocument();
    });

    it("renders the correct title for reject", () => {
      renderDialog({ action: rejectAction });
      expect(screen.getByText("Reject Device")).toBeInTheDocument();
    });

    it("renders the correct title for remove", () => {
      renderDialog({ action: removeAction });
      expect(screen.getByText("Remove Device")).toBeInTheDocument();
    });

    it("uses entityType in the title", () => {
      renderDialog({ action: acceptAction, entityType: "container" });
      expect(screen.getByText("Accept Container")).toBeInTheDocument();
    });

    it("calls onSuccess then onClose on successful confirm", async () => {
      const props = renderDialog();
      await userEvent.click(screen.getByRole("button", { name: "Accept" }));
      await waitFor(() => expect(props.onSuccess).toHaveBeenCalledWith("accept"));
      expect(props.onClose).toHaveBeenCalled();
    });

    it("does not call onSuccess on cancel", async () => {
      const props = renderDialog();
      await userEvent.click(screen.getByRole("button", { name: "Cancel" }));
      expect(props.onClose).toHaveBeenCalled();
      expect(props.onSuccess).not.toHaveBeenCalled();
    });
  });

  describe("error handling — accept", () => {
    it("shows accept error message for non-cloud 402", async () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "enterprise" });
      const runAction = vi.fn().mockRejectedValue(makeSdkError(402));
      renderDialog({ runAction });
      await userEvent.click(screen.getByRole("button", { name: "Accept" }));
      await waitFor(() => expect(screen.getByRole("alert")).toBeInTheDocument());
      expect(screen.getByRole("alert")).toHaveTextContent(/license/i);
    });

    it("shows permission error for 403", async () => {
      const runAction = vi.fn().mockRejectedValue(makeSdkError(403));
      renderDialog({ runAction });
      await userEvent.click(screen.getByRole("button", { name: "Accept" }));
      await waitFor(() => expect(screen.getByRole("alert")).toHaveTextContent(/permission/i));
    });

    it("shows rename error for 409", async () => {
      const runAction = vi.fn().mockRejectedValue(makeSdkError(409));
      renderDialog({ runAction });
      await userEvent.click(screen.getByRole("button", { name: "Accept" }));
      await waitFor(() => expect(screen.getByRole("alert")).toHaveTextContent(/already exists/i));
    });

    it("does not call onClose on error", async () => {
      const runAction = vi.fn().mockRejectedValue(makeSdkError(500));
      const props = renderDialog({ runAction });
      await userEvent.click(screen.getByRole("button", { name: "Accept" }));
      await waitFor(() => expect(screen.getByRole("alert")).toBeInTheDocument());
      expect(props.onClose).not.toHaveBeenCalled();
    });
  });

  describe("error handling — reject/remove", () => {
    it("shows generic error for reject failure", async () => {
      const runAction = vi.fn().mockRejectedValue(makeSdkError(500));
      renderDialog({ action: rejectAction, runAction });
      await userEvent.click(screen.getByRole("button", { name: "Reject" }));
      await waitFor(() =>
        expect(screen.getByRole("alert")).toHaveTextContent(/failed to reject device/i),
      );
    });

    it("shows generic error for remove failure", async () => {
      const runAction = vi.fn().mockRejectedValue(makeSdkError(500));
      renderDialog({ action: removeAction, runAction });
      await userEvent.click(screen.getByRole("button", { name: "Remove" }));
      await waitFor(() =>
        expect(screen.getByRole("alert")).toHaveTextContent(/failed to remove device/i),
      );
    });

    it("interpolates container in generic error", async () => {
      const runAction = vi.fn().mockRejectedValue(makeSdkError(500));
      renderDialog({ action: removeAction, entityType: "container", runAction });
      await userEvent.click(screen.getByRole("button", { name: "Remove" }));
      await waitFor(() =>
        expect(screen.getByRole("alert")).toHaveTextContent(/failed to remove container/i),
      );
    });
  });

  describe("billing dialog — cloud 402 on accept", () => {
    beforeEach(() => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
    });

    it("shows billing ConfirmDialog for owners", async () => {
      useAuthStore.setState({ role: "owner" });
      const runAction = vi.fn().mockRejectedValue(makeSdkError(402));
      renderDialog({ runAction });
      await userEvent.click(screen.getByRole("button", { name: "Accept" }));
      await waitFor(() =>
        expect(screen.getByText("Device limit reached")).toBeInTheDocument(),
      );
      expect(screen.getByRole("button", { name: "Go to billing" })).toBeInTheDocument();
      expect(screen.getByRole("button", { name: "Not now" })).toBeInTheDocument();
    });

    it("shows BaseDialog with single Close button for non-owners", async () => {
      useAuthStore.setState({ role: "observer" });
      const runAction = vi.fn().mockRejectedValue(makeSdkError(402));
      renderDialog({ runAction });
      await userEvent.click(screen.getByRole("button", { name: "Accept" }));
      await waitFor(() =>
        expect(screen.getByText("Device limit reached")).toBeInTheDocument(),
      );
      expect(screen.getByRole("button", { name: /^Close$/ })).toBeInTheDocument();
      expect(screen.queryByRole("button", { name: "Go to billing" })).not.toBeInTheDocument();
    });

    it("navigates to billing on owner confirm", async () => {
      useAuthStore.setState({ role: "owner" });
      const runAction = vi.fn().mockRejectedValue(makeSdkError(402));
      const props = renderDialog({ runAction });
      await userEvent.click(screen.getByRole("button", { name: "Accept" }));
      await waitFor(() => screen.getByText("Device limit reached"));
      await userEvent.click(screen.getByRole("button", { name: "Go to billing" }));
      expect(mockNavigate).toHaveBeenCalledWith("/settings#billing");
      expect(props.onClose).toHaveBeenCalled();
    });

    it("uses container label in billing title", async () => {
      const runAction = vi.fn().mockRejectedValue(makeSdkError(402));
      renderDialog({ entityType: "container", runAction });
      await userEvent.click(screen.getByRole("button", { name: "Accept" }));
      await waitFor(() =>
        expect(screen.getByText("Container limit reached")).toBeInTheDocument(),
      );
    });

    it("does not trigger billing dialog for reject 402", async () => {
      const runAction = vi.fn().mockRejectedValue(makeSdkError(402));
      renderDialog({ action: rejectAction, runAction });
      await userEvent.click(screen.getByRole("button", { name: "Reject" }));
      await waitFor(() => expect(screen.getByRole("alert")).toBeInTheDocument());
      expect(screen.queryByText("Device limit reached")).not.toBeInTheDocument();
    });
  });
});
