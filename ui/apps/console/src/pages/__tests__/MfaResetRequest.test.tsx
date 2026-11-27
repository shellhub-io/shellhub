import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes, useLocation } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { useAuthStore } from "@/stores/authStore";
import { createTestWrapper } from "@/tests/wrapper";
import MfaResetRequest from "../MfaResetRequest";

function VerifyPage() {
  const state = useLocation().state as {
    token?: string;
    identifier?: string;
  } | null;
  return (
    <dl aria-label="Verify Page">
      <dt>Token</dt>
      <dd>{state?.token}</dd>
      <dt>Identifier</dt>
      <dd>{state?.identifier}</dd>
    </dl>
  );
}

function renderPage() {
  return render(
    <MemoryRouter initialEntries={["/mfa-reset-request"]}>
      <Routes>
        <Route path="/mfa-reset-request" element={<MfaResetRequest />} />
        <Route path="/mfa-reset-verify" element={<VerifyPage />} />
        <Route path="/login" element={<div>Login Page</div>} />
      </Routes>
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

function serveResetRequest(response: () => Response) {
  const bodies: unknown[] = [];
  server.use(
    http.post("*/api/user/mfa/reset", async ({ request }) => {
      bodies.push(await request.json());
      return response();
    }),
  );
  return bodies;
}

async function sendCodes() {
  const user = userEvent.setup();
  await user.click(
    screen.getByRole("button", { name: /send verification codes/i }),
  );
}

beforeEach(() => {
  useAuthStore.setState({ user: "admin", username: null, mfaToken: "tok" });
});

describe("MfaResetRequest", () => {
  it("asks for the reset by the identifier and carries its token to the verify page", async () => {
    const bodies = serveResetRequest(() =>
      HttpResponse.json({ token: "reset-token" }),
    );
    renderPage();

    await sendCodes();

    const verify = await screen.findByLabelText("Verify Page");
    expect(within(verify).getByText("reset-token")).toBeInTheDocument();
    expect(within(verify).getByText("admin")).toBeInTheDocument();
    expect(bodies).toEqual([{ identifier: "admin" }]);
  });

  it.each([
    [500, /unable to send reset emails/i],
    [429, /already requested in the last minute/i],
  ])("stays on the page and explains a %i", async (status, message) => {
    serveResetRequest(() => HttpResponse.json({}, { status }));
    renderPage();

    await sendCodes();

    expect(await screen.findByText(message)).toBeInTheDocument();
    expect(screen.queryByLabelText("Verify Page")).not.toBeInTheDocument();
  });

  it("redirects to login when there is no identifier and no mfaToken", async () => {
    useAuthStore.setState({ user: null, username: null, mfaToken: null });

    renderPage();

    expect(await screen.findByText("Login Page")).toBeInTheDocument();
  });

  it("offers no reset request when there is no identifier but an active mfaToken", () => {
    useAuthStore.setState({
      user: null,
      username: null,
      mfaToken: "active-tok",
    });

    renderPage();

    expect(
      screen.queryByRole("button", { name: /send verification codes/i }),
    ).not.toBeInTheDocument();
  });
});
