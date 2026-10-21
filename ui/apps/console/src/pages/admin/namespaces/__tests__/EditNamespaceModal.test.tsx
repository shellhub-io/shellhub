import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import type { Namespace } from "@/client";
import EditNamespaceModal from "../EditNamespaceModal";

vi.mock("@/components/common/Modal", async () => ({
  default: (await import("@/tests/mocks")).MockModal,
}));

const Wrapper = createTestWrapper();

const mockSettings = {
  session_record: true,
  connection_announcement: "hello",
  legacy: true,
  require_verified_device_key: false,
};

const mockNamespace: Namespace = {
  name: "my-namespace",
  owner: "owner-1",
  tenant_id: "tenant-abc",
  members: [],
  settings: mockSettings,
  max_devices: 10,
  created_at: "2024-01-01T00:00:00Z",
  billing: null,
  devices_pending_count: 0,
  devices_accepted_count: 3,
  devices_rejected_count: 0,
  devices_removed_count: 0,
};

const editSpy = vi.fn();

function renderModal(
  overrides: Partial<{
    open: boolean;
    onClose: () => void;
    namespace: Namespace | null;
  }> = {},
) {
  const defaults = { open: true, onClose: vi.fn(), namespace: mockNamespace };
  const props = { ...defaults, ...overrides };
  return {
    onClose: props.onClose,
    ...render(<EditNamespaceModal {...props} />, { wrapper: Wrapper }),
  };
}

describe("EditNamespaceModal", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    editSpy.mockReset();
    server.use(
      http.put(
        "*/admin/api/namespaces-update/:tenantID",
        async ({ request, params }) => {
          let body = await request.json();
          if (typeof body === "string") body = JSON.parse(body);
          editSpy({
            path: { tenantID: params.tenantID },
            body,
          });
          return HttpResponse.json({});
        },
      ),
    );
  });

  describe("form enabling", () => {
    it("disables submit button when name is cleared", async () => {
      renderModal();
      await userEvent.clear(screen.getByLabelText("Namespace Name"));
      expect(
        screen.getByRole("button", { name: /save changes/i }),
      ).toBeDisabled();
    });

    it("re-enables submit button when name is typed back in", async () => {
      renderModal();
      const nameInput = screen.getByLabelText("Namespace Name");
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, "new-name");
      expect(
        screen.getByRole("button", { name: /save changes/i }),
      ).not.toBeDisabled();
    });
  });

  describe("submit — success", () => {
    it("calls editNamespaceAdmin with the correct payload", async () => {
      renderModal();

      const nameInput = screen.getByLabelText("Namespace Name");
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, "updated-namespace");

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(editSpy).toHaveBeenCalledWith(
          expect.objectContaining({
            path: { tenantID: "tenant-abc" },
            body: expect.objectContaining({
              name: "updated-namespace",
              max_devices: 10,
              settings: expect.objectContaining({
                session_record: true,
              }),
            }),
          }),
        );
      });
    });

    it("spreads the original namespace fields into the body", async () => {
      renderModal();

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(editSpy).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({
              owner: "owner-1",
              tenant_id: "tenant-abc",
            }),
          }),
        );
      });
    });

    it("passes the updated session_record value when checkbox is toggled", async () => {
      renderModal();

      await userEvent.click(screen.getByLabelText(/session recording/i));

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(editSpy).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({
              settings: expect.objectContaining({ session_record: false }),
            }),
          }),
        );
      });
    });
  });

  describe("device signature enforcement", () => {
    it("shows whether the namespace enforces device signatures", () => {
      renderModal({
        namespace: {
          ...mockNamespace,
          settings: {
            ...mockSettings,
            require_verified_device_key: true,
          },
        },
      });

      expect(
        screen.getByLabelText(/device signature enforcement/i),
      ).toBeChecked();
    });

    it("sends the requirement once it is ticked", async () => {
      renderModal();

      await userEvent.click(
        screen.getByLabelText(/device signature enforcement/i),
      );
      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(editSpy).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({
              settings: expect.objectContaining({
                require_verified_device_key: true,
              }),
            }),
          }),
        );
      });
    });
  });

  describe("submit — error handling", () => {
    it("shows conflict error message for 409 responses", async () => {
      server.use(
        http.put("*/admin/api/namespaces-update/:tenantID", () =>
          HttpResponse.json({}, { status: 409 }),
        ),
      );
      renderModal();

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(
          screen.getByText("A namespace with this name already exists."),
        ).toBeInTheDocument();
      });
    });

    it("shows generic error for non-409 SDK errors", async () => {
      server.use(
        http.put("*/admin/api/namespaces-update/:tenantID", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      renderModal();

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(
          screen.getByText(/failed to update namespace/i),
        ).toBeInTheDocument();
      });
    });
  });

  describe("state reset on reopen", () => {
    it("reloads namespace data when modal is closed then reopened", async () => {
      const { rerender } = renderModal({ namespace: mockNamespace });

      const nameInput = screen.getByLabelText("Namespace Name");
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, "changed-name");

      rerender(
        <EditNamespaceModal
          open={false}
          onClose={vi.fn()}
          namespace={mockNamespace}
        />,
      );
      rerender(
        <EditNamespaceModal
          open={true}
          onClose={vi.fn()}
          namespace={mockNamespace}
        />,
      );

      expect(screen.getByLabelText("Namespace Name")).toHaveValue(
        "my-namespace",
      );
    });

    it("clears any error when closed then reopened", async () => {
      server.use(
        http.put("*/admin/api/namespaces-update/:tenantID", () =>
          HttpResponse.error(),
        ),
      );
      const { rerender } = renderModal({ namespace: mockNamespace });

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );
      await waitFor(() => screen.getByRole("alert"));

      rerender(
        <EditNamespaceModal
          open={false}
          onClose={vi.fn()}
          namespace={mockNamespace}
        />,
      );
      rerender(
        <EditNamespaceModal
          open={true}
          onClose={vi.fn()}
          namespace={mockNamespace}
        />,
      );

      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });
});
