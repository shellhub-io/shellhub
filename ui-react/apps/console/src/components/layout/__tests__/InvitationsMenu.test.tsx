import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  render,
  screen,
  cleanup,
  waitFor,
  fireEvent,
  within,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useAuthStore } from "../../../stores/authStore";
import type { MembershipInvitation } from "../../../client";
import InvitationsMenu from "../InvitationsMenu";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

// Stub ConfirmDialog — jsdom lacks HTMLDialogElement.showModal()
vi.mock("../../../components/common/ConfirmDialog", () => ({
  default: ({
    open,
    onClose,
    onConfirm,
    title,
    confirmLabel = "Confirm",
    cancelLabel = "Cancel",
    errorMessage,
  }: {
    open: boolean;
    onClose: () => void;
    onConfirm: () => Promise<void> | void;
    title: string;
    confirmLabel?: string;
    cancelLabel?: string;
    errorMessage?: string | null;
  }) => {
    if (!open) return null;
    return (
      <div role="dialog">
        <h2>{title}</h2>
        {errorMessage ? <div role="alert">{errorMessage}</div> : null}
        <button onClick={onClose}>{cancelLabel}</button>
        <button onClick={() => void onConfirm()}>{confirmLabel}</button>
      </div>
    );
  },
}));

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

// Cloud mode on by default — tests that need it off override in their own beforeEach
vi.mock("../../../env", () => ({
  getConfig: vi.fn(() => ({ cloud: true })),
}));

const mockUserInvitations = vi.fn<
  () => {
    invitations: MembershipInvitation[];
    totalCount: number;
    isLoading: boolean;
  }
>();

vi.mock("../../../hooks/useInvitations", () => ({
  useUserInvitations: () => mockUserInvitations(),
}));

const mockAcceptMutateAsync = vi.fn();
const mockDeclineMutateAsync = vi.fn();
const mockSwitchNamespaceMutateAsync = vi.fn();

vi.mock("../../../hooks/useInvitationMutations", () => ({
  useAcceptInvite: () => ({
    mutateAsync: mockAcceptMutateAsync,
    isPending: false,
  }),
  useDeclineInvite: () => ({
    mutateAsync: mockDeclineMutateAsync,
    isPending: false,
  }),
}));

vi.mock("../../../hooks/useNamespaceMutations", () => ({
  useSwitchNamespace: () => ({
    mutateAsync: mockSwitchNamespaceMutateAsync,
    isPending: false,
  }),
}));

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

function makeInvitation(
  overrides: Partial<MembershipInvitation> = {},
): MembershipInvitation {
  return {
    namespace: { tenant_id: "t1", name: "my-namespace" },
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

function renderMenu() {
  return render(<InvitationsMenu />, {
    wrapper: createWrapper(),
  });
}

async function openMenu() {
  await userEvent.click(
    screen.getByRole("button", { name: /pending invitations/i }),
  );
}

/* ------------------------------------------------------------------ */
/* Setup / teardown                                                    */
/* ------------------------------------------------------------------ */

afterEach(cleanup);

beforeEach(() => {
  vi.clearAllMocks();
  useAuthStore.setState({
    token: "test-token",
    user: "alice",
    userId: "user-1",
    email: "alice@example.com",
    tenant: "t1",
    role: "owner",
    name: "Alice",
    loading: false,
  });
  mockUserInvitations.mockReturnValue({
    invitations: [],
    totalCount: 0,
    isLoading: false,
  });
  mockAcceptMutateAsync.mockResolvedValue(undefined);
  mockDeclineMutateAsync.mockResolvedValue(undefined);
  mockSwitchNamespaceMutateAsync.mockResolvedValue(undefined);
});

/* ================================================================== */
/* Tests                                                               */
/* ================================================================== */

describe("InvitationsMenu", () => {
  describe("cloud gating", () => {
    it("returns null in non-cloud mode", async () => {
      const { getConfig } = await import("../../../env");
      vi.mocked(getConfig).mockReturnValue({
        cloud: false,
        version: "",
        enterprise: false,
        announcements: false,
        webEndpoints: false,
        onboardingUrl: "",
        stripePublishableKey: "",
          chatwootWebsiteToken: "",
          chatwootBaseUrl: "",
      });

      const { container } = renderMenu();
      expect(container).toBeEmptyDOMElement();

      // Restore cloud mode for subsequent tests
      vi.mocked(getConfig).mockReturnValue({
        cloud: true,
        version: "",
        enterprise: false,
        announcements: false,
        webEndpoints: false,
        onboardingUrl: "",
        stripePublishableKey: "",
          chatwootWebsiteToken: "",
          chatwootBaseUrl: "",
      });
    });

    it("returns null when there is no auth token", () => {
      useAuthStore.setState({ token: null });
      const { container } = renderMenu();
      expect(container).toBeEmptyDOMElement();
    });
  });

  describe("trigger button", () => {
    it("renders the envelope button in cloud+authenticated mode", () => {
      renderMenu();
      expect(
        screen.getByRole("button", { name: /pending invitations/i }),
      ).toBeInTheDocument();
    });

    it("shows no badge when there are zero pending invitations", () => {
      renderMenu();
      // aria-hidden badge is absent
      expect(screen.queryByText(/[0-9]+/)).not.toBeInTheDocument();
    });

    it("shows badge count when there are pending invitations", () => {
      mockUserInvitations.mockReturnValue({
        invitations: [
          makeInvitation(),
          makeInvitation({ user: { id: "u2", email: "bob@example.com" } }),
        ],
        totalCount: 2,
        isLoading: false,
      });
      renderMenu();
      expect(screen.getByText("2")).toBeInTheDocument();
    });

    it("shows '9+' badge when there are more than 9 pending invitations", () => {
      mockUserInvitations.mockReturnValue({
        invitations: Array.from({ length: 10 }, (_, i) =>
          makeInvitation({
            user: { id: `u${i}`, email: `user${i}@example.com` },
          }),
        ),
        totalCount: 10,
        isLoading: false,
      });
      renderMenu();
      expect(screen.getByText("9+")).toBeInTheDocument();
    });

    it("includes count in aria-label when invitations are present", () => {
      mockUserInvitations.mockReturnValue({
        invitations: [makeInvitation()],
        totalCount: 1,
        isLoading: false,
      });
      renderMenu();
      expect(
        screen.getByRole("button", { name: /pending invitations \(1\)/i }),
      ).toBeInTheDocument();
    });
  });

  describe("dropdown open/close", () => {
    it("shows the dropdown menu after clicking the trigger", async () => {
      renderMenu();
      await openMenu();
      expect(screen.getByRole("menu")).toBeInTheDocument();
    });

    it("hides the dropdown after a second click on the trigger", async () => {
      const user = userEvent.setup();
      renderMenu();
      await user.click(
        screen.getByRole("button", { name: /pending invitations/i }),
      );
      await user.click(
        screen.getByRole("button", { name: /pending invitations/i }),
      );
      expect(screen.queryByRole("menu")).not.toBeInTheDocument();
    });

    it("dismisses on Escape key", async () => {
      renderMenu();
      await openMenu();
      // The escape handler is on the container element — use fireEvent
      fireEvent.keyDown(document, { key: "Escape" });
      await waitFor(() =>
        expect(screen.queryByRole("menu")).not.toBeInTheDocument(),
      );
    });

    it("dismisses on click outside", async () => {
      renderMenu();
      await openMenu();
      // Click outside the container
      fireEvent.mouseDown(document.body);
      await waitFor(() =>
        expect(screen.queryByRole("menu")).not.toBeInTheDocument(),
      );
    });
  });

  describe("dropdown content", () => {
    it("shows loading state when invitations are loading", async () => {
      mockUserInvitations.mockReturnValue({
        invitations: [],
        totalCount: 0,
        isLoading: true,
      });
      renderMenu();
      await openMenu();
      expect(screen.getByRole("status")).toBeInTheDocument();
    });

    it("shows empty state when there are no pending invitations", async () => {
      renderMenu();
      await openMenu();
      expect(screen.getByText(/you're all caught up/i)).toBeInTheDocument();
    });

    it("renders invitation cards when invitations exist", async () => {
      mockUserInvitations.mockReturnValue({
        invitations: [makeInvitation()],
        totalCount: 1,
        isLoading: false,
      });
      renderMenu();
      await openMenu();
      expect(screen.getByText("my-namespace")).toBeInTheDocument();
    });

    it("shows Accept and Decline buttons for each invitation card", async () => {
      mockUserInvitations.mockReturnValue({
        invitations: [makeInvitation()],
        totalCount: 1,
        isLoading: false,
      });
      renderMenu();
      await openMenu();
      expect(
        screen.getByRole("button", { name: /accept/i }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: /decline/i }),
      ).toBeInTheDocument();
    });

    it("disables Accept button for expired invitations", async () => {
      mockUserInvitations.mockReturnValue({
        invitations: [makeInvitation({ expires_at: "2020-01-01T00:00:00Z" })],
        totalCount: 1,
        isLoading: false,
      });
      renderMenu();
      await openMenu();
      expect(screen.getByRole("button", { name: /accept/i })).toBeDisabled();
    });

    it("shows Expired text for expired invitations", async () => {
      mockUserInvitations.mockReturnValue({
        invitations: [makeInvitation({ expires_at: "2020-01-01T00:00:00Z" })],
        totalCount: 1,
        isLoading: false,
      });
      renderMenu();
      await openMenu();
      expect(screen.getByText("Expired")).toBeInTheDocument();
    });
  });

  describe("accept flow", () => {
    it("shows Accept confirmation dialog after clicking Accept", async () => {
      const user = userEvent.setup();
      mockUserInvitations.mockReturnValue({
        invitations: [makeInvitation()],
        totalCount: 1,
        isLoading: false,
      });
      renderMenu();
      await openMenu();

      await user.click(screen.getByRole("button", { name: /^accept$/i }));

      expect(
        screen.getByRole("heading", { name: /accept invitation/i }),
      ).toBeInTheDocument();
    });

    it("calls acceptInvite and then useSwitchNamespace with the invitation's tenant", async () => {
      const user = userEvent.setup();
      mockUserInvitations.mockReturnValue({
        invitations: [
          makeInvitation({
            namespace: { tenant_id: "t1", name: "my-namespace" },
          }),
        ],
        totalCount: 1,
        isLoading: false,
      });
      renderMenu();
      await openMenu();

      await user.click(screen.getByRole("button", { name: /^accept$/i }));

      // Click the confirm button inside the dialog (not the card button)
      const dialog = screen.getByRole("dialog");
      await user.click(
        within(dialog).getByRole("button", { name: /^accept$/i }),
      );

      await waitFor(() =>
        expect(mockAcceptMutateAsync).toHaveBeenCalledWith({
          path: { tenant: "t1" },
        }),
      );
      // useSwitchNamespace is what actually switches the active namespace and
      // hard-navigates to /dashboard — the menu no longer calls setSession or
      // useNavigate itself for the accept flow.
      await waitFor(() =>
        expect(mockSwitchNamespaceMutateAsync).toHaveBeenCalledWith({
          tenantId: "t1",
          redirectTo: "/dashboard",
        }),
      );
    });

    it("does not switch namespace when acceptInvite fails", async () => {
      mockAcceptMutateAsync.mockRejectedValue(new Error("server error"));
      const user = userEvent.setup();
      mockUserInvitations.mockReturnValue({
        invitations: [
          makeInvitation({
            namespace: { tenant_id: "t1", name: "my-namespace" },
          }),
        ],
        totalCount: 1,
        isLoading: false,
      });
      renderMenu();
      await openMenu();

      await user.click(screen.getByRole("button", { name: /^accept$/i }));

      const dialog = screen.getByRole("dialog");
      await user.click(
        within(dialog).getByRole("button", { name: /^accept$/i }),
      );

      await waitFor(() => expect(mockAcceptMutateAsync).toHaveBeenCalled());
      expect(mockSwitchNamespaceMutateAsync).not.toHaveBeenCalled();
      // Error banner appears inline in the dialog.
      await waitFor(() =>
        expect(within(dialog).getByRole("alert")).toHaveTextContent(
          /failed to accept the invitation/i,
        ),
      );
    });
  });

  describe("decline flow", () => {
    it("shows Decline confirmation dialog after clicking Decline", async () => {
      const user = userEvent.setup();
      mockUserInvitations.mockReturnValue({
        invitations: [makeInvitation()],
        totalCount: 1,
        isLoading: false,
      });
      renderMenu();
      await openMenu();

      await user.click(screen.getByRole("button", { name: /decline/i }));

      expect(
        screen.getByRole("heading", { name: /decline invitation/i }),
      ).toBeInTheDocument();
    });

    it("calls declineInvite mutation on confirm", async () => {
      const user = userEvent.setup();
      mockUserInvitations.mockReturnValue({
        invitations: [
          makeInvitation({
            namespace: { tenant_id: "t1", name: "my-namespace" },
          }),
        ],
        totalCount: 1,
        isLoading: false,
      });
      renderMenu();
      await openMenu();

      // Click Decline on the invitation card
      await user.click(screen.getByRole("button", { name: /decline/i }));

      // Confirm inside the dialog
      const dialog = screen.getByRole("dialog");
      await user.click(
        within(dialog).getByRole("button", { name: /^decline$/i }),
      );

      await waitFor(() =>
        expect(mockDeclineMutateAsync).toHaveBeenCalledWith({
          path: { tenant: "t1" },
        }),
      );
    });
  });
});
