import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { useAuthStore } from "@/stores/authStore";
import { createTestWrapper } from "@/tests/wrapper";
import { mockUserAuth } from "@/tests/factories";
import MfaRecover from "../MfaRecover";

function renderRecover() {
  return render(
    <MemoryRouter>
      <MfaRecover />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

function serveRecovery(expiresAt?: number) {
  const bodies: unknown[] = [];
  server.use(
    http.post("*/api/user/mfa/recover", async ({ request }) => {
      bodies.push(await request.json());
      return HttpResponse.json(mockUserAuth(), {
        headers: expiresAt ? { "x-expires-at": String(expiresAt) } : {},
      });
    }),
  );
  return bodies;
}

function refuseRecovery() {
  server.use(
    http.post("*/api/user/mfa/recover", () =>
      HttpResponse.json({}, { status: 403 }),
    ),
  );
}

async function typeRecoveryCode(code: string, user = userEvent.setup()) {
  await user.type(screen.getByPlaceholderText(/recovery code/i), code);
  return user;
}

beforeEach(() => {
  vi.clearAllMocks();
  server.use(
    http.put(
      "*/api/user/mfa/recovery/disable",
      () => new HttpResponse(null, { status: 204 }),
    ),
  );
  useAuthStore.setState(useAuthStore.getInitialState(), true);
  useAuthStore.setState({ user: "admin", mfaToken: "mfa-temp-token" });
});

describe("MfaRecover", () => {
  it("submits recovery code successfully", async () => {
    const bodies = serveRecovery();

    renderRecover();

    const user = userEvent.setup();
    await user.type(
      screen.getByPlaceholderText(/recovery code/i),
      "ABC-123-XYZ",
    );
    await user.click(screen.getByRole("button", { name: /recover/i }));

    await waitFor(() => {
      expect(bodies).toEqual([
        { token: "mfa-temp-token", recovery_code: "ABC-123-XYZ" },
      ]);
    });
  });

  it("displays error message on invalid code", async () => {
    refuseRecovery();

    renderRecover();

    const user = userEvent.setup();
    await user.type(screen.getByPlaceholderText(/recovery code/i), "BAD-CODE");
    await user.click(screen.getByRole("button", { name: /recover/i }));

    await waitFor(() => {
      expect(screen.getByText("Invalid recovery code")).toBeInTheDocument();
    });
  });

  it("disables submit button when input is empty", () => {
    renderRecover();

    expect(screen.getByRole("button", { name: /recover/i })).toBeDisabled();
  });

  it("enables submit button after typing a valid code", async () => {
    renderRecover();

    await typeRecoveryCode("ABC-123-XYZ");

    expect(screen.getByRole("button", { name: /recover/i })).not.toBeDisabled();
  });

  it("shows loading state during submission", async () => {
    server.use(http.post("*/api/user/mfa/recover", () => new Promise(() => {})));
    renderRecover();

    const user = await typeRecoveryCode("ABC-123-XYZ");
    await user.click(screen.getByRole("button", { name: /recover/i }));

    expect(await screen.findByText(/Recovering.../i)).toBeInTheDocument();
  });

  it("has link back to MFA login", () => {
    renderRecover();

    expect(screen.getByText(/Back to verification/i)).toHaveAttribute(
      "href",
      "/mfa-login",
    );
  });

  it("shows timeout modal after successful recovery", async () => {
    serveRecovery(Math.floor(Date.now() / 1000) + 300);

    renderRecover();

    const user = userEvent.setup();
    await user.type(
      screen.getByPlaceholderText(/recovery code/i),
      "ABC-123-XYZ",
    );
    await user.click(screen.getByRole("button", { name: /recover/i }));

    expect(
      await screen.findByText(/Recovery Window Active/i),
    ).toBeInTheDocument();
  });

  it("redirects to login without the password step's MFA token", async () => {
    useAuthStore.setState({ mfaToken: null, token: null });

    render(
      <MemoryRouter initialEntries={["/recover-mfa"]}>
        <Routes>
          <Route path="/recover-mfa" element={<MfaRecover />} />
          <Route path="/login" element={<div>Login Page</div>} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createTestWrapper() },
    );

    await waitFor(() => {
      expect(screen.getByText("Login Page")).toBeInTheDocument();
    });
  });

  it("clears the recovery code field after a failed submission", async () => {
    refuseRecovery();

    renderRecover();

    const user = userEvent.setup();
    const input = screen.getByPlaceholderText(/recovery code/i);
    await user.type(input, "BAD-CODE");
    await user.click(screen.getByRole("button", { name: /recover/i }));

    await waitFor(() => {
      expect(input).toHaveValue("");
    });
  });

  it("clears the recovery code field after a successful submission", async () => {
    serveRecovery();

    renderRecover();

    const user = userEvent.setup();
    const input = screen.getByPlaceholderText(/recovery code/i);
    await user.type(input, "ABC-123-XYZ");
    await user.click(screen.getByRole("button", { name: /recover account/i }));

    await waitFor(() => {
      expect(input).toHaveValue("");
    });
  });

  it("calls the disable MFA endpoint during recovery window", async () => {
    const disableCalled = vi.fn();
    server.use(
      http.put("*/api/user/mfa/recovery/disable", () => {
        disableCalled();
        return new HttpResponse(null, { status: 204 });
      }),
    );
    serveRecovery(Math.floor(Date.now() / 1000) + 600);

    render(
      <MemoryRouter initialEntries={["/recover-mfa"]}>
        <Routes>
          <Route path="/recover-mfa" element={<MfaRecover />} />
          <Route path="/dashboard" element={<div>Dashboard</div>} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createTestWrapper() },
    );

    const user = userEvent.setup();
    await user.type(
      screen.getByPlaceholderText(/recovery code/i),
      "MY-RECOVERY-CODE",
    );
    await user.click(screen.getByRole("button", { name: /recover account/i }));

    await waitFor(() => {
      expect(screen.getByText(/Recovery Window Active/i)).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: /disable mfa/i }));

    await waitFor(() => {
      expect(disableCalled).toHaveBeenCalled();
    });
  });

  it("navigates to dashboard and updates MFA status after successful disable", async () => {
    serveRecovery(Math.floor(Date.now() / 1000) + 600);

    render(
      <MemoryRouter initialEntries={["/recover-mfa"]}>
        <Routes>
          <Route path="/recover-mfa" element={<MfaRecover />} />
          <Route path="/dashboard" element={<div>Dashboard</div>} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createTestWrapper() },
    );

    const user = userEvent.setup();
    await user.type(
      screen.getByPlaceholderText(/recovery code/i),
      "MY-RECOVERY-CODE",
    );
    await user.click(screen.getByRole("button", { name: /recover account/i }));

    await waitFor(() => {
      expect(screen.getByText(/Recovery Window Active/i)).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: /disable mfa/i }));

    expect(await screen.findByText("Dashboard")).toBeInTheDocument();
    expect(useAuthStore.getState().mfaEnabled).toBe(false);
  });
});
