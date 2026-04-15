import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useAuthStore } from "../../../stores/authStore";
import type { MembershipInvitation } from "../../../client";
import InvitationsTab from "../InvitationsTab";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

const mockInvitations = vi.fn<
  () => {
    invitations: MembershipInvitation[];
    totalCount: number;
    isLoading: boolean;
  }
>();
const mockCancelMutateAsync = vi.fn();
const mockResendMutateAsync = vi.fn();

vi.mock("../../../hooks/useInvitations", () => ({
  useNamespaceInvitations: () => mockInvitations(),
}));

vi.mock("../../../hooks/useInvitationMutations", () => ({
  useCancelMembershipInvitation: () => ({
    mutateAsync: mockCancelMutateAsync,
    isPending: false,
  }),
  useGenerateInvitationLink: () => ({
    mutateAsync: mockResendMutateAsync,
    isPending: false,
  }),
}));

// Stub ConfirmDialog so jsdom's lack of HTMLDialogElement.showModal() doesn't crash
vi.mock("../../../components/common/ConfirmDialog", () => ({
  default: ({
    open,
    onClose,
    onConfirm,
    title,
    confirmLabel = "Confirm",
    cancelLabel = "Cancel",
  }: {
    open: boolean;
    onClose: () => void;
    onConfirm: () => Promise<void> | void;
    title: string;
    confirmLabel?: string;
    cancelLabel?: string;
  }) => {
    if (!open) return null;
    return (
      <div role="dialog">
        <h2>{title}</h2>
        <button onClick={onClose}>{cancelLabel}</button>
        <button onClick={() => void onConfirm()}>{confirmLabel}</button>
      </div>
    );
  },
}));

// Stub drawers to keep tests focused on InvitationsTab logic
vi.mock("../InvitationDrawer", () => ({
  default: ({ open, onClose }: { open: boolean; onClose: () => void }) =>
    open ? (
      <div data-testid="invitation-drawer">
        <button onClick={onClose}>Close Invite Drawer</button>
      </div>
    ) : null,
}));

vi.mock("../EditInvitationDrawer", () => ({
  default: ({ open, onClose }: { open: boolean; onClose: () => void }) =>
    open ? (
      <div data-testid="edit-invitation-drawer">
        <button onClick={onClose}>Close Edit Drawer</button>
      </div>
    ) : null,
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

function makeExpiredInvitation(
  overrides: Partial<MembershipInvitation> = {},
): MembershipInvitation {
  return makeInvitation({
    expires_at: "2020-01-01T00:00:00Z",
    ...overrides,
  });
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

function renderTab(tenantId = "t1") {
  return render(<InvitationsTab tenantId={tenantId} />, {
    wrapper: createWrapper(),
  });
}

/* ------------------------------------------------------------------ */
/* Setup                                                               */
/* ------------------------------------------------------------------ */

beforeEach(() => {
  vi.clearAllMocks();
  // Default to owner so RestrictedAction does not block anything
  useAuthStore.setState({ role: "owner" });
  mockInvitations.mockReturnValue({
    invitations: [],
    totalCount: 0,
    isLoading: false,
  });
  mockCancelMutateAsync.mockResolvedValue(undefined);
  mockResendMutateAsync.mockResolvedValue(undefined);
});

/* ================================================================== */
/* Tests                                                               */
/* ================================================================== */

describe("InvitationsTab", () => {
  describe("rendering", () => {
    it("shows the invitation count from totalCount", () => {
      mockInvitations.mockReturnValue({
        invitations: [],
        totalCount: 5,
        isLoading: false,
      });
      renderTab();
      expect(screen.getByText(/5 invitation/i)).toBeInTheDocument();
    });

    it("uses singular 'invitation' when count is 1", () => {
      mockInvitations.mockReturnValue({
        invitations: [],
        totalCount: 1,
        isLoading: false,
      });
      renderTab();
      // "1 invitation" without an 's'
      expect(screen.getByText("1 invitation")).toBeInTheDocument();
    });

    it("renders the status filter dropdown with Pending selected by default", () => {
      renderTab();
      const select = screen.getByRole("combobox", {
        name: /filter invitations by status/i,
      });
      expect(select).toHaveValue("pending");
    });

    it("renders all four status options in the dropdown", () => {
      renderTab();
      expect(
        screen.getByRole("option", { name: /pending/i }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("option", { name: /accepted/i }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("option", { name: /rejected/i }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("option", { name: /cancelled/i }),
      ).toBeInTheDocument();
    });

    it("shows the Invite Member button for owners", () => {
      renderTab();
      expect(
        screen.getByRole("button", { name: /invite member/i }),
      ).toBeInTheDocument();
    });

    it("shows a row for each invitation", () => {
      mockInvitations.mockReturnValue({
        invitations: [
          makeInvitation({ user: { id: "u1", email: "alice@example.com" } }),
          makeInvitation({ user: { id: "u2", email: "bob@example.com" } }),
        ],
        totalCount: 2,
        isLoading: false,
      });
      renderTab();
      expect(screen.getByText("alice@example.com")).toBeInTheDocument();
      expect(screen.getByText("bob@example.com")).toBeInTheDocument();
    });

    it("shows loading message while fetching", () => {
      mockInvitations.mockReturnValue({
        invitations: [],
        totalCount: 0,
        isLoading: true,
      });
      renderTab();
      expect(screen.getByText(/loading invitations/i)).toBeInTheDocument();
    });

    it("shows empty state when there are no pending invitations", () => {
      mockInvitations.mockReturnValue({
        invitations: [],
        totalCount: 0,
        isLoading: false,
      });
      renderTab();
      expect(screen.getByText(/no pending invitations/i)).toBeInTheDocument();
    });

    it("shows expired badge for pending invitation past expires_at", () => {
      mockInvitations.mockReturnValue({
        invitations: [makeExpiredInvitation()],
        totalCount: 1,
        isLoading: false,
      });
      renderTab();
      expect(screen.getByText("Expired")).toBeInTheDocument();
    });
  });

  describe("status filter", () => {
    it("changes the displayed status when a new option is selected", async () => {
      const user = userEvent.setup();
      renderTab();

      const select = screen.getByRole("combobox", {
        name: /filter invitations by status/i,
      });
      await user.selectOptions(select, "accepted");

      expect(select).toHaveValue("accepted");
    });
  });

  describe("action buttons — enablement rules", () => {
    it("Edit button is enabled for pending invitations", () => {
      mockInvitations.mockReturnValue({
        invitations: [makeInvitation({ status: "pending" })],
        totalCount: 1,
        isLoading: false,
      });
      renderTab();
      expect(
        screen.getByRole("button", { name: /edit invitation role/i }),
      ).not.toBeDisabled();
    });

    it("Edit button is disabled for accepted invitations", () => {
      mockInvitations.mockReturnValue({
        invitations: [makeInvitation({ status: "accepted" })],
        totalCount: 1,
        isLoading: false,
      });
      renderTab();
      expect(
        screen.getByRole("button", { name: /edit invitation role/i }),
      ).toBeDisabled();
    });

    it("Cancel button is enabled for pending invitations", () => {
      mockInvitations.mockReturnValue({
        invitations: [makeInvitation({ status: "pending" })],
        totalCount: 1,
        isLoading: false,
      });
      renderTab();
      expect(
        screen.getByRole("button", { name: /cancel invitation/i }),
      ).not.toBeDisabled();
    });

    it("Cancel button is disabled for cancelled invitations", () => {
      mockInvitations.mockReturnValue({
        invitations: [makeInvitation({ status: "cancelled" })],
        totalCount: 1,
        isLoading: false,
      });
      renderTab();
      expect(
        screen.getByRole("button", { name: /cancel invitation/i }),
      ).toBeDisabled();
    });

    it("Resend button is enabled for cancelled invitations", () => {
      mockInvitations.mockReturnValue({
        invitations: [makeInvitation({ status: "cancelled" })],
        totalCount: 1,
        isLoading: false,
      });
      renderTab();
      expect(
        screen.getByRole("button", { name: /resend invitation/i }),
      ).not.toBeDisabled();
    });

    it("Resend button is enabled for expired pending invitations", () => {
      mockInvitations.mockReturnValue({
        invitations: [makeExpiredInvitation({ status: "pending" })],
        totalCount: 1,
        isLoading: false,
      });
      renderTab();
      expect(
        screen.getByRole("button", { name: /resend invitation/i }),
      ).not.toBeDisabled();
    });

    it("Resend button is disabled for non-expired pending invitations", () => {
      mockInvitations.mockReturnValue({
        invitations: [makeInvitation({ status: "pending" })],
        totalCount: 1,
        isLoading: false,
      });
      renderTab();
      expect(
        screen.getByRole("button", { name: /resend invitation/i }),
      ).toBeDisabled();
    });

    it("Resend button is disabled for accepted invitations", () => {
      mockInvitations.mockReturnValue({
        invitations: [makeInvitation({ status: "accepted" })],
        totalCount: 1,
        isLoading: false,
      });
      renderTab();
      expect(
        screen.getByRole("button", { name: /resend invitation/i }),
      ).toBeDisabled();
    });
  });

  describe("Invite Member drawer", () => {
    it("opens InvitationDrawer when Invite Member is clicked", async () => {
      const user = userEvent.setup();
      renderTab();

      await user.click(screen.getByRole("button", { name: /invite member/i }));

      expect(screen.getByTestId("invitation-drawer")).toBeInTheDocument();
    });

    it("closes InvitationDrawer when the drawer's close action is triggered", async () => {
      const user = userEvent.setup();
      renderTab();

      await user.click(screen.getByRole("button", { name: /invite member/i }));
      await user.click(
        screen.getByRole("button", { name: /close invite drawer/i }),
      );

      expect(screen.queryByTestId("invitation-drawer")).not.toBeInTheDocument();
    });
  });

  describe("Edit invitation drawer", () => {
    it("opens EditInvitationDrawer when Edit button is clicked on a pending invitation", async () => {
      const user = userEvent.setup();
      mockInvitations.mockReturnValue({
        invitations: [makeInvitation({ status: "pending" })],
        totalCount: 1,
        isLoading: false,
      });
      renderTab();

      await user.click(
        screen.getByRole("button", { name: /edit invitation role/i }),
      );

      expect(screen.getByTestId("edit-invitation-drawer")).toBeInTheDocument();
    });
  });

  describe("Cancel invitation", () => {
    it("opens confirmation dialog when Cancel button is clicked on a pending invitation", async () => {
      const user = userEvent.setup();
      mockInvitations.mockReturnValue({
        invitations: [makeInvitation({ status: "pending" })],
        totalCount: 1,
        isLoading: false,
      });
      renderTab();

      await user.click(
        screen.getByRole("button", { name: /cancel invitation/i }),
      );

      expect(
        screen.getByRole("heading", { name: /cancel invitation/i }),
      ).toBeInTheDocument();
    });

    it("calls cancelMembershipInvitation mutation when confirmed", async () => {
      const user = userEvent.setup();
      const inv = makeInvitation({
        status: "pending",
        user: { id: "u1", email: "alice@example.com" },
      });
      mockInvitations.mockReturnValue({
        invitations: [inv],
        totalCount: 1,
        isLoading: false,
      });
      renderTab("t1");

      // Open dialog by clicking the table action button (aria-label)
      await user.click(
        screen.getByRole("button", { name: "Cancel invitation" }),
      );

      // Confirm inside the dialog — the confirm button renders with the confirmLabel prop
      const dialog = screen.getByRole("dialog");
      await user.click(
        within(dialog).getByRole("button", { name: /cancel invitation/i }),
      );

      await waitFor(() =>
        expect(mockCancelMutateAsync).toHaveBeenCalledWith({
          path: { tenant: "t1", "user-id": "u1" },
        }),
      );
    });
  });

  describe("Resend invitation", () => {
    it("opens confirmation dialog when Resend is clicked on a cancelled invitation", async () => {
      const user = userEvent.setup();
      mockInvitations.mockReturnValue({
        invitations: [makeInvitation({ status: "cancelled" })],
        totalCount: 1,
        isLoading: false,
      });
      renderTab();

      await user.click(
        screen.getByRole("button", { name: /resend invitation/i }),
      );

      expect(
        screen.getByRole("heading", { name: /resend invitation/i }),
      ).toBeInTheDocument();
    });

    it("calls generateInvitationLink mutation when confirmed", async () => {
      const user = userEvent.setup();
      const inv = makeInvitation({
        status: "cancelled",
        user: { id: "u1", email: "alice@example.com" },
        role: "operator",
      });
      mockInvitations.mockReturnValue({
        invitations: [inv],
        totalCount: 1,
        isLoading: false,
      });
      renderTab("t1");

      await user.click(
        screen.getByRole("button", { name: /resend invitation/i }),
      );
      await user.click(screen.getByRole("button", { name: /^resend$/i }));

      await waitFor(() =>
        expect(mockResendMutateAsync).toHaveBeenCalledWith({
          path: { tenant: "t1" },
          body: { email: "alice@example.com", role: "operator" },
        }),
      );
    });
  });

  describe("permission gating", () => {
    it("restricts Invite Member button for non-owner roles", () => {
      useAuthStore.setState({ role: "observer" });
      renderTab();
      // RestrictedAction wraps it in aria-disabled span
      const wrapper = screen
        .getByRole("button", { name: /invite member/i })
        .closest("[aria-disabled='true']");
      expect(wrapper).toBeInTheDocument();
    });
  });
});
