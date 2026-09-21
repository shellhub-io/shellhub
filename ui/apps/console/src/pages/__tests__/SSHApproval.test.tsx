import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Routes, Route } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import SSHApproval from "../SSHApproval";

function approvalData(overrides: Record<string, unknown> = {}) {
  return {
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
  };
}

function setApproval(overrides: Record<string, unknown> = {}) {
  server.use(
    http.get("*/api/ssh-approvals/:code", () =>
      HttpResponse.json(approvalData(overrides)),
    ),
  );
}

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
    server.use(
      http.get("*/api/namespaces", () =>
        jsonWithTotal([mockNamespace()]),
      ),
      http.get("*/api/auth/token/:tenant", () =>
        HttpResponse.json({ token: "jwt-token" }),
      ),
      http.get("*/api/ssh-approvals/:code", () =>
        HttpResponse.json(approvalData()),
      ),
      http.post(
        "*/api/ssh-approvals/:code/confirm",
        () => new HttpResponse(null, { status: 204 }),
      ),
      http.post(
        "*/api/ssh-approvals/:code/reject",
        () => new HttpResponse(null, { status: 204 }),
      ),
      http.post(
        "*/api/web-terminal/reauth",
        () => new HttpResponse(null, { status: 204 }),
      ),
    );
  });

  it("asks to add the key, and names the account and namespace it lands in", async () => {
    setApproval();

    renderAt("/ssh-identities/new/WXYZ2K7Q");

    expect(
      await screen.findByText(/add this ssh key to your identities/i),
    ).toBeInTheDocument();
    expect(screen.getByText("Admin User")).toBeInTheDocument();
    expect(screen.getByText("my-namespace")).toBeInTheDocument();
    expect(screen.getByText("SHA256:abc")).toBeInTheDocument();
  });

  it("asks to re-authenticate, and says the window covers more than this login", async () => {
    setApproval({ kind: "reauth", reauth_period: 43200 });

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
    setApproval({ kind: "reauth", reauth_period: 0 });

    renderAt("/ssh-identities/confirm/WXYZ2K7Q");

    expect(
      await screen.findByText(/re-authenticate to continue/i),
    ).toBeInTheDocument();
    expect(
      screen.queryByText(/won't ask for the next/i),
    ).not.toBeInTheDocument();
  });

  it("redirects a reauth code opened on the add route", async () => {
    setApproval({ kind: "reauth" });

    renderAt("/ssh-identities/new/WXYZ2K7Q");

    expect(
      await screen.findByText(/re-authenticate to continue/i),
    ).toBeInTheDocument();
  });

  it("redirects an identity code opened on the reauth route", async () => {
    setApproval();

    renderAt("/ssh-identities/confirm/WXYZ2K7Q");

    expect(
      await screen.findByText(/add this ssh key to your identities/i),
    ).toBeInTheDocument();
  });

  it("confirms the request and reports the outcome", async () => {
    setApproval();
    renderAt("/ssh-identities/new/WXYZ2K7Q");

    await userEvent.click(
      await screen.findByRole("button", { name: /add key/i }),
    );

    expect(await screen.findByText("Key added")).toBeInTheDocument();
  });

  it("rejects the request and reports the outcome", async () => {
    setApproval();
    renderAt("/ssh-identities/new/WXYZ2K7Q");

    await userEvent.click(
      await screen.findByRole("button", { name: /reject/i }),
    );

    expect(await screen.findByText("Rejected")).toBeInTheDocument();
  });

  it("reads a 404 as an expired request", async () => {
    server.use(
      http.get("*/api/ssh-approvals/:code", () =>
        HttpResponse.json({}, { status: 404 }),
      ),
    );

    renderAt("/ssh-identities/new/WXYZ2K7Q");

    expect(await screen.findByText("Request expired")).toBeInTheDocument();
  });

  it("keeps the factor out of sight until the login has been reviewed", async () => {
    const user = userEvent.setup();
    setApproval({ kind: "reauth" });

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
    setApproval({ kind: "reauth" });

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
    expect(
      screen.queryByText(/re-authenticated/i),
    ).not.toBeInTheDocument();
  });

  it("proves the password and reports the login released", async () => {
    const user = userEvent.setup();
    setApproval({ kind: "reauth" });

    renderAt("/ssh-identities/confirm/WXYZ2K7Q");
    await screen.findByText(/re-authenticate to continue/i);
    await user.click(screen.getByRole("button", { name: /continue/i }));

    await user.type(
      await screen.findByLabelText(/account password/i),
      "hunter2",
    );
    await user.click(screen.getByRole("button", { name: /re-authenticate/i }));

    expect(await screen.findByText(/re-authenticated/i)).toBeInTheDocument();
  });

  it("leaves the add-key flow at a single step", async () => {
    setApproval();

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
