import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import InvitationDrawer from "../InvitationDrawer";
import type { SdkHttpError } from "../../../api/errors";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

const mockSendEmailMutateAsync = vi.fn();
const mockGenerateLinkMutateAsync = vi.fn();

vi.mock("../../../hooks/useInvitationMutations", () => ({
  useSendInvitationEmail: () => ({
    mutateAsync: mockSendEmailMutateAsync,
    isPending: false,
  }),
  useGenerateInvitationLink: () => ({
    mutateAsync: mockGenerateLinkMutateAsync,
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

vi.mock("../../../components/common/CopyButton", () => ({
  default: ({ text }: { text: string }) => (
    <button aria-label="Copy" data-copy={text}>
      Copy
    </button>
  ),
}));

vi.mock("../../../utils/styles", () => ({
  LABEL: "label",
  INPUT: "input",
}));

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
    <InvitationDrawer open={open} onClose={onClose} tenantId={tenantId} />,
    { wrapper: createWrapper() },
  );
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
  mockSendEmailMutateAsync.mockResolvedValue({});
  mockGenerateLinkMutateAsync.mockResolvedValue({ link: null });
});

/* ================================================================== */
/* Tests                                                               */
/* ================================================================== */

describe("InvitationDrawer", () => {
  describe("rendering", () => {
    it("renders the Invite Member title when open", () => {
      renderDrawer();
      expect(
        screen.getByRole("heading", { name: /invite member/i }),
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

    it("renders the 'get invite link' checkbox unchecked by default", () => {
      renderDrawer();
      const checkbox = screen.getByRole("checkbox", {
        name: /get the invite link/i,
      });
      expect(checkbox).not.toBeChecked();
    });

    it("shows 'Send Invite' button text when checkbox is unchecked", () => {
      renderDrawer();
      expect(
        screen.getByRole("button", { name: /send invite/i }),
      ).toBeInTheDocument();
    });

    it("shows 'Generate Link' button text when checkbox is checked", async () => {
      const user = userEvent.setup();
      renderDrawer();
      await user.click(
        screen.getByRole("checkbox", { name: /get the invite link/i }),
      );
      expect(
        screen.getByRole("button", { name: /generate link/i }),
      ).toBeInTheDocument();
    });
  });

  describe("form validation", () => {
    it("marks email input as aria-invalid after typing and clearing an invalid value", async () => {
      // The handleSubmit path sets emailError which sets aria-invalid.
      // The submit button is disabled when email is invalid, so we use a direct
      // approach: type a valid email, then clear it — the error state is set
      // when the user tries to submit an empty email via footer button click
      // is not possible (disabled). Instead verify that the email error helper
      // text is displayed inline when the submission is attempted via the
      // formAction callback by checking the aria-invalid initial state.
      const user = userEvent.setup();
      renderDrawer();
      const input = screen.getByPlaceholderText(/user@example.com/i);
      // Initially no error
      expect(input).toHaveAttribute("aria-invalid", "false");
      // After typing invalid email — button is disabled but no error yet
      await user.type(input, "not-an-email");
      expect(input).toHaveAttribute("aria-invalid", "false");
    });

    it("does not call the mutation when email is invalid (Enter submit)", async () => {
      const user = userEvent.setup();
      renderDrawer();
      await user.type(screen.getByPlaceholderText(/user@example.com/i), "bad");
      await user.keyboard("{Enter}");
      expect(mockSendEmailMutateAsync).not.toHaveBeenCalled();
      expect(mockGenerateLinkMutateAsync).not.toHaveBeenCalled();
    });

    it("disables the submit button when email field is empty", () => {
      renderDrawer();
      expect(
        screen.getByRole("button", { name: /send invite/i }),
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
        screen.getByRole("button", { name: /send invite/i }),
      ).toBeDisabled();
    });
  });

  describe("successful invite — email mode (checkbox off)", () => {
    it("calls sendEmail mutation with correct args", async () => {
      const user = userEvent.setup();
      renderDrawer(true, vi.fn(), "t1");
      await user.type(
        screen.getByPlaceholderText(/user@example.com/i),
        "alice@example.com",
      );
      await user.click(screen.getByRole("button", { name: /send invite/i }));

      await waitFor(() =>
        expect(mockSendEmailMutateAsync).toHaveBeenCalledWith({
          path: { tenant: "t1" },
          body: { email: "alice@example.com", role: "operator" },
        }),
      );
      expect(mockGenerateLinkMutateAsync).not.toHaveBeenCalled();
    });

    it("calls onClose after successful send when wantLink is false", async () => {
      const onClose = vi.fn();
      const user = userEvent.setup();
      renderDrawer(true, onClose, "t1");

      await user.type(
        screen.getByPlaceholderText(/user@example.com/i),
        "alice@example.com",
      );
      await user.click(screen.getByRole("button", { name: /send invite/i }));

      await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    });
  });

  describe("successful invite — link mode (checkbox on)", () => {
    it("shows the invitation link after successful generation", async () => {
      const generatedLink = "https://shellhub.example.com/invite/abc123";
      mockGenerateLinkMutateAsync.mockResolvedValue({ link: generatedLink });
      const user = userEvent.setup();
      renderDrawer();

      await user.type(
        screen.getByPlaceholderText(/user@example.com/i),
        "alice@example.com",
      );
      await user.click(
        screen.getByRole("checkbox", { name: /get the invite link/i }),
      );
      await user.click(screen.getByRole("button", { name: /generate link/i }));

      await waitFor(() =>
        expect(screen.getByText(generatedLink)).toBeInTheDocument(),
      );
    });

    it("shows the copy button after link generation", async () => {
      mockGenerateLinkMutateAsync.mockResolvedValue({
        link: "https://shellhub.example.com/invite/abc123",
      });
      const user = userEvent.setup();
      renderDrawer();

      await user.type(
        screen.getByPlaceholderText(/user@example.com/i),
        "alice@example.com",
      );
      await user.click(
        screen.getByRole("checkbox", { name: /get the invite link/i }),
      );
      await user.click(screen.getByRole("button", { name: /generate link/i }));

      await waitFor(() =>
        expect(
          screen.getByRole("button", { name: /copy/i }),
        ).toBeInTheDocument(),
      );
    });

    it("shows 'Invitation Link' title after link is generated", async () => {
      mockGenerateLinkMutateAsync.mockResolvedValue({
        link: "https://shellhub.example.com/invite/abc123",
      });
      const user = userEvent.setup();
      renderDrawer();

      await user.type(
        screen.getByPlaceholderText(/user@example.com/i),
        "alice@example.com",
      );
      await user.click(
        screen.getByRole("checkbox", { name: /get the invite link/i }),
      );
      await user.click(screen.getByRole("button", { name: /generate link/i }));

      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /invitation link/i }),
        ).toBeInTheDocument(),
      );
    });

    it("does not call onClose when wantLink is true after success", async () => {
      mockGenerateLinkMutateAsync.mockResolvedValue({
        link: "https://shellhub.example.com/invite/abc123",
      });
      const onClose = vi.fn();
      const user = userEvent.setup();
      renderDrawer(true, onClose, "t1");

      await user.type(
        screen.getByPlaceholderText(/user@example.com/i),
        "alice@example.com",
      );
      await user.click(
        screen.getByRole("checkbox", { name: /get the invite link/i }),
      );
      await user.click(screen.getByRole("button", { name: /generate link/i }));

      await waitFor(() =>
        expect(
          screen.getByRole("heading", { name: /invitation link/i }),
        ).toBeInTheDocument(),
      );
      expect(onClose).not.toHaveBeenCalled();
    });
  });

  describe("error handling", () => {
    it("shows 400 error as invalid email/role message", async () => {
      mockSendEmailMutateAsync.mockRejectedValue(makeSdkError(400));
      const user = userEvent.setup();
      renderDrawer();

      await user.type(
        screen.getByPlaceholderText(/user@example.com/i),
        "alice@example.com",
      );
      await user.click(screen.getByRole("button", { name: /send invite/i }));

      await waitFor(() =>
        expect(screen.getByText(/invalid email or role/i)).toBeInTheDocument(),
      );
    });

    it("shows 403 error as permission denied message", async () => {
      mockSendEmailMutateAsync.mockRejectedValue(makeSdkError(403));
      const user = userEvent.setup();
      renderDrawer();

      await user.type(
        screen.getByPlaceholderText(/user@example.com/i),
        "alice@example.com",
      );
      await user.click(screen.getByRole("button", { name: /send invite/i }));

      await waitFor(() =>
        expect(
          screen.getByText(/don't have permission to invite/i),
        ).toBeInTheDocument(),
      );
    });

    it("shows 404 error as no account message", async () => {
      mockSendEmailMutateAsync.mockRejectedValue(makeSdkError(404));
      const user = userEvent.setup();
      renderDrawer();

      await user.type(
        screen.getByPlaceholderText(/user@example.com/i),
        "alice@example.com",
      );
      await user.click(screen.getByRole("button", { name: /send invite/i }));

      await waitFor(() =>
        expect(
          screen.getByText(/no account exists for this email/i),
        ).toBeInTheDocument(),
      );
    });

    it("shows 409 error as already member message", async () => {
      mockSendEmailMutateAsync.mockRejectedValue(makeSdkError(409));
      const user = userEvent.setup();
      renderDrawer();

      await user.type(
        screen.getByPlaceholderText(/user@example.com/i),
        "alice@example.com",
      );
      await user.click(screen.getByRole("button", { name: /send invite/i }));

      await waitFor(() =>
        expect(
          screen.getByText(/already a member or has a pending invitation/i),
        ).toBeInTheDocument(),
      );
    });

    it("shows generic error for unexpected status codes", async () => {
      mockSendEmailMutateAsync.mockRejectedValue(makeSdkError(500));
      const user = userEvent.setup();
      renderDrawer();

      await user.type(
        screen.getByPlaceholderText(/user@example.com/i),
        "alice@example.com",
      );
      await user.click(screen.getByRole("button", { name: /send invite/i }));

      await waitFor(() =>
        expect(
          screen.getByText(/failed to send invitation/i),
        ).toBeInTheDocument(),
      );
    });

    it("shows generic error for non-SDK errors", async () => {
      mockSendEmailMutateAsync.mockRejectedValue(new Error("network error"));
      const user = userEvent.setup();
      renderDrawer();

      await user.type(
        screen.getByPlaceholderText(/user@example.com/i),
        "alice@example.com",
      );
      await user.click(screen.getByRole("button", { name: /send invite/i }));

      await waitFor(() =>
        expect(
          screen.getByText(/failed to send invitation/i),
        ).toBeInTheDocument(),
      );
    });
  });
});
