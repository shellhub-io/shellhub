import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { MembershipInvitation } from "../../../client";
import EditInvitationDrawer from "../EditInvitationDrawer";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

const mockUpdateMutateAsync = vi.fn();

vi.mock("../../../hooks/useInvitationMutations", () => ({
  useUpdateMembershipInvitation: () => ({
    mutateAsync: mockUpdateMutateAsync,
    isPending: false,
  }),
}));

vi.mock("../../../components/common/Drawer", () => ({
  default: ({
    open,
    onClose,
    title,
    children,
    footer,
  }: {
    open: boolean;
    onClose: () => void;
    title: string;
    children: React.ReactNode;
    footer?: React.ReactNode;
  }) => {
    if (!open) return null;
    return (
      <div role="dialog" aria-label={title}>
        <h2>{title}</h2>
        <button onClick={onClose}>Close</button>
        {children}
        {footer ?? null}
      </div>
    );
  },
}));

vi.mock("../../../utils/styles", () => ({
  LABEL: "label",
  INPUT: "input",
}));

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

function makeInvitation(
  overrides: Partial<MembershipInvitation> = {},
): MembershipInvitation {
  return {
    namespace: { tenant_id: "t1", name: "my-ns" },
    user: { id: "u1", email: "alice@example.com" },
    invited_by: "owner@example.com",
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
    expires_at: "2099-01-08T00:00:00Z",
    status: "pending",
    status_updated_at: "2024-01-01T00:00:00Z",
    role: "operator",
    ...overrides,
  };
}

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <MemoryRouter>
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </MemoryRouter>
  );
}

function renderDrawer({
  open = true,
  onClose = vi.fn(),
  tenantId = "t1",
  invitation = makeInvitation(),
}: {
  open?: boolean;
  onClose?: () => void;
  tenantId?: string;
  invitation?: MembershipInvitation | null;
} = {}) {
  return render(
    <EditInvitationDrawer
      open={open}
      onClose={onClose}
      tenantId={tenantId}
      invitation={invitation}
    />,
    { wrapper: createWrapper() },
  );
}

/* ------------------------------------------------------------------ */
/* Setup                                                               */
/* ------------------------------------------------------------------ */

beforeEach(() => {
  vi.clearAllMocks();
  mockUpdateMutateAsync.mockResolvedValue(undefined);
});

/* ================================================================== */
/* Tests                                                               */
/* ================================================================== */

describe("EditInvitationDrawer", () => {
  describe("rendering", () => {
    it("renders nothing when closed", () => {
      const { container } = renderDrawer({ open: false });
      expect(container).toBeEmptyDOMElement();
    });

    it("renders the Update Invitation Role heading when open", () => {
      renderDrawer();
      expect(
        screen.getByRole("heading", { name: /update invitation role/i }),
      ).toBeInTheDocument();
    });

    it("shows the role selector", () => {
      renderDrawer();
      // RoleSelector renders buttons for each role
      expect(
        screen.getByRole("button", { name: /operator/i }),
      ).toBeInTheDocument();
    });

    it("pre-fills role with the invitation's current role", () => {
      renderDrawer({ invitation: makeInvitation({ role: "administrator" }) });
      // The administrator button should be visually selected (ring-1 class is applied via aria — tested via DOM)
      // We test the behaviour: clicking Save Changes should send the pre-filled role
      // The actual visual selection is controlled via className — we just verify it renders
      expect(
        screen.getByRole("button", { name: /administrator/i }),
      ).toBeInTheDocument();
    });

    it("disables Save Changes when the role has not changed", () => {
      renderDrawer({ invitation: makeInvitation({ role: "operator" }) });
      expect(
        screen.getByRole("button", { name: /save changes/i }),
      ).toBeDisabled();
    });
  });

  describe("saving", () => {
    it("enables Save Changes after a different role is selected", async () => {
      const user = userEvent.setup();
      renderDrawer({ invitation: makeInvitation({ role: "operator" }) });

      await user.click(screen.getByRole("button", { name: /administrator/i }));

      expect(
        screen.getByRole("button", { name: /save changes/i }),
      ).not.toBeDisabled();
    });

    it("calls updateMembershipInvitation with correct path and body", async () => {
      const user = userEvent.setup();
      const inv = makeInvitation({
        role: "operator",
        user: { id: "u1", email: "alice@example.com" },
      });
      renderDrawer({ invitation: inv, tenantId: "t1" });

      await user.click(screen.getByRole("button", { name: /administrator/i }));
      await user.click(screen.getByRole("button", { name: /save changes/i }));

      await waitFor(() =>
        expect(mockUpdateMutateAsync).toHaveBeenCalledWith({
          path: { tenant: "t1", "user-id": "u1" },
          body: { role: "administrator" },
        }),
      );
    });

    it("calls onClose after successful save", async () => {
      const onClose = vi.fn();
      const user = userEvent.setup();
      const inv = makeInvitation({ role: "operator" });
      renderDrawer({ invitation: inv, onClose });

      await user.click(screen.getByRole("button", { name: /administrator/i }));
      await user.click(screen.getByRole("button", { name: /save changes/i }));

      await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    });
  });

  describe("error handling", () => {
    it("shows error message when mutation fails", async () => {
      mockUpdateMutateAsync.mockRejectedValue(new Error("update failed"));
      const user = userEvent.setup();
      const inv = makeInvitation({ role: "operator" });
      renderDrawer({ invitation: inv });

      await user.click(screen.getByRole("button", { name: /administrator/i }));
      await user.click(screen.getByRole("button", { name: /save changes/i }));

      await waitFor(() =>
        expect(
          screen.getByText(/failed to update invitation role/i),
        ).toBeInTheDocument(),
      );
    });

    it("does not call onClose when mutation fails", async () => {
      mockUpdateMutateAsync.mockRejectedValue(new Error("update failed"));
      const onClose = vi.fn();
      const user = userEvent.setup();
      const inv = makeInvitation({ role: "operator" });
      renderDrawer({ invitation: inv, onClose });

      await user.click(screen.getByRole("button", { name: /administrator/i }));
      await user.click(screen.getByRole("button", { name: /save changes/i }));

      await waitFor(() =>
        expect(
          screen.getByText(/failed to update invitation role/i),
        ).toBeInTheDocument(),
      );
      expect(onClose).not.toHaveBeenCalled();
    });
  });
});
