import { afterEach, describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { Route, Routes } from "react-router-dom";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import {
  PENDING_DEVICE_CODE_KEY,
  setPendingDeviceCode,
} from "@/utils/navigation";
import SsoLink from "../SsoLink";

function renderPage() {
  return render(
    <Routes>
      <Route path="/sso-link" element={<SsoLink />} />
      <Route path="/dashboard" element={<div>dashboard</div>} />
      <Route path="/accept-device" element={<div>accept device</div>} />
    </Routes>,
    { wrapper: createTestWrapper({ initialEntries: ["/sso-link"] }) },
  );
}

function answer(status: number) {
  let calls = 0;
  server.use(
    http.post("*/api/user/saml/link", () => {
      calls += 1;
      return status === 200
        ? new HttpResponse(null, { status })
        : HttpResponse.json({ message: "error" }, { status });
    }),
  );
  return () => calls;
}

describe("SsoLink", () => {
  afterEach(() => {
    localStorage.removeItem(PENDING_DEVICE_CODE_KEY);
  });

  it("links the account and continues where the sign-in was going", async () => {
    const user = userEvent.setup();
    const calls = answer(200);
    setPendingDeviceCode("WXYZ2K7Q");
    renderPage();

    await user.click(screen.getByRole("button", { name: /link account/i }));

    expect(
      await screen.findByRole("heading", { name: /account linked/i }),
    ).toBeInTheDocument();
    expect(calls()).toBe(1);

    await user.click(screen.getByRole("button", { name: /continue/i }));

    expect(await screen.findByText("accept device")).toBeInTheDocument();
  });

  it("says no link is waiting when the server refuses it", async () => {
    const user = userEvent.setup();
    answer(403);
    renderPage();

    await user.click(screen.getByRole("button", { name: /link account/i }));

    expect(
      await screen.findByText(/no single sign-on link waiting/i),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /link account/i })).toBeEnabled();
  });

  it("offers a retry when the link fails for another reason", async () => {
    const user = userEvent.setup();
    answer(500);
    renderPage();

    await user.click(screen.getByRole("button", { name: /link account/i }));

    expect(
      await screen.findByText(/failed to link this account/i),
    ).toBeInTheDocument();
  });

  it("lets the user leave without linking", async () => {
    const user = userEvent.setup();
    const calls = answer(200);
    renderPage();

    await user.click(screen.getByRole("button", { name: /not now/i }));

    expect(await screen.findByText("dashboard")).toBeInTheDocument();
    expect(calls()).toBe(0);
  });

  it("still acts on a pending device code when the user leaves without linking", async () => {
    const user = userEvent.setup();
    setPendingDeviceCode("WXYZ2K7Q");
    renderPage();

    await user.click(screen.getByRole("button", { name: /not now/i }));

    expect(await screen.findByText("accept device")).toBeInTheDocument();
  });
});
