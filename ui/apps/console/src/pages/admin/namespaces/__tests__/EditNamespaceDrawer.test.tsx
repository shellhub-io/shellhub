import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import type { Namespace } from "@/client/model";
import EditNamespaceDrawer from "../EditNamespaceDrawer";

vi.mock("@/components/common/Drawer", async () => ({
  default: (await import("@/tests/mocks")).MockDrawer,
}));

const Wrapper = createTestWrapper();

const mockNamespace: Namespace = {
  name: "my-namespace",
  owner: "owner-1",
  tenant_id: "tenant-abc",
  members: [],
  settings: {
    session_record: true,
    connection_announcement: "hello",
    ssh_access_mode: "legacy",
    ssh_legacy_allowed: true,
  },
  max_devices: 10,
  created_at: "2024-01-01T00:00:00Z",
  billing: null,
  devices_pending_count: 0,
  devices_accepted_count: 3,
  devices_rejected_count: 0,
};

const editSpy = vi.fn();

function renderDrawer(
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
    ...render(<EditNamespaceDrawer {...props} />, { wrapper: Wrapper }),
  };
}

describe("EditNamespaceDrawer", () => {
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

  describe("form pre-filling", () => {
    it("uses default max_devices of -1 when namespace has no max_devices", () => {
      renderDrawer({
        namespace: { ...mockNamespace, max_devices: -1 },
      });
      expect(screen.getByLabelText(/^max devices$/i)).toHaveValue("-1");
    });
  });

  describe("form enabling", () => {
    it("submit button is enabled when name is non-empty", async () => {
      renderDrawer();
      await waitFor(() =>
        expect(
          screen.getByRole("button", { name: /save changes/i }),
        ).not.toBeDisabled(),
      );
    });

    it("disables submit button when name is cleared", async () => {
      renderDrawer();
      await userEvent.clear(screen.getByLabelText("Namespace Name"));
      expect(
        screen.getByRole("button", { name: /save changes/i }),
      ).toBeDisabled();
    });

    it("re-enables submit button when name is typed back in", async () => {
      renderDrawer();
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
      renderDrawer();

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
      renderDrawer();

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
      renderDrawer();

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

  describe("submit — error handling", () => {
    it("shows conflict error message for 409 responses", async () => {
      server.use(
        http.put("*/admin/api/namespaces-update/:tenantID", () =>
          HttpResponse.json({}, { status: 409 }),
        ),
      );
      renderDrawer();

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
      renderDrawer();

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );

      await waitFor(() => {
        expect(
          screen.getByText(/failed to update namespace/i),
        ).toBeInTheDocument();
      });
    });

    it("shows generic error for non-SDK errors", async () => {
      server.use(
        http.put("*/admin/api/namespaces-update/:tenantID", () =>
          HttpResponse.error(),
        ),
      );
      renderDrawer();

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
    it("reloads namespace data when drawer is closed then reopened", async () => {
      const { rerender } = renderDrawer({ namespace: mockNamespace });

      const nameInput = screen.getByLabelText("Namespace Name");
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, "changed-name");

      rerender(
        <EditNamespaceDrawer
          open={false}
          onClose={vi.fn()}
          namespace={mockNamespace}
        />,
      );
      rerender(
        <EditNamespaceDrawer
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
      const { rerender } = renderDrawer({ namespace: mockNamespace });

      await userEvent.click(
        screen.getByRole("button", { name: /save changes/i }),
      );
      await waitFor(() => screen.getByRole("alert"));

      rerender(
        <EditNamespaceDrawer
          open={false}
          onClose={vi.fn()}
          namespace={mockNamespace}
        />,
      );
      rerender(
        <EditNamespaceDrawer
          open={true}
          onClose={vi.fn()}
          namespace={mockNamespace}
        />,
      );

      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });
});
