import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { useAuthStore } from "@/stores/authStore";
import { simulateBrowserTranslation } from "@/tests/simulateBrowserTranslation";
import { createTestWrapper } from "@/tests/wrapper";
import { VALID_JWT } from "@/tests/seedAuthStore";
import { mockUserAuth } from "@/tests/factories";
import { getConfig, defaultConfig } from "@/env";
import { setEdition } from "@/tests/edition";
import AcceptInvite from "../AcceptInvite";

vi.mock("@/components/common/ConfirmDialog", async () => ({
  default: (await import("@/tests/mocks")).MockConfirmDialog,
}));

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

const INVITE_CODE = "INVITECODE12";
const VALID_PARAMS = `invite=${INVITE_CODE}`;

const resolvedData = {
  tenant_id: "t1",
  user_id: "u1",
  email: "alice@example.com",
  status: "confirmed",
};

function serveRegister(token: string | null) {
  const captured: { body?: Record<string, unknown> } = {};
  server.use(
    http.post("*/api/register", async ({ request }) => {
      captured.body = (await request.json()) as Record<string, unknown>;
      return token
        ? HttpResponse.json(mockUserAuth({ token }))
        : new HttpResponse(null, { status: 204 });
    }),
  );
  return captured;
}

function renderPage(search = VALID_PARAMS) {
  return render(<AcceptInvite />, {
    wrapper: createTestWrapper({
      initialEntries: [`/accept-invite${search ? "?" + search : ""}`],
    }),
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  setEdition("community");
  useAuthStore.setState({
    token: null,
    user: null,
    userId: null,
    email: null,
    tenant: null,
    role: null,
    name: null,
    loginWithToken: async () => {
      useAuthStore.setState({
        token: VALID_JWT,
        user: "alice",
        userId: "u1",
        email: "alice@example.com",
        name: "Alice",
      });
    },
  });
  server.use(
    http.get("*/api/invitations/resolve", () =>
      HttpResponse.json(resolvedData),
    ),
    http.patch(
      "*/api/namespaces/:tenant/invitations/accept",
      () => new HttpResponse(null, { status: 204 }),
    ),
    http.get("*/api/auth/token/:tenant", () =>
      HttpResponse.json({ token: VALID_JWT, role: "owner" }),
    ),
  );
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
      server.use(
        http.get("*/api/invitations/resolve", () => new Promise(() => {})),
      );
      renderPage(VALID_PARAMS);
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText(/checking invitation/i)).toBeInTheDocument();
    });
  });

  describe("branch: error (resolve rejects)", () => {
    it("renders the Invitation Unavailable heading", async () => {
      server.use(
        http.get("*/api/invitations/resolve", () =>
          HttpResponse.json({}, { status: 404 }),
        ),
      );
      renderPage(VALID_PARAMS);
      expect(
        await screen.findByRole("heading", { name: /invitation unavailable/i }),
      ).toBeInTheDocument();
    });
  });

  describe("branch: accept (authenticated as the invited user)", () => {
    beforeEach(() => {
      useAuthStore.setState({
        token: VALID_JWT,
        userId: "u1",
        email: "alice@example.com",
      });
    });

    it("accepts with the link's code and switches namespace using the resolved tenant", async () => {
      let accepted: unknown;
      server.use(
        http.patch(
          "*/api/namespaces/:tenant/invitations/accept",
          async ({ request }) => {
            accepted = await request.json();
            return new HttpResponse(null, { status: 204 });
          },
        ),
      );
      const user = userEvent.setup();
      renderPage(VALID_PARAMS);
      await user.click(
        await screen.findByRole("button", { name: /^accept$/i }),
      );
      const dialog = screen.getByRole("dialog");
      await user.click(
        within(dialog).getByRole("button", { name: /^accept invitation$/i }),
      );

      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /you're in/i }),
        ).toBeInTheDocument(),
      );
      expect(accepted).toEqual({ code: INVITE_CODE });
    });
  });

  describe("branch: wrong-user (authenticated as a different user)", () => {
    it("renders the Different Account Signed In heading", async () => {
      useAuthStore.setState({
        token: VALID_JWT,
        userId: "other-user-id",
        email: "other@example.com",
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
      server.use(
        http.get("*/api/invitations/resolve", () =>
          HttpResponse.json({ ...resolvedData, status: "invited" }),
        ),
      );
    });

    async function fillForm(user: ReturnType<typeof userEvent.setup>) {
      await user.type(screen.getByLabelText(/^name$/i), "Alice");
      await user.type(screen.getByLabelText(/^username$/i), "alice");
      await user.type(screen.getByLabelText(/^password$/i), "Secret123");
      await user.type(screen.getByLabelText(/confirm password/i), "Secret123");
    }

    it("shows the resolved email as text instead of an editable field", async () => {
      renderPage(VALID_PARAMS);
      await screen.findByRole("heading", { name: /you've been invited/i });
      expect(screen.getByText(/alice@example.com/)).toBeInTheDocument();
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
      const captured = serveRegister("tok");
      const user = userEvent.setup();
      renderPage(VALID_PARAMS);
      await screen.findByRole("heading", { name: /you've been invited/i });
      await fillForm(user);
      await user.click(screen.getByRole("button", { name: /join namespace/i }));

      await waitFor(() =>
        expect(captured.body).toEqual(
          expect.objectContaining({
            email: "alice@example.com",
            sig: INVITE_CODE,
            email_marketing: false,
          }),
        ),
      );

      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /you're in/i }),
        ).toBeInTheDocument(),
      );
    });

    it("announces the loading state to screen readers when switching namespace", async () => {
      serveRegister("tok");
      server.use(
        http.get("*/api/auth/token/:tenant", () => new Promise(() => {})),
      );
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

    it("continues to the email confirmation on Cloud when no token is returned", async () => {
      vi.mocked(getConfig).mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });
      serveRegister(null);
      const user = userEvent.setup();
      renderPage(VALID_PARAMS);
      await screen.findByRole("heading", { name: /you've been invited/i });
      await fillForm(user);
      await user.click(screen.getByRole("button", { name: /join namespace/i }));

      await waitFor(() =>
        expect(mockNavigate).toHaveBeenCalledWith(
          "/confirm-account?identifier=alice%40example.com",
        ),
      );
    });

    it("shows the Waiting for Approval screen when no token is returned off Cloud", async () => {
      vi.mocked(getConfig).mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
      serveRegister(null);
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
      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it("says the username or email is taken when the server refuses a field", async () => {
      server.use(
        http.post("*/api/register", () =>
          HttpResponse.json(
            { message: "user invalid", fields: { username: "invalid" } },
            { status: 400 },
          ),
        ),
      );
      const user = userEvent.setup();
      renderPage(VALID_PARAMS);
      await screen.findByRole("heading", { name: /you've been invited/i });
      await fillForm(user);
      await user.click(screen.getByRole("button", { name: /join namespace/i }));

      expect(
        await screen.findByText(/that username or email is already in use/i),
      ).toBeInTheDocument();
    });

    it("shows no sign-up error left from an earlier attempt", async () => {
      server.use(
        http.post("*/api/register", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      const user = userEvent.setup();
      const first = renderPage(VALID_PARAMS);
      await screen.findByRole("heading", { name: /you've been invited/i });
      await fillForm(user);
      await user.click(screen.getByRole("button", { name: /join namespace/i }));
      await screen.findByText(/something went wrong/i);
      first.unmount();

      renderPage(VALID_PARAMS);
      await screen.findByRole("heading", { name: /you've been invited/i });

      expect(screen.queryByText(/something went wrong/i)).not.toBeInTheDocument();
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
      server.use(
        http.get("*/api/invitations/resolve", () =>
          HttpResponse.json({ ...resolvedData, status: "not-confirmed" }),
        ),
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
      server.use(
        http.get("*/api/invitations/resolve", () =>
          HttpResponse.json({ ...resolvedData, status: "invited" }),
        ),
      );
      serveRegister("tok");
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
        token: VALID_JWT,
        userId: "u1",
        email: "alice@example.com",
      });
      const user = userEvent.setup();
      const { container } = renderPage(VALID_PARAMS);
      await screen.findByRole("heading", { name: /namespace invitation/i });

      simulateBrowserTranslation(container);
      await user.click(screen.getByRole("button", { name: /^accept$/i }));
      const dialog = screen.getByRole("dialog");
      await user.click(
        within(dialog).getByRole("button", { name: /^accept invitation$/i }),
      );

      expect(
        await screen.findByRole("heading", { name: /you're in/i }),
      ).toBeInTheDocument();
    });
  });
});
