import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Routes, Route } from "react-router-dom";
import { createTestWrapper } from "@/tests/wrapper";
import SSHApproval from "../SSHApproval";
const mockGetSshApproval = vi.hoisted(() => vi.fn());
const mockWebTerminalReauth = vi.hoisted(() => vi.fn());

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return {
    ...actual,
    getSshApproval: mockGetSshApproval,
    webTerminalReauth: mockWebTerminalReauth,
  };
});

const mockConfirmMutateAsync = vi.fn();
const mockRejectMutateAsync = vi.fn();

vi.mock("@/hooks/useSSHIdentityMutations", () => ({
  useConfirmSSHApproval: () => ({ mutateAsync: mockConfirmMutateAsync }),
  useRejectSSHApproval: () => ({ mutateAsync: mockRejectMutateAsync }),
}));

vi.mock("@/hooks/useNamespaces", () => ({
  useNamespaces: () => ({
    namespaces: [{ name: "dev", tenant_id: "tenant1" }],
  }),
}));

vi.mock("@/hooks/useNamespaceMutations", () => ({
  useSwitchNamespace: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

vi.mock("@/stores/authStore", () => ({
  useAuthStore: Object.assign(
    (selector?: (s: unknown) => unknown) => {
      const state = {
        tenant: "tenant1",
        name: "Gustavo",
        email: "gustavo@example.com",
        user: "gustavo",
        logout: vi.fn(),
      };
      return selector ? selector(state) : state;
    },
    { getState: () => ({}) },
  ),
}));

const approval = (overrides: Record<string, unknown> = {}) => ({
  data: {
    code: "WXYZ2K7Q",
    kind: "identity",
    fingerprint: "SHA256:abc",
    sshid: "root@dev.device",
    device_name: "device",
    username: "root",
    ip_address: "10.0.0.1",
    requested_at: "2026-07-27T12:00:00Z",
    expires_in_seconds: 90,
    namespace: "dev",
    state: "pending",
    ...overrides,
  },
});

function renderAt(path: string) {
  return render(
    <Routes>
      <Route
        path="/ssh-identities/new/:code"
        element={<SSHApproval flow="new" />}
      />
      <Route
        path="/ssh-identities/confirm/:code"
        element={<SSHApproval flow="confirm" />}
      />
    </Routes>,
    { wrapper: createTestWrapper({ initialEntries: [path] }) },
  );
}

describe("SSHApproval", () => {
  beforeEach(() => {
    mockGetSshApproval.mockReset();
    mockWebTerminalReauth.mockReset();
    mockConfirmMutateAsync.mockReset();
    mockRejectMutateAsync.mockReset();
  });

  it("asks to add the key, and names the account and namespace it lands in", async () => {
    mockGetSshApproval.mockResolvedValue(approval());

    renderAt("/ssh-identities/new/WXYZ2K7Q");

    expect(
      await screen.findByText(/add this ssh key to your identities/i),
    ).toBeInTheDocument();
    expect(screen.getByText("Gustavo")).toBeInTheDocument();
    expect(screen.getByText("dev")).toBeInTheDocument();
    expect(screen.getByText("SHA256:abc")).toBeInTheDocument();
  });

  it("asks to re-authenticate, and says the window covers more than this login", async () => {
    mockGetSshApproval.mockResolvedValue(
      approval({ kind: "reauth", reauth_period: 43200 }),
    );

    renderAt("/ssh-identities/confirm/WXYZ2K7Q");

    expect(
      await screen.findByText(/re-authenticate to continue/i),
    ).toBeInTheDocument();
    expect(
      screen.getByText(/won't ask for the next 12 hours/i),
    ).toBeInTheDocument();
    // Nothing is being bound, so the key-to-account card must not appear.
    expect(screen.queryByText(/added to/i)).not.toBeInTheDocument();
  });

  it("says nothing about a window when the policy asks every time", async () => {
    mockGetSshApproval.mockResolvedValue(
      approval({ kind: "reauth", reauth_period: 0 }),
    );

    renderAt("/ssh-identities/confirm/WXYZ2K7Q");

    expect(
      await screen.findByText(/re-authenticate to continue/i),
    ).toBeInTheDocument();
    expect(
      screen.queryByText(/won't ask for the next/i),
    ).not.toBeInTheDocument();
  });

  it("redirects a reauth code opened on the add route", async () => {
    mockGetSshApproval.mockResolvedValue(approval({ kind: "reauth" }));

    renderAt("/ssh-identities/new/WXYZ2K7Q");

    expect(
      await screen.findByText(/re-authenticate to continue/i),
    ).toBeInTheDocument();
  });

  it("redirects an identity code opened on the reauth route", async () => {
    mockGetSshApproval.mockResolvedValue(approval());

    renderAt("/ssh-identities/confirm/WXYZ2K7Q");

    expect(
      await screen.findByText(/add this ssh key to your identities/i),
    ).toBeInTheDocument();
  });

  it("confirms the request and reports the outcome", async () => {
    mockGetSshApproval.mockResolvedValue(approval());
    mockConfirmMutateAsync.mockResolvedValue({});

    renderAt("/ssh-identities/new/WXYZ2K7Q");

    await userEvent.click(
      await screen.findByRole("button", { name: /add key/i }),
    );

    await waitFor(() =>
      expect(mockConfirmMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({ path: { code: "WXYZ2K7Q" } }),
      ),
    );
    expect(await screen.findByText("Key added")).toBeInTheDocument();
  });

  it("rejects the request and reports the outcome", async () => {
    mockGetSshApproval.mockResolvedValue(approval());
    mockRejectMutateAsync.mockResolvedValue({});

    renderAt("/ssh-identities/new/WXYZ2K7Q");

    await userEvent.click(
      await screen.findByRole("button", { name: /reject/i }),
    );

    await waitFor(() => expect(mockRejectMutateAsync).toHaveBeenCalled());
    expect(await screen.findByText("Rejected")).toBeInTheDocument();
  });

  it("reads a 404 as an expired request", async () => {
    mockGetSshApproval.mockRejectedValue({ status: 404 });

    renderAt("/ssh-identities/new/WXYZ2K7Q");

    expect(await screen.findByText("Request expired")).toBeInTheDocument();
  });

  // The details exist so somebody checks them. A secret field beside them
  // invites typing past the check, so it only appears once the check is done.
  it("keeps the factor out of sight until the login has been reviewed", async () => {
    const user = userEvent.setup();
    mockGetSshApproval.mockResolvedValue(approval({ kind: "reauth" }));

    renderAt("/ssh-identities/confirm/WXYZ2K7Q");
    await screen.findByText(/re-authenticate to continue/i);

    expect(
      screen.queryByLabelText(/account password/i),
    ).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /continue/i }));
    expect(
      await screen.findByLabelText(/account password/i),
    ).toBeInTheDocument();
  });

  it("goes back to the details without deciding anything", async () => {
    const user = userEvent.setup();
    mockGetSshApproval.mockResolvedValue(approval({ kind: "reauth" }));

    renderAt("/ssh-identities/confirm/WXYZ2K7Q");
    await screen.findByText(/re-authenticate to continue/i);
    await user.click(screen.getByRole("button", { name: /continue/i }));

    await user.click(screen.getByRole("button", { name: /back/i }));

    expect(
      await screen.findByText(/re-authenticate to continue/i),
    ).toBeInTheDocument();
    expect(
      screen.queryByLabelText(/account password/i),
    ).not.toBeInTheDocument();
    expect(mockWebTerminalReauth).not.toHaveBeenCalled();
  });

  // Proving the factor is what releases the held login, so the call has to carry
  // the code and the key it was asked for.
  it("proves the password and reports the login released", async () => {
    const user = userEvent.setup();
    mockGetSshApproval.mockResolvedValue(approval({ kind: "reauth" }));
    mockWebTerminalReauth.mockResolvedValue({
      data: undefined,
    });

    renderAt("/ssh-identities/confirm/WXYZ2K7Q");
    await screen.findByText(/re-authenticate to continue/i);
    await user.click(screen.getByRole("button", { name: /continue/i }));

    await user.type(
      await screen.findByLabelText(/account password/i),
      "hunter2",
    );
    await user.click(screen.getByRole("button", { name: /re-authenticate/i }));

    await waitFor(() =>
      expect(mockWebTerminalReauth).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({
            password: "hunter2",
            approval_code: "WXYZ2K7Q",
            fingerprint: "SHA256:abc",
          }),
        }),
      ),
    );
    expect(await screen.findByText(/re-authenticated/i)).toBeInTheDocument();
  });

  // Adding a key asks for no factor, so it must not have grown a step.
  it("leaves the add-key flow at a single step", async () => {
    mockGetSshApproval.mockResolvedValue(approval());

    renderAt("/ssh-identities/new/WXYZ2K7Q");
    await screen.findByText(/add this ssh key/i);

    expect(
      screen.getByRole("button", { name: /add key/i }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /continue/i }),
    ).not.toBeInTheDocument();
  });
});
