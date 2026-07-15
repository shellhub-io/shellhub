import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import AddMemberDrawer from "../AddMemberDrawer";
import { defaultConfig, getConfig } from "@/env";
import type { SdkHttpError } from "@/api/errors";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

const mockGenerateLinkMutateAsync = vi.fn();

vi.mock("@/hooks/useInvitationMutations", () => ({
  useGenerateInvitationLink: () => ({
    mutateAsync: mockGenerateLinkMutateAsync,
    isPending: false,
  }),
}));

vi.mock("@/components/common/Drawer", () => ({
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
        <button type="button" onClick={onClose}>
          Close
        </button>
        {children}
        {footer ?? null}
      </div>
    );
  },
}));

vi.mock("@/components/common/CopyButton", () => ({
  default: ({ text }: { text: string }) => (
    <button type="button" aria-label="Copy" data-copy={text}>
      Copy
    </button>
  ),
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

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

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

function renderDrawer(open = true, onClose = vi.fn(), tenantId = "t1") {
  return render(
    <AddMemberDrawer open={open} onClose={onClose} tenantId={tenantId} />,
    { wrapper: createWrapper() },
  );
}

async function submit(
  user: ReturnType<typeof userEvent.setup>,
  email = "alice@example.com",
) {
  await user.type(screen.getByPlaceholderText(/user@example.com/i), email);
  await user.click(screen.getByRole("button", { name: /add member/i }));
}

function makeSdkError(status: number): SdkHttpError {
  return Object.assign(new Error("request failed"), {
    status,
    headers: new Headers(),
  });
}

/* ------------------------------------------------------------------ */
/* Setup                                                               */
/* ------------------------------------------------------------------ */

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
  mockGenerateLinkMutateAsync.mockResolvedValue({ link: null });
});

/* ================================================================== */
/* Tests                                                               */
/* ================================================================== */

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
      mockGenerateLinkMutateAsync.mockResolvedValue({
        link: "https://shellhub.example/accept-invite?invite=abc",
      });
      renderDrawer(true, vi.fn(), "t1");
      await submit(user, "bob@example.com");

      await waitFor(() =>
        expect(mockGenerateLinkMutateAsync).toHaveBeenCalledWith({
          path: { tenant: "t1" },
          body: { email: "bob@example.com", role: "operator" },
        }),
      );
    });

    it("shows the invitation link and copy button after generation", async () => {
      const generatedLink = "https://shellhub.example.com/invite/abc123";
      mockGenerateLinkMutateAsync.mockResolvedValue({ link: generatedLink });
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
      mockGenerateLinkMutateAsync.mockResolvedValue({
        link: "https://shellhub.example.com/invite/abc123",
      });
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
      mockGenerateLinkMutateAsync.mockResolvedValue({
        link: "https://shellhub.example.com/invite/abc123",
      });
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
      mockGenerateLinkMutateAsync.mockResolvedValue({ link: null });
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
      mockGenerateLinkMutateAsync.mockResolvedValue({
        link: "https://shellhub.example.com/invite/abc123",
      });
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
      expect(mockGenerateLinkMutateAsync).not.toHaveBeenCalled();
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
      mockGenerateLinkMutateAsync.mockRejectedValue(makeSdkError(400));
      const user = userEvent.setup();
      renderDrawer();
      await submit(user);

      await waitFor(() =>
        expect(screen.getByText(/invalid email or role/i)).toBeInTheDocument(),
      );
    });

    it("shows 403 error as permission denied message", async () => {
      mockGenerateLinkMutateAsync.mockRejectedValue(makeSdkError(403));
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
      mockGenerateLinkMutateAsync.mockRejectedValue(makeSdkError(404));
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
      mockGenerateLinkMutateAsync.mockRejectedValue(makeSdkError(409));
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
      mockGenerateLinkMutateAsync.mockRejectedValue(makeSdkError(500));
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
      mockGenerateLinkMutateAsync.mockRejectedValue(new Error("network error"));
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
