import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  render,
  screen,
  cleanup,
  waitFor,
  within,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useAuthStore } from "../../stores/authStore";
import AcceptInvite from "../AcceptInvite";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

// Stub ConfirmDialog — jsdom lacks HTMLDialogElement.showModal()
vi.mock("../../components/common/ConfirmDialog", () => ({
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

const mockLookupUserStatus = vi.fn();

vi.mock("../../client", () => ({
  lookupUserStatus: (...args: unknown[]): unknown =>
    mockLookupUserStatus(...args),
}));

const mockAcceptMutateAsync = vi.fn();
const mockDeclineMutateAsync = vi.fn();
const mockSwitchNamespaceMutateAsync = vi.fn();

vi.mock("../../hooks/useInvitationMutations", () => ({
  useAcceptInvite: () => ({
    mutateAsync: mockAcceptMutateAsync,
    isPending: false,
  }),
  useDeclineInvite: () => ({
    mutateAsync: mockDeclineMutateAsync,
    isPending: false,
  }),
}));

vi.mock("../../hooks/useNamespaceMutations", () => ({
  useSwitchNamespace: () => ({
    mutateAsync: mockSwitchNamespaceMutateAsync,
    isPending: false,
  }),
}));

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

const VALID_PARAMS =
  "tenant-id=t1&user-id=u1&sig=abc123&email=alice@example.com";

function createWrapper(initialSearch = "") {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <MemoryRouter
      initialEntries={[
        `/accept-invite${initialSearch ? "?" + initialSearch : ""}`,
      ]}
    >
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </MemoryRouter>
  );
}

function renderPage(search = VALID_PARAMS) {
  return render(<AcceptInvite />, {
    wrapper: createWrapper(search),
  });
}

/* ------------------------------------------------------------------ */
/* Setup / teardown                                                    */
/* ------------------------------------------------------------------ */

afterEach(cleanup);

beforeEach(() => {
  vi.clearAllMocks();
  // Default: unauthenticated
  useAuthStore.setState({
    token: null,
    user: null,
    userId: null,
    email: null,
    tenant: null,
    role: null,
    name: null,
    loading: false,
  });
  mockAcceptMutateAsync.mockResolvedValue(undefined);
  mockDeclineMutateAsync.mockResolvedValue(undefined);
  mockSwitchNamespaceMutateAsync.mockResolvedValue(undefined);
});

/* ================================================================== */
/* Tests                                                               */
/* ================================================================== */

describe("AcceptInvite", () => {
  describe("branch: missing-params", () => {
    it("renders the Invalid Invitation heading when query params are missing", async () => {
      renderPage("");
      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /invalid invitation/i }),
        ).toBeInTheDocument(),
      );
    });

    it("renders a Back to Login link", async () => {
      renderPage("");
      await waitFor(() =>
        expect(
          screen.getByRole("link", { name: /back to login/i }),
        ).toBeInTheDocument(),
      );
    });
  });

  describe("branch: ready (authenticated as the correct user)", () => {
    beforeEach(() => {
      useAuthStore.setState({
        token: "jwt-token",
        userId: "u1",
        email: "alice@example.com",
        user: "alice",
        tenant: "t1",
        role: "owner",
        name: "Alice",
        loading: false,
      });
    });

    it("renders the Namespace Invitation heading", async () => {
      renderPage(VALID_PARAMS);
      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /namespace invitation/i }),
        ).toBeInTheDocument(),
      );
    });

    it("renders Accept and Decline buttons", async () => {
      renderPage(VALID_PARAMS);
      await waitFor(() =>
        expect(
          screen.getByRole("button", { name: /accept/i }),
        ).toBeInTheDocument(),
      );
      expect(
        screen.getByRole("button", { name: /decline/i }),
      ).toBeInTheDocument();
    });

    it("does not show a loading spinner once ready", async () => {
      renderPage(VALID_PARAMS);
      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /namespace invitation/i }),
        ).toBeInTheDocument(),
      );
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });

    describe("accept flow", () => {
      it("shows Accept confirmation dialog when Accept button is clicked", async () => {
        const user = userEvent.setup();
        renderPage(VALID_PARAMS);

        await waitFor(() =>
          expect(
            screen.getByRole("button", { name: /^accept$/i }),
          ).toBeInTheDocument(),
        );
        await user.click(screen.getByRole("button", { name: /^accept$/i }));

        expect(
          screen.getByRole("heading", { name: /accept invitation/i }),
        ).toBeInTheDocument();
      });

      it("calls acceptInvite mutation with the tenant from query params", async () => {
        const user = userEvent.setup();
        renderPage(VALID_PARAMS);

        await waitFor(() =>
          expect(
            screen.getByRole("button", { name: /^accept$/i }),
          ).toBeInTheDocument(),
        );
        // Open dialog
        await user.click(screen.getByRole("button", { name: /^accept$/i }));
        // Confirm inside dialog
        const dialog = screen.getByRole("dialog");
        await user.click(
          within(dialog).getByRole("button", { name: /^accept$/i }),
        );

        await waitFor(() =>
          expect(mockAcceptMutateAsync).toHaveBeenCalledWith({
            path: { tenant: "t1" },
          }),
        );
      });

      it("switches namespace via useSwitchNamespace after acceptInvite succeeds", async () => {
        const user = userEvent.setup();
        renderPage(VALID_PARAMS);

        await waitFor(() =>
          expect(
            screen.getByRole("button", { name: /^accept$/i }),
          ).toBeInTheDocument(),
        );
        await user.click(screen.getByRole("button", { name: /^accept$/i }));
        const dialog = screen.getByRole("dialog");
        await user.click(
          within(dialog).getByRole("button", { name: /^accept$/i }),
        );

        // acceptInvite runs first, then switchNamespace mints a fresh
        // namespace-scoped token and hard-navigates to /dashboard.
        await waitFor(() =>
          expect(mockAcceptMutateAsync).toHaveBeenCalledWith({
            path: { tenant: "t1" },
          }),
        );
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
        renderPage(VALID_PARAMS);

        await waitFor(() =>
          expect(
            screen.getByRole("button", { name: /^accept$/i }),
          ).toBeInTheDocument(),
        );
        await user.click(screen.getByRole("button", { name: /^accept$/i }));
        const dialog = screen.getByRole("dialog");
        await user.click(
          within(dialog).getByRole("button", { name: /^accept$/i }),
        );

        await waitFor(() => expect(mockAcceptMutateAsync).toHaveBeenCalled());
        expect(mockSwitchNamespaceMutateAsync).not.toHaveBeenCalled();
      });

      it("shows action error when acceptInvite mutation fails", async () => {
        mockAcceptMutateAsync.mockRejectedValue(new Error("server error"));
        const user = userEvent.setup();
        renderPage(VALID_PARAMS);

        await waitFor(() =>
          expect(
            screen.getByRole("button", { name: /^accept$/i }),
          ).toBeInTheDocument(),
        );
        await user.click(screen.getByRole("button", { name: /^accept$/i }));
        const dialog = screen.getByRole("dialog");
        await user.click(
          within(dialog).getByRole("button", { name: /^accept$/i }),
        );

        await waitFor(() =>
          expect(
            screen.getByText(/failed to accept the invitation/i),
          ).toBeInTheDocument(),
        );
      });
    });

    describe("decline flow", () => {
      it("shows Decline confirmation dialog when Decline button is clicked", async () => {
        const user = userEvent.setup();
        renderPage(VALID_PARAMS);

        await waitFor(() =>
          expect(
            screen.getByRole("button", { name: /decline/i }),
          ).toBeInTheDocument(),
        );
        await user.click(screen.getByRole("button", { name: /decline/i }));

        expect(
          screen.getByRole("heading", { name: /decline invitation/i }),
        ).toBeInTheDocument();
      });

      it("calls declineInvite mutation and navigates to /dashboard after confirm", async () => {
        const user = userEvent.setup();
        renderPage(VALID_PARAMS);

        await waitFor(() =>
          expect(
            screen.getByRole("button", { name: /decline/i }),
          ).toBeInTheDocument(),
        );
        await user.click(screen.getByRole("button", { name: /decline/i }));
        const dialog = screen.getByRole("dialog");
        await user.click(
          within(dialog).getByRole("button", { name: /^decline$/i }),
        );

        await waitFor(() =>
          expect(mockDeclineMutateAsync).toHaveBeenCalledWith({
            path: { tenant: "t1" },
          }),
        );
        await waitFor(() =>
          expect(mockNavigate).toHaveBeenCalledWith("/dashboard"),
        );
      });

      it("shows action error when declineInvite mutation fails", async () => {
        mockDeclineMutateAsync.mockRejectedValue(new Error("server error"));
        const user = userEvent.setup();
        renderPage(VALID_PARAMS);

        await waitFor(() =>
          expect(
            screen.getByRole("button", { name: /decline/i }),
          ).toBeInTheDocument(),
        );
        await user.click(screen.getByRole("button", { name: /decline/i }));
        const dialog = screen.getByRole("dialog");
        await user.click(
          within(dialog).getByRole("button", { name: /^decline$/i }),
        );

        await waitFor(() =>
          expect(
            screen.getByText(/failed to decline the invitation/i),
          ).toBeInTheDocument(),
        );
      });
    });
  });

  describe("branch: wrong-user (authenticated as different user)", () => {
    beforeEach(() => {
      useAuthStore.setState({
        token: "jwt-token",
        userId: "other-user-id", // Mismatch with user-id=u1 in params
        email: "other@example.com",
        user: "other",
        tenant: "t2",
        role: "owner",
        name: "Other",
        loading: false,
      });
    });

    it("renders the Different Account Signed In heading", async () => {
      renderPage(VALID_PARAMS);
      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /different account signed in/i }),
        ).toBeInTheDocument(),
      );
    });

    it("renders a Sign Out button", async () => {
      renderPage(VALID_PARAMS);
      await waitFor(() =>
        expect(
          screen.getByRole("button", { name: /sign out/i }),
        ).toBeInTheDocument(),
      );
    });

    it("calls logout and navigates to login with redirect on Sign Out", async () => {
      const logout = vi.fn();
      useAuthStore.setState({ logout } as never);
      const user = userEvent.setup();
      renderPage(VALID_PARAMS);

      await waitFor(() =>
        expect(
          screen.getByRole("button", { name: /sign out/i }),
        ).toBeInTheDocument(),
      );
      await user.click(screen.getByRole("button", { name: /sign out/i }));

      expect(logout).toHaveBeenCalledTimes(1);
      await waitFor(() =>
        expect(mockNavigate).toHaveBeenCalledWith(
          expect.stringContaining("/login"),
        ),
      );
    });
  });

  describe("branch: unauthenticated — lookupUserStatus invited → navigate to /sign-up", () => {
    it("navigates to /sign-up with email and sig when status is 'invited'", async () => {
      mockLookupUserStatus.mockResolvedValue({
        data: { status: "invited" },
        response: new Response(),
      });
      renderPage(VALID_PARAMS);

      await waitFor(() =>
        expect(mockNavigate).toHaveBeenCalledWith(
          expect.stringMatching(/^\/sign-up\?/),
        ),
      );
      const call = mockNavigate.mock.calls[0][0] as string;
      expect(call).toContain("email=alice%40example.com");
      expect(call).toContain("sig=abc123");
    });
  });

  describe("branch: unauthenticated — lookupUserStatus confirmed → navigate to /login", () => {
    it("navigates to /login with redirect param when status is 'confirmed'", async () => {
      mockLookupUserStatus.mockResolvedValue({
        data: { status: "confirmed" },
        response: new Response(),
      });
      renderPage(VALID_PARAMS);

      await waitFor(() =>
        expect(mockNavigate).toHaveBeenCalledWith(
          expect.stringMatching(/^\/login\?/),
        ),
      );
      const call = mockNavigate.mock.calls[0][0] as string;
      expect(call).toContain("redirect=");
      expect(decodeURIComponent(call)).toContain("accept-invite");
    });
  });

  describe("branch: unauthenticated — lookupUserStatus not-confirmed → navigate to /login", () => {
    it("navigates to /login when status is 'not-confirmed'", async () => {
      mockLookupUserStatus.mockResolvedValue({
        data: { status: "not-confirmed" },
        response: new Response(),
      });
      renderPage(VALID_PARAMS);

      await waitFor(() =>
        expect(mockNavigate).toHaveBeenCalledWith(
          expect.stringMatching(/^\/login\?redirect=/),
        ),
      );
    });
  });

  describe("branch: error (lookupUserStatus rejects)", () => {
    it("renders the Invitation Unavailable heading when lookupUserStatus throws", async () => {
      mockLookupUserStatus.mockRejectedValue(new Error("network failure"));
      renderPage(VALID_PARAMS);

      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /invitation unavailable/i }),
        ).toBeInTheDocument(),
      );
    });

    it("renders a Back to Login link in the error state", async () => {
      mockLookupUserStatus.mockRejectedValue(new Error("network failure"));
      renderPage(VALID_PARAMS);

      await waitFor(() =>
        expect(
          screen.getByRole("link", { name: /back to login/i }),
        ).toBeInTheDocument(),
      );
    });
  });

  describe("initial loading state", () => {
    it("shows the checking invitation spinner initially when params are present and unauthenticated", async () => {
      // Delay resolution so we can observe the loading state
      mockLookupUserStatus.mockReturnValue(new Promise(() => {}));
      renderPage(VALID_PARAMS);

      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText(/checking invitation/i)).toBeInTheDocument();
    });
  });
});
