import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useAuthStore } from "@/stores/authStore";
import { simulateBrowserTranslation } from "@/tests/simulateBrowserTranslation";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import AcceptInvite from "../AcceptInvite";

vi.mock("@/components/common/ConfirmDialog", async () => ({
  default: (await import("@/tests/mocks")).MockConfirmDialog,
}));

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

const sdk = vi.hoisted(() =>
  mockSdkGen({
    resolveInvitation: vi.fn(),
    acceptInvite: vi.fn(),
    getNamespaceToken: vi.fn(),
  }),
);

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

const INVITE_CODE = "INVITECODE12";
const VALID_PARAMS = `invite=${INVITE_CODE}`;

const resolvedData = {
  tenant_id: "t1",
  user_id: "u1",
  email: "alice@example.com",
  status: "confirmed",
};

function renderPage(search = VALID_PARAMS) {
  return render(<AcceptInvite />, {
    wrapper: createTestWrapper({
      initialEntries: [`/accept-invite${search ? "?" + search : ""}`],
    }),
  });
}

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
    loginWithToken: async (token: string) => {
      useAuthStore.setState({
        token,
        user: "alice",
        userId: "u1",
        email: "alice@example.com",
        name: "Alice",
        loading: false,
      });
    },
  });
  signUpState.signUpLoading = false;
  signUpState.signUpError = null;
  signUpState.signUpServerFields = [];
  sdk.resolveInvitation.mockResolvedValue(mockSdkResponse(resolvedData));
  sdk.acceptInvite.mockResolvedValue(mockSdkResponse(undefined));
  sdk.getNamespaceToken.mockResolvedValue(
    mockSdkResponse({ token: "jwt-token", role: "owner" }),
  );
  mockNavigate.mockReset();
});

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
    it("shows the checking invitation spinner while resolving", () => {
      sdk.resolveInvitation.mockReturnValue(new Promise(() => {}));
      renderPage(VALID_PARAMS);
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText(/checking invitation/i)).toBeInTheDocument();
    });
  });

  describe("branch: error (resolve rejects)", () => {
    it("renders the Invitation Unavailable heading", async () => {
      sdk.resolveInvitation.mockRejectedValue({ status: 404 });
      renderPage(VALID_PARAMS);
      expect(
        await screen.findByRole("heading", { name: /invitation unavailable/i }),
      ).toBeInTheDocument();
    });
  });

  describe("branch: accept (authenticated as the invited user)", () => {
    beforeEach(() => {
      useAuthStore.setState({
        token: "jwt-token",
        userId: "u1",
        email: "alice@example.com",
        loading: false,
      });
    });

    it("renders the Namespace Invitation heading with Accept", async () => {
      renderPage(VALID_PARAMS);
      expect(
        await screen.findByRole("heading", { name: /namespace invitation/i }),
      ).toBeInTheDocument();
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
      await user.click(
        await screen.findByRole("button", { name: /^accept$/i }),
      );
      const dialog = screen.getByRole("dialog");
      await user.click(
        within(dialog).getByRole("button", { name: /^accept$/i }),
      );

      await waitFor(() =>
        expect(sdk.acceptInvite).toHaveBeenCalledWith(
          expect.objectContaining({
            path: { tenant: "t1" },
            throwOnError: true,
          }),
        ),
      );

      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /you're in/i }),
        ).toBeInTheDocument(),
      );
      expect(sdk.getNamespaceToken).not.toHaveBeenCalled();

      await user.click(
        screen.getByRole("button", { name: /go to dashboard/i }),
      );
      await waitFor(() =>
        expect(sdk.getNamespaceToken).toHaveBeenCalledWith(
          expect.objectContaining({
            path: { tenant: "t1" },
            throwOnError: true,
          }),
        ),
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
      });
      renderPage(VALID_PARAMS);
      expect(
        await screen.findByRole("heading", {
          name: /different account signed in/i,
        }),
      ).toBeInTheDocument();
    });
  });

  describe("branch: sign-up (unauthenticated, status invited)", () => {
    beforeEach(() => {
      sdk.resolveInvitation.mockResolvedValue(
        mockSdkResponse({ ...resolvedData, status: "invited" }),
      );
    });

    async function fillForm(user: ReturnType<typeof userEvent.setup>) {
      await user.type(screen.getByLabelText(/^name$/i), "Alice");
      await user.type(screen.getByLabelText(/^username$/i), "alice");
      await user.type(screen.getByLabelText(/^password$/i), "Secret123");
      await user.type(screen.getByLabelText(/confirm password/i), "Secret123");
    }

    it("renders the invite completion form with the resolved email", async () => {
      renderPage(VALID_PARAMS);
      expect(
        await screen.findByRole("heading", { name: /you've been invited/i }),
      ).toBeInTheDocument();
      expect(screen.getByText(/alice@example.com/)).toBeInTheDocument();
      expect(screen.getByLabelText(/^name$/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/^username$/i)).toBeInTheDocument();
      expect(screen.queryByLabelText(/^email$/i)).not.toBeInTheDocument();
    });

    it("links the form to the email hint via aria-describedby", async () => {
      renderPage(VALID_PARAMS);
      await screen.findByRole("heading", { name: /you've been invited/i });
      const form = screen.getByRole("form", {
        name: /complete your account/i,
      });
      const describedById = form.getAttribute("aria-describedby");
      expect(describedById).toBeTruthy();
      const hint = document.getElementById(describedById!);
      expect(hint).toHaveTextContent(/alice@example.com/);
    });

    it("submits with the invite code as sig, no marketing, and switches namespace on token", async () => {
      mockSignUp.mockResolvedValue("tok");
      const user = userEvent.setup();
      renderPage(VALID_PARAMS);
      await screen.findByRole("heading", { name: /you've been invited/i });
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

      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /you're in/i }),
        ).toBeInTheDocument(),
      );

      await user.click(
        screen.getByRole("button", { name: /go to dashboard/i }),
      );
      await waitFor(() =>
        expect(sdk.getNamespaceToken).toHaveBeenCalledWith(
          expect.objectContaining({
            path: { tenant: "t1" },
            throwOnError: true,
          }),
        ),
      );
    });

    it("announces the loading state to screen readers when switching namespace", async () => {
      mockSignUp.mockResolvedValue("tok");
      sdk.getNamespaceToken.mockReturnValue(new Promise(() => {}));
      const user = userEvent.setup();
      renderPage(VALID_PARAMS);
      await screen.findByRole("heading", { name: /you've been invited/i });
      await fillForm(user);
      await user.click(screen.getByRole("button", { name: /join namespace/i }));

      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /you're in/i }),
        ).toBeInTheDocument(),
      );

      await user.click(
        screen.getByRole("button", { name: /go to dashboard/i }),
      );

      await waitFor(() => {
        const statuses = screen.getAllByRole("status");
        expect(statuses.some((el) => el.textContent?.match(/switching/i))).toBe(
          true,
        );
      });
    });

    it("shows the Waiting for Approval screen when no token is returned", async () => {
      mockSignUp.mockResolvedValue(null);
      const user = userEvent.setup();
      renderPage(VALID_PARAMS);
      await screen.findByRole("heading", { name: /you've been invited/i });
      await fillForm(user);
      await user.click(screen.getByRole("button", { name: /join namespace/i }));

      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /waiting for approval/i }),
        ).toBeInTheDocument(),
      );
      expect(sdk.getNamespaceToken).not.toHaveBeenCalled();
    });
  });

  describe("branch: unauthenticated with an existing account → login", () => {
    it("navigates to /login with a redirect back to accept-invite when confirmed", async () => {
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
      sdk.resolveInvitation.mockResolvedValue(
        mockSdkResponse({ ...resolvedData, status: "not-confirmed" }),
      );
      renderPage(VALID_PARAMS);
      await waitFor(() =>
        expect(mockNavigate).toHaveBeenCalledWith(
          expect.stringMatching(/^\/login\?redirect=/),
        ),
      );
    });
  });

  describe("under a browser-translated DOM", () => {
    it("still reaches the joined confirmation after signing up", async () => {
      sdk.resolveInvitation.mockResolvedValue(
        mockSdkResponse({ ...resolvedData, status: "invited" }),
      );
      mockSignUp.mockResolvedValue("tok");
      const user = userEvent.setup();
      const { container } = renderPage(VALID_PARAMS);
      await screen.findByRole("heading", { name: /you've been invited/i });

      simulateBrowserTranslation(container);
      await user.type(screen.getByLabelText(/^name$/i), "Alice");
      await user.type(screen.getByLabelText(/^username$/i), "alice");
      await user.type(screen.getByLabelText(/^password$/i), "Secret123");
      await user.type(screen.getByLabelText(/confirm password/i), "Secret123");
      await user.click(screen.getByRole("button", { name: /join namespace/i }));

      expect(
        await screen.findByRole("heading", { name: /you're in/i }),
      ).toBeInTheDocument();
    });

    it("still reaches the joined confirmation after accepting", async () => {
      useAuthStore.setState({
        token: "jwt-token",
        userId: "u1",
        email: "alice@example.com",
        loading: false,
      });
      const user = userEvent.setup();
      const { container } = renderPage(VALID_PARAMS);
      await screen.findByRole("heading", { name: /namespace invitation/i });

      simulateBrowserTranslation(container);
      await user.click(screen.getByRole("button", { name: /^accept$/i }));
      const dialog = screen.getByRole("dialog");
      await user.click(
        within(dialog).getByRole("button", { name: /^accept$/i }),
      );

      expect(
        await screen.findByRole("heading", { name: /you're in/i }),
      ).toBeInTheDocument();
    });
  });
});
