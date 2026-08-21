import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse, makeSdkError } from "@/tests/sdk";
import { defaultConfig, getConfig } from "@/env";
import AddMemberDrawer from "../AddMemberDrawer";

const mockGenerateInvitationLink = vi.hoisted(() => vi.fn());

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return { ...actual, generateInvitationLink: mockGenerateInvitationLink };
});

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

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
  mockGenerateInvitationLink.mockResolvedValue(mockSdkResponse({ link: null }));
});

describe("AddMemberDrawer", () => {
  describe("rendering", () => {
    it("renders the Add Member title when open", () => {
      renderDrawer();
      expect(
        screen.getByRole("heading", { name: /add member/i }),
      ).toBeInTheDocument();
    });

    it("renders nothing when closed", () => {
      const { container } = renderDrawer(false);
      expect(container).toBeEmptyDOMElement();
    });

    it("renders the email input", () => {
      renderDrawer();
      expect(
        screen.getByPlaceholderText(/user@example.com/i),
      ).toBeInTheDocument();
    });

    it("has no delivery-choice checkbox — the flow always both emails and returns a link", () => {
      renderDrawer();
      expect(
        screen.queryByRole("checkbox", { name: /link instead/i }),
      ).not.toBeInTheDocument();
    });
  });

  describe("submit", () => {
    it("always generates the invitation link (single channel)", async () => {
      const user = userEvent.setup();
      mockGenerateInvitationLink.mockResolvedValue(
        mockSdkResponse({
          link: "https://shellhub.example/accept-invite?invite=abc",
        }),
      );
      renderDrawer(true, vi.fn(), "t1");
      await submit(user, "bob@example.com");

      await waitFor(() =>
        expect(mockGenerateInvitationLink).toHaveBeenCalledWith(
          expect.objectContaining({
            path: { tenant: "t1" },
            body: { email: "bob@example.com", role: "operator" },
          }),
        ),
      );
    });

    it("shows the invitation link and copy button after generation", async () => {
      const generatedLink = "https://shellhub.example.com/invite/abc123";
      mockGenerateInvitationLink.mockResolvedValue(
        mockSdkResponse({ link: generatedLink }),
      );
      const user = userEvent.setup();
      renderDrawer();
      await submit(user);

      await waitFor(() =>
        expect(screen.getByText(generatedLink)).toBeInTheDocument(),
      );
      expect(screen.getByRole("button", { name: /copy/i })).toBeInTheDocument();
      expect(
        screen.getByRole("heading", { name: /invitation link/i }),
      ).toBeInTheDocument();
    });

    it("mentions the email on cloud", async () => {
      mockGenerateInvitationLink.mockResolvedValue(
        mockSdkResponse({
          link: "https://shellhub.example.com/invite/abc123",
        }),
      );
      const user = userEvent.setup();
      renderDrawer();
      await submit(user);

      await waitFor(() =>
        expect(
          screen.getByText(/we emailed the invitation/i),
        ).toBeInTheDocument(),
      );
    });

    it("does not mention email on a non-cloud edition (link-only)", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
      mockGenerateInvitationLink.mockResolvedValue(
        mockSdkResponse({
          link: "https://shellhub.example.com/invite/abc123",
        }),
      );
      const user = userEvent.setup();
      renderDrawer();
      await submit(user);

      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /invitation link/i }),
        ).toBeInTheDocument(),
      );
      expect(
        screen.queryByText(/we emailed the invitation/i),
      ).not.toBeInTheDocument();
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
      mockGenerateInvitationLink.mockResolvedValue(
        mockSdkResponse({
          link: "https://shellhub.example.com/invite/abc123",
        }),
      );
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
      const user = userEvent.setup();
      renderDrawer();
      await user.type(screen.getByPlaceholderText(/user@example.com/i), "bad");
      await user.keyboard("{Enter}");
      expect(mockGenerateInvitationLink).not.toHaveBeenCalled();
    });

    it("disables the submit button when email field is empty", () => {
      renderDrawer();
      expect(
        screen.getByRole("button", { name: /add member/i }),
      ).toBeDisabled();
    });

    it("disables the submit button when email is invalid (non-empty)", async () => {
      const user = userEvent.setup();
      renderDrawer();
      await user.type(
        screen.getByPlaceholderText(/user@example.com/i),
        "not-an-email",
      );
      expect(
        screen.getByRole("button", { name: /add member/i }),
      ).toBeDisabled();
    });
  });

  describe("error handling", () => {
    it("shows 400 error as invalid email/role message", async () => {
      mockGenerateInvitationLink.mockRejectedValue(makeSdkError(400));
      const user = userEvent.setup();
      renderDrawer();
      await submit(user);

      await waitFor(() =>
        expect(screen.getByText(/invalid email or role/i)).toBeInTheDocument(),
      );
    });

    it("shows 403 error as permission denied message", async () => {
      mockGenerateInvitationLink.mockRejectedValue(makeSdkError(403));
      const user = userEvent.setup();
      renderDrawer();
      await submit(user);

      await waitFor(() =>
        expect(
          screen.getByText(/don't have permission to invite/i),
        ).toBeInTheDocument(),
      );
    });

    it("shows 404 error as no account message", async () => {
      mockGenerateInvitationLink.mockRejectedValue(makeSdkError(404));
      const user = userEvent.setup();
      renderDrawer();
      await submit(user);

      await waitFor(() =>
        expect(
          screen.getByText(/no account exists for this email/i),
        ).toBeInTheDocument(),
      );
    });

    it("shows 409 error as already member message", async () => {
      mockGenerateInvitationLink.mockRejectedValue(makeSdkError(409));
      const user = userEvent.setup();
      renderDrawer();
      await submit(user);

      await waitFor(() =>
        expect(
          screen.getByText(/already a member or has a pending invitation/i),
        ).toBeInTheDocument(),
      );
    });

    it("shows generic error for unexpected status codes", async () => {
      mockGenerateInvitationLink.mockRejectedValue(makeSdkError(500));
      const user = userEvent.setup();
      renderDrawer();
      await submit(user);

      await waitFor(() =>
        expect(
          screen.getByText(/failed to send invitation/i),
        ).toBeInTheDocument(),
      );
    });

    it("shows generic error for non-SDK errors", async () => {
      mockGenerateInvitationLink.mockRejectedValue(new Error("network error"));
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
