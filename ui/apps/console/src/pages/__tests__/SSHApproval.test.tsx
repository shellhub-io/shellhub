import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Routes, Route } from "react-router-dom";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse, paginatedResponse } from "@/tests/sdk";
import { mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import SSHApproval from "../SSHApproval";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getSshApproval: vi.fn(),
    webTerminalReauth: vi.fn(),
    confirmSshApproval: vi.fn(),
    rejectSshApproval: vi.fn(),
    getNamespaces: vi.fn(),
    getNamespaceToken: vi.fn(),
  }),
);

const approval = (overrides: Record<string, unknown> = {}) => ({
  data: {
    code: "WXYZ2K7Q",
    kind: "identity",
    fingerprint: "SHA256:abc",
    sshid: "root@my-namespace.device",
    device_name: "device",
    username: "root",
    ip_address: "10.0.0.1",
    requested_at: "2026-07-27T12:00:00Z",
    expires_in_seconds: 90,
    namespace: "my-namespace",
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
    vi.clearAllMocks();
    seedAuthStore();
    sdk.getNamespaces.mockResolvedValue(paginatedResponse([mockNamespace()]));
    sdk.getNamespaceToken.mockResolvedValue(
      mockSdkResponse({ token: "jwt-token" }),
    );
  });

  it("asks to add the key, and names the account and namespace it lands in", async () => {
    sdk.getSshApproval.mockResolvedValue(approval());

    renderAt("/ssh-identities/new/WXYZ2K7Q");

    expect(
      await screen.findByText(/add this ssh key to your identities/i),
    ).toBeInTheDocument();
    expect(screen.getByText("Admin User")).toBeInTheDocument();
    expect(screen.getByText("my-namespace")).toBeInTheDocument();
    expect(screen.getByText("SHA256:abc")).toBeInTheDocument();
  });

  it("asks to re-authenticate, and says the window covers more than this login", async () => {
    sdk.getSshApproval.mockResolvedValue(
      approval({ kind: "reauth", reauth_period: 43200 }),
    );

    renderAt("/ssh-identities/confirm/WXYZ2K7Q");

    expect(
      await screen.findByText(/re-authenticate to continue/i),
    ).toBeInTheDocument();
    expect(
      screen.getByText(/won't ask for the next 12 hours/i),
    ).toBeInTheDocument();
    expect(screen.queryByText(/added to/i)).not.toBeInTheDocument();
  });

  it("says nothing about a window when the policy asks every time", async () => {
    sdk.getSshApproval.mockResolvedValue(
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
    sdk.getSshApproval.mockResolvedValue(approval({ kind: "reauth" }));

    renderAt("/ssh-identities/new/WXYZ2K7Q");

    expect(
      await screen.findByText(/re-authenticate to continue/i),
    ).toBeInTheDocument();
  });

  it("redirects an identity code opened on the reauth route", async () => {
    sdk.getSshApproval.mockResolvedValue(approval());

    renderAt("/ssh-identities/confirm/WXYZ2K7Q");

    expect(
      await screen.findByText(/add this ssh key to your identities/i),
    ).toBeInTheDocument();
  });

  it("confirms the request and reports the outcome", async () => {
    sdk.getSshApproval.mockResolvedValue(approval());
    sdk.confirmSshApproval.mockResolvedValue(mockSdkResponse(undefined));

    renderAt("/ssh-identities/new/WXYZ2K7Q");

    await userEvent.click(
      await screen.findByRole("button", { name: /add key/i }),
    );

    await waitFor(() =>
      expect(sdk.confirmSshApproval).toHaveBeenCalledWith(
        expect.objectContaining({
          path: { code: "WXYZ2K7Q" },
          throwOnError: true,
        }),
      ),
    );
    expect(await screen.findByText("Key added")).toBeInTheDocument();
  });

  it("rejects the request and reports the outcome", async () => {
    sdk.getSshApproval.mockResolvedValue(approval());
    sdk.rejectSshApproval.mockResolvedValue(mockSdkResponse(undefined));

    renderAt("/ssh-identities/new/WXYZ2K7Q");

    await userEvent.click(
      await screen.findByRole("button", { name: /reject/i }),
    );

    await waitFor(() => expect(sdk.rejectSshApproval).toHaveBeenCalled());
    expect(await screen.findByText("Rejected")).toBeInTheDocument();
  });

  it("reads a 404 as an expired request", async () => {
    sdk.getSshApproval.mockRejectedValue({ status: 404 });

    renderAt("/ssh-identities/new/WXYZ2K7Q");

    expect(await screen.findByText("Request expired")).toBeInTheDocument();
  });

  it("keeps the factor out of sight until the login has been reviewed", async () => {
    const user = userEvent.setup();
    sdk.getSshApproval.mockResolvedValue(approval({ kind: "reauth" }));

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
    sdk.getSshApproval.mockResolvedValue(approval({ kind: "reauth" }));

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
    expect(sdk.webTerminalReauth).not.toHaveBeenCalled();
  });

  it("proves the password and reports the login released", async () => {
    const user = userEvent.setup();
    sdk.getSshApproval.mockResolvedValue(approval({ kind: "reauth" }));
    sdk.webTerminalReauth.mockResolvedValue({
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
      expect(sdk.webTerminalReauth).toHaveBeenCalledWith(
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

  it("leaves the add-key flow at a single step", async () => {
    sdk.getSshApproval.mockResolvedValue(approval());

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
