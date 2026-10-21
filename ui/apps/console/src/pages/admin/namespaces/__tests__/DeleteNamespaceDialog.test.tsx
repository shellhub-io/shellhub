import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { setEdition } from "@/tests/edition";
import DeleteNamespaceDialog from "../DeleteNamespaceDialog";

vi.mock("@/components/common/ConfirmDialog", async () => ({
  default: (await import("@/tests/mocks")).MockConfirmDialog,
}));

const mockNamespace = {
  tenant_id: "tenant-xyz",
  name: "test-namespace",
};

const Wrapper = createTestWrapper();

const deleteSpy = vi.fn();

beforeEach(() => {
  vi.clearAllMocks();
  setEdition("community");
  deleteSpy.mockReset();
  server.use(
    http.delete("*/admin/api/namespaces/:tenant", ({ params }) => {
      deleteSpy({ path: { tenant: params.tenant } });
      return new HttpResponse(null, { status: 204 });
    }),
  );
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
  describe("confirm — success", () => {
    it("calls deleteNamespaceAdmin with the correct tenant_id", async () => {
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete namespace$/i }));

      await waitFor(() => {
        expect(deleteSpy).toHaveBeenCalledWith(
          expect.objectContaining({
            path: { tenant: "tenant-xyz" },
          }),
        );
      });
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

      await userEvent.click(screen.getByRole("button", { name: /^delete namespace$/i }));

      await waitFor(() => expect(onDeleted).toHaveBeenCalledTimes(1));
      expect(callOrder).toEqual(["onClose", "onDeleted"]);
    });
  });

  describe("confirm — error handling", () => {
    it("shows generic error message on failure", async () => {
      server.use(
        http.delete("*/admin/api/namespaces/:tenant", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete namespace$/i }));

      await waitFor(() => {
        expect(
          screen.getByText(/failed to delete namespace/i),
        ).toBeInTheDocument();
      });
    });

    it("explains a deletion that billing blocks", async () => {
      setEdition("cloud");
      server.use(
        http.delete("*/admin/api/namespaces/:tenant", () =>
          HttpResponse.json({}, { status: 402 }),
        ),
      );
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete namespace$/i }));

      await waitFor(() => {
        expect(
          screen.getByText(
            "This namespace has an active subscription or an unpaid invoice, which blocks its deletion.",
          ),
        ).toBeInTheDocument();
      });
    });

    it("does not blame billing for a 402 outside cloud, where only the license answers it", async () => {
      setEdition("enterprise");
      server.use(
        http.delete("*/admin/api/namespaces/:tenant", () =>
          HttpResponse.json({}, { status: 402 }),
        ),
      );
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete namespace$/i }));

      await waitFor(() => screen.getByText(/failed to delete namespace/i));
      expect(screen.queryByText(/active subscription/i)).not.toBeInTheDocument();
    });

    it("does not call onDeleted when deletion fails", async () => {
      server.use(
        http.delete("*/admin/api/namespaces/:tenant", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      const { onDeleted } = renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete namespace$/i }));

      await waitFor(() => screen.getByText(/failed to delete namespace/i));
      expect(onDeleted).not.toHaveBeenCalled();
    });

    it("does not call onClose when deletion fails", async () => {
      server.use(
        http.delete("*/admin/api/namespaces/:tenant", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      const { onClose } = renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete namespace$/i }));

      await waitFor(() => screen.getByText(/failed to delete namespace/i));
      expect(onClose).not.toHaveBeenCalled();
    });
  });

  describe("null namespace", () => {
    it("does not call deleteNamespaceAdmin when confirmed with null namespace", async () => {
      renderDialog({ namespace: null });
      await userEvent.click(screen.getByRole("button", { name: /^delete namespace$/i }));
      await waitFor(() => expect(deleteSpy).not.toHaveBeenCalled());
    });
  });
});
