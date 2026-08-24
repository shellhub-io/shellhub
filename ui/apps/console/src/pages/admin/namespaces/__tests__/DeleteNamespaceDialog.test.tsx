import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import DeleteNamespaceDialog from "../DeleteNamespaceDialog";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    deleteNamespaceAdmin: vi.fn(),
  }),
);

vi.mock("@/components/common/ConfirmDialog", async () => ({
  default: (await import("@/tests/mocks")).MockConfirmDialog,
}));

const mockNamespace = {
  tenant_id: "tenant-xyz",
  name: "test-namespace",
};

const Wrapper = createTestWrapper();

beforeEach(() => {
  vi.clearAllMocks();
  sdk.deleteNamespaceAdmin.mockResolvedValue(mockSdkResponse(undefined));
});

function renderDialog(
  overrides: Partial<{
    open: boolean;
    onClose: () => void;
    namespace: typeof mockNamespace | null;
    onDeleted: () => void;
  }> = {},
) {
  const defaults = {
    open: true,
    onClose: vi.fn(),
    namespace: mockNamespace,
    onDeleted: vi.fn(),
  };
  const props = { ...defaults, ...overrides };
  return {
    onClose: props.onClose,
    onDeleted: props.onDeleted,
    ...render(
      <Wrapper>
        <DeleteNamespaceDialog {...props} />
      </Wrapper>,
    ),
  };
}

describe("DeleteNamespaceDialog", () => {
  describe("rendering — closed", () => {
    it("renders nothing when open is false", () => {
      renderDialog({ open: false });
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });

  describe("rendering — open", () => {
    it("renders the dialog when open is true", () => {
      renderDialog();
      expect(screen.getByRole("dialog")).toBeInTheDocument();
    });

    it("renders the 'Delete Namespace' title", () => {
      renderDialog();
      expect(screen.getByText("Delete Namespace")).toBeInTheDocument();
    });

    it("renders the namespace name in the description", () => {
      renderDialog();
      expect(screen.getByText("test-namespace")).toBeInTheDocument();
    });

    it("renders the cascade warning about devices, sessions, public keys, and API keys", () => {
      renderDialog();
      expect(
        screen.getByText(/devices.*sessions.*public keys.*api keys/i),
      ).toBeInTheDocument();
    });

    it("renders the 'Delete' confirm button", () => {
      renderDialog();
      expect(
        screen.getByRole("button", { name: /^delete$/i }),
      ).toBeInTheDocument();
    });

    it("renders the Cancel button", () => {
      renderDialog();
      expect(
        screen.getByRole("button", { name: /cancel/i }),
      ).toBeInTheDocument();
    });
  });

  describe("confirm — success", () => {
    it("calls deleteNamespaceAdmin with the correct tenant_id", async () => {
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => {
        expect(sdk.deleteNamespaceAdmin).toHaveBeenCalledWith(
          expect.objectContaining({
            path: { tenant: "tenant-xyz" },
          }),
        );
      });
    });

    it("calls onDeleted callback after successful deletion", async () => {
      const { onDeleted } = renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => expect(onDeleted).toHaveBeenCalledTimes(1));
    });

    it("calls onClose after successful deletion", async () => {
      const { onClose } = renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    });

    it("calls onClose before onDeleted", async () => {
      const callOrder: string[] = [];
      const onClose = vi.fn(() => callOrder.push("onClose"));
      const onDeleted = vi.fn(() => callOrder.push("onDeleted"));
      render(
        <Wrapper>
          <DeleteNamespaceDialog
            open={true}
            onClose={onClose}
            namespace={mockNamespace}
            onDeleted={onDeleted}
          />
        </Wrapper>,
      );

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => expect(onDeleted).toHaveBeenCalledTimes(1));
      expect(callOrder).toEqual(["onClose", "onDeleted"]);
    });
  });

  describe("confirm — error handling", () => {
    it("shows generic error message on failure", async () => {
      sdk.deleteNamespaceAdmin.mockRejectedValue(new Error("server error"));
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => {
        expect(
          screen.getByText(/failed to delete namespace/i),
        ).toBeInTheDocument();
      });
    });

    it("shows error for SDK errors", async () => {
      sdk.deleteNamespaceAdmin.mockRejectedValue({ status: 500 });
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => {
        expect(
          screen.getByText(/failed to delete namespace/i),
        ).toBeInTheDocument();
      });
    });

    it("does not call onDeleted when deletion fails", async () => {
      sdk.deleteNamespaceAdmin.mockRejectedValue(new Error("server error"));
      const { onDeleted } = renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => screen.getByText(/failed to delete namespace/i));
      expect(onDeleted).not.toHaveBeenCalled();
    });

    it("does not call onClose when deletion fails", async () => {
      sdk.deleteNamespaceAdmin.mockRejectedValue(new Error("server error"));
      const { onClose } = renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => screen.getByText(/failed to delete namespace/i));
      expect(onClose).not.toHaveBeenCalled();
    });
  });

  describe("cancel", () => {
    it("calls onClose when Cancel is clicked", async () => {
      const { onClose } = renderDialog();
      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(onClose).toHaveBeenCalledTimes(1);
    });

    it("does not call deleteNamespaceAdmin when Cancel is clicked", async () => {
      renderDialog();
      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(sdk.deleteNamespaceAdmin).not.toHaveBeenCalled();
    });

    it("does not call onDeleted when Cancel is clicked", async () => {
      const { onDeleted } = renderDialog();
      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(onDeleted).not.toHaveBeenCalled();
    });
  });

  describe("null namespace", () => {
    it("renders nothing meaningful in the description when namespace is null", () => {
      renderDialog({ namespace: null });
      expect(screen.queryByText("test-namespace")).not.toBeInTheDocument();
    });

    it("does not call deleteNamespaceAdmin when confirmed with null namespace", async () => {
      renderDialog({ namespace: null });
      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));
      await waitFor(() =>
        expect(sdk.deleteNamespaceAdmin).not.toHaveBeenCalled(),
      );
    });
  });

  describe("optional onDeleted callback", () => {
    it("does not throw when onDeleted is not provided and deletion succeeds", async () => {
      const { onClose } = renderDialog({ onDeleted: undefined });

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    });
  });
});
