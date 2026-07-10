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
import { useAuthStore } from "@/stores/authStore";
import AcceptInvite from "../AcceptInvite";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

// Stub ConfirmDialog — jsdom lacks HTMLDialogElement.showModal()
vi.mock("@/components/common/ConfirmDialog", () => ({
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
        <button type="button" onClick={onClose}>
          {cancelLabel}
        </button>
        <button type="button" onClick={() => void onConfirm()}>
          {confirmLabel}
        </button>
      </div>
    );
  },
}));

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

const mockResolveInvitation = vi.fn();

vi.mock("@/client", () => ({
  resolveInvitation: (...args: unknown[]): unknown =>
    mockResolveInvitation(...args),
}));

const { mockSignUp, signUpState } = vi.hoisted(() => {
  const mockSignUp = vi.fn();
  const signUpState = {
    signUp: mockSignUp,
    signUpLoading: false,
    signUpError: null as string | null,
    signUpServerFields: [] as string[],
  };
  return { mockSignUp, signUpState };
});

vi.mock("@/stores/signUpStore", () => ({
  useSignUpStore: Object.assign(
    (selector: (s: typeof signUpState) => unknown) => selector(signUpState),
    { getState: () => signUpState },
  ),
}));

const mockAcceptMutateAsync = vi.fn();
const mockSwitchNamespaceMutateAsync = vi.fn();

vi.mock("@/hooks/useInvitationMutations", () => ({
  useAcceptInvite: () => ({
    mutateAsync: mockAcceptMutateAsync,
    isPending: false,
  }),
}));

vi.mock("@/hooks/useNamespaceMutations", () => ({
  useSwitchNamespace: () => ({
    mutateAsync: mockSwitchNamespaceMutateAsync,
    isPending: false,
  }),
}));

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

const INVITE_CODE = "INVITECODE12";
const VALID_PARAMS = `invite=${INVITE_CODE}`;

function resolved(status: string) {
  return {
    data: {
      tenant_id: "t1",
      user_id: "u1",
      email: "alice@example.com",
      status,
    },
    response: new Response(),
  };
}

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
  return render(<AcceptInvite />, { wrapper: createWrapper(search) });
}

/* ------------------------------------------------------------------ */
/* Setup / teardown                                                    */
/* ------------------------------------------------------------------ */

afterEach(cleanup);

beforeEach(() => {
  vi.clearAllMocks();
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
  signUpState.signUpLoading = false;
  signUpState.signUpError = null;
  signUpState.signUpServerFields = [];
  mockAcceptMutateAsync.mockResolvedValue(undefined);
  mockSwitchNamespaceMutateAsync.mockResolvedValue(undefined);
  mockResolveInvitation.mockResolvedValue(resolved("confirmed"));
});

/* ================================================================== */
/* Tests                                                               */
/* ================================================================== */

describe("AcceptInvite", () => {
  describe("branch: missing-params", () => {
    it("renders the Invalid Invitation heading when the invite code is missing", async () => {
      renderPage("");
      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /invalid invitation/i }),
        ).toBeInTheDocument(),
      );
    });
  });

  describe("initial loading state", () => {
    it("shows the checking invitation spinner while resolving", async () => {
      mockResolveInvitation.mockReturnValue(new Promise(() => {}));
      renderPage(VALID_PARAMS);
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText(/checking invitation/i)).toBeInTheDocument();
    });
  });

  describe("branch: error (resolve rejects)", () => {
    it("renders the Invitation Unavailable heading", async () => {
      mockResolveInvitation.mockRejectedValue(new Error("network failure"));
      renderPage(VALID_PARAMS);
      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /invitation unavailable/i }),
        ).toBeInTheDocument(),
      );
    });
  });

  describe("branch: ready (authenticated as the invited user)", () => {
    beforeEach(() => {
      useAuthStore.setState({
        token: "jwt-token",
        userId: "u1",
        email: "alice@example.com",
        loading: false,
      } as never);
    });

    it("renders the Namespace Invitation heading with Accept", async () => {
      renderPage(VALID_PARAMS);
      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /namespace invitation/i }),
        ).toBeInTheDocument(),
      );
      expect(
        screen.getByRole("button", { name: /accept/i }),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /decline/i }),
      ).not.toBeInTheDocument();
    });

    it("accepts and switches namespace using the resolved tenant", async () => {
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
        expect(mockAcceptMutateAsync).toHaveBeenCalledWith({
          path: { tenant: "t1" },
        }),
      );

      // Accepting lands on the "You're in" confirmation; entering is a deliberate click.
      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /you're in/i }),
        ).toBeInTheDocument(),
      );
      expect(mockSwitchNamespaceMutateAsync).not.toHaveBeenCalled();

      await user.click(
        screen.getByRole("button", { name: /go to dashboard/i }),
      );
      await waitFor(() =>
        expect(mockSwitchNamespaceMutateAsync).toHaveBeenCalledWith({
          tenantId: "t1",
          redirectTo: "/dashboard",
        }),
      );
    });
  });

  describe("branch: wrong-user (authenticated as a different user)", () => {
    it("renders the Different Account Signed In heading", async () => {
      useAuthStore.setState({
        token: "jwt-token",
        userId: "other-user-id",
        email: "other@example.com",
        loading: false,
      } as never);
      renderPage(VALID_PARAMS);
      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /different account signed in/i }),
        ).toBeInTheDocument(),
      );
    });
  });

  describe("branch: complete (unauthenticated, status invited)", () => {
    beforeEach(() => {
      mockResolveInvitation.mockResolvedValue(resolved("invited"));
    });

    async function fillForm(user: ReturnType<typeof userEvent.setup>) {
      await user.type(screen.getByLabelText(/^name$/i), "Alice");
      await user.type(screen.getByLabelText(/^username$/i), "alice");
      await user.type(screen.getByLabelText(/^password$/i), "Secret123");
      await user.type(screen.getByLabelText(/confirm password/i), "Secret123");
    }

    it("renders the invite completion form with the resolved email", async () => {
      renderPage(VALID_PARAMS);
      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /you've been invited/i }),
        ).toBeInTheDocument(),
      );
      expect(screen.getByText(/alice@example.com/)).toBeInTheDocument();
      expect(screen.getByLabelText(/^name$/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/^username$/i)).toBeInTheDocument();
      // Email is NOT an editable field
      expect(screen.queryByLabelText(/^email$/i)).not.toBeInTheDocument();
    });

    it("submits with the invite code as sig, no marketing, and switches namespace on token", async () => {
      mockSignUp.mockResolvedValue("tok");
      const user = userEvent.setup();
      renderPage(VALID_PARAMS);
      await waitFor(() =>
        expect(
          screen.getByRole("button", { name: /join namespace/i }),
        ).toBeInTheDocument(),
      );
      await fillForm(user);
      await user.click(screen.getByRole("button", { name: /join namespace/i }));

      await waitFor(() => expect(mockSignUp).toHaveBeenCalledTimes(1));
      expect(mockSignUp).toHaveBeenCalledWith(
        expect.objectContaining({
          email: "alice@example.com",
          sig: INVITE_CODE,
          email_marketing: false,
        }),
      );

      // Completing with a token lands on the "You're in" confirmation, not straight in.
      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /you're in/i }),
        ).toBeInTheDocument(),
      );
      await user.click(
        screen.getByRole("button", { name: /go to dashboard/i }),
      );
      await waitFor(() =>
        expect(mockSwitchNamespaceMutateAsync).toHaveBeenCalledWith({
          tenantId: "t1",
          redirectTo: "/dashboard",
        }),
      );
      // The completion token must establish the session, otherwise switchNamespace fires
      // unauthenticated and bounces the invitee to /login.
      expect(useAuthStore.getState().token).toBe("tok");
    });

    it("shows the Waiting for Approval screen when no token is returned", async () => {
      mockSignUp.mockResolvedValue(null);
      const user = userEvent.setup();
      renderPage(VALID_PARAMS);
      await waitFor(() =>
        expect(
          screen.getByRole("button", { name: /join namespace/i }),
        ).toBeInTheDocument(),
      );
      await fillForm(user);
      await user.click(screen.getByRole("button", { name: /join namespace/i }));

      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /waiting for approval/i }),
        ).toBeInTheDocument(),
      );
      expect(mockSwitchNamespaceMutateAsync).not.toHaveBeenCalled();
    });
  });

  describe("branch: unauthenticated with an existing account → login", () => {
    it("navigates to /login with a redirect back to accept-invite when confirmed", async () => {
      mockResolveInvitation.mockResolvedValue(resolved("confirmed"));
      renderPage(VALID_PARAMS);
      await waitFor(() =>
        expect(mockNavigate).toHaveBeenCalledWith(
          expect.stringMatching(/^\/login\?redirect=/),
        ),
      );
      const call = mockNavigate.mock.calls[0][0] as string;
      expect(decodeURIComponent(call)).toContain("accept-invite");
    });

    it("navigates to /login when not-confirmed", async () => {
      mockResolveInvitation.mockResolvedValue(resolved("not-confirmed"));
      renderPage(VALID_PARAMS);
      await waitFor(() =>
        expect(mockNavigate).toHaveBeenCalledWith(
          expect.stringMatching(/^\/login\?redirect=/),
        ),
      );
    });
  });
});
