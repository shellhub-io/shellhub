import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { defaultConfig, getConfig } from "@/env";
import AddMemberDrawer from "../AddMemberDrawer";

vi.mock("@/components/common/Drawer", async () => ({
  default: (await import("@/tests/mocks")).MockDrawer,
}));

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
}));

vi.mock("@/utils/styles", () => ({
  LABEL: "label",
  INPUT: "input",
  INPUT_BASE: "input-base",
  INPUT_ERROR: "input-error",
  INPUT_MONO: "input-mono",
  INPUT_MONO_ERROR: "input-mono-error",
}));

const INVITE_LINK = "https://shellhub.example.com/invite/abc123";

const mockGetConfig = vi.mocked(getConfig);

function renderDrawer(open = true, onClose = vi.fn(), tenantId = "t1") {
  return render(
    <AddMemberDrawer open={open} onClose={onClose} tenantId={tenantId} />,
    { wrapper: createTestWrapper({ initialEntries: ["/"] }) },
  );
}

async function submit(
  user: ReturnType<typeof userEvent.setup>,
  email = "alice@example.com",
) {
  await user.type(screen.getByPlaceholderText(/user@example.com/i), email);
  await user.click(screen.getByRole("button", { name: /add member/i }));
}

function setInviteResponse(link: string | null) {
  server.use(
    http.post("*/api/namespaces/:tenant/invitations/links", () =>
      HttpResponse.json({ link }),
    ),
  );
}

function setInviteError(status: number) {
  server.use(
    http.post("*/api/namespaces/:tenant/invitations/links", () =>
      HttpResponse.json({}, { status }),
    ),
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
  setInviteResponse(null);
});

describe("AddMemberDrawer", () => {
  it("has no delivery-choice checkbox — the flow always both emails and returns a link", () => {
    renderDrawer();
    expect(
      screen.queryByRole("checkbox", { name: /link instead/i }),
    ).not.toBeInTheDocument();
  });

  describe("submit", () => {
    it("shows the invitation link and copy button after generation", async () => {
      setInviteResponse(INVITE_LINK);
      const user = userEvent.setup();
      renderDrawer();
      await submit(user);

      await waitFor(() =>
        expect(screen.getByText(INVITE_LINK)).toBeInTheDocument(),
      );
      expect(screen.getByRole("button", { name: /copy/i })).toBeInTheDocument();
    });

    it.each([
      ["cloud", true],
      ["enterprise", false],
    ] as const)("the %s result screen mentions the email: %s", async (
      edition,
      mentionsEmail,
    ) => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, edition });
      setInviteResponse(INVITE_LINK);
      const user = userEvent.setup();
      renderDrawer();
      await submit(user);

      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /invitation link/i }),
        ).toBeInTheDocument(),
      );
      const email = screen.queryByText(/we emailed the invitation/i);
      if (mentionsEmail) expect(email).toBeInTheDocument();
      else expect(email).not.toBeInTheDocument();
    });

    it("shows 'Member Added' when an existing account is added directly (no link)", async () => {
      const user = userEvent.setup();
      renderDrawer();
      await submit(user, "bob@example.com");

      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /member added/i }),
        ).toBeInTheDocument(),
      );
      expect(
        screen.queryByRole("button", { name: /copy/i }),
      ).not.toBeInTheDocument();
    });

    it("does not close on success — the result screen stays until 'Done'", async () => {
      setInviteResponse(INVITE_LINK);
      const onClose = vi.fn();
      const user = userEvent.setup();
      renderDrawer(true, onClose, "t1");
      await submit(user);

      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /invitation link/i }),
        ).toBeInTheDocument(),
      );
      expect(onClose).not.toHaveBeenCalled();
    });
  });

  describe("form validation", () => {
    it("does not call the mutation when email is invalid (Enter submit)", async () => {
      const apiCalled = vi.fn();
      server.use(
        http.post("*/api/namespaces/:tenant/invitations/links", () => {
          apiCalled();
          return HttpResponse.json({ link: null });
        }),
      );
      const user = userEvent.setup();
      renderDrawer();
      await user.type(screen.getByPlaceholderText(/user@example.com/i), "bad");
      await user.keyboard("{Enter}");
      expect(apiCalled).not.toHaveBeenCalled();
    });

    it.each([
      ["empty", ""],
      ["invalid", "not-an-email"],
    ])("disables the submit button when email is %s", async (_label, email) => {
      const user = userEvent.setup();
      renderDrawer();
      if (email)
        await user.type(screen.getByPlaceholderText(/user@example.com/i), email);
      expect(
        screen.getByRole("button", { name: /add member/i }),
      ).toBeDisabled();
    });
  });

  describe("error handling", () => {
    it.each([
      [400, /invalid email or role/i],
      [403, /don't have permission to invite/i],
      [404, /no account exists for this email/i],
      [409, /already a member or has a pending invitation/i],
      [500, /failed to send invitation/i],
    ])("a %i response reports '%s'", async (status, message) => {
      setInviteError(status);
      const user = userEvent.setup();
      renderDrawer();
      await submit(user);

      await waitFor(() =>
        expect(screen.getByText(message)).toBeInTheDocument(),
      );
    });

    it("shows generic error for network errors", async () => {
      server.use(
        http.post("*/api/namespaces/:tenant/invitations/links", () =>
          HttpResponse.error(),
        ),
      );
      const user = userEvent.setup();
      renderDrawer();
      await submit(user);

      await waitFor(() =>
        expect(
          screen.getByText(/failed to send invitation/i),
        ).toBeInTheDocument(),
      );
    });
  });
});
