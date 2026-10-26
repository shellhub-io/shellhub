import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes, useLocation } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { useAuthStore } from "@/stores/authStore";
import { useSignUpStore } from "@/stores/signUpStore";
import { mockUserAuth } from "@/tests/factories";
import ConfirmAccount from "../ConfirmAccount";

function RouteState() {
  const { search, state } = useLocation();
  return (
    <>
      <output aria-label="Route state">{JSON.stringify(state)}</output>
      <output aria-label="Route search">{search}</output>
    </>
  );
}

function renderConfirmAccount(search = "", state?: unknown) {
  const user = userEvent.setup(
    vi.isFakeTimers() ? { advanceTimers: vi.advanceTimersByTime } : {},
  );
  render(
    <MemoryRouter
      initialEntries={[{ pathname: "/confirm-account", search, state }]}
    >
      <Routes>
        <Route
          path="/confirm-account"
          element={
            <>
              <ConfirmAccount />
              <RouteState />
            </>
          }
        />
        <Route path="/login" element={<div>Login Page</div>} />
        <Route path="/dashboard" element={<div>Dashboard</div>} />
      </Routes>
    </MemoryRouter>,
  );
  return user;
}

async function submitCode(
  user: ReturnType<typeof userEvent.setup>,
  code: string,
) {
  await user.click(screen.getAllByRole("textbox")[0]);
  await user.paste(code);
  await user.click(screen.getByRole("button", { name: "Confirm" }));
}

function respondToCode(status: number, headers: Record<string, string> = {}) {
  server.use(
    http.post("*/api/user/confirm_account", () =>
      HttpResponse.json({}, { status, headers }),
    ),
  );
}

function confirmWith(sent: unknown[]) {
  server.use(
    http.post("*/api/user/confirm_account", async ({ request }) => {
      sent.push(await request.json());
      return HttpResponse.json(mockUserAuth({ token: "session-token" }));
    }),
  );
}

function resendWith(sent: unknown[]) {
  server.use(
    http.post("*/api/user/resend_email", async ({ request }) => {
      sent.push(await request.json());
      return new HttpResponse(null, { status: 200 });
    }),
  );
}

beforeEach(() => {
  useAuthStore.setState(useAuthStore.getInitialState(), true);
  useSignUpStore.setState({ resendLoading: false, resendError: null });
  server.use(
    http.post(
      "*/api/user/resend_email",
      () => new HttpResponse(null, { status: 200 }),
    ),
  );
});

describe("ConfirmAccount", () => {
  it("sends to the login without an identifier", async () => {
    renderConfirmAccount();

    expect(await screen.findByText("Login Page")).toBeInTheDocument();
  });

  it("names the address the code went to", () => {
    renderConfirmAccount("?identifier=john%40example.com");

    expect(screen.getByText(/john@example\.com/)).toBeInTheDocument();
  });

  it("signs in and opens the console on the right code", async () => {
    const sent: unknown[] = [];
    confirmWith(sent);

    const user = renderConfirmAccount("?identifier=john_doe");
    await submitCode(user, "042917");

    expect(await screen.findByText("Dashboard")).toBeInTheDocument();
    expect(sent).toEqual([{ identifier: "john_doe", code: "042917" }]);
    expect(useAuthStore.getState().token).toBe("session-token");
  });

  it("submits the code a link carries without typing it", async () => {
    const sent: unknown[] = [];
    confirmWith(sent);

    renderConfirmAccount("?identifier=john%40example.com&code=042917");

    expect(await screen.findByText("Dashboard")).toBeInTheDocument();
    expect(sent).toEqual([{ identifier: "john@example.com", code: "042917" }]);
  });

  it("sends an account that must sign in with its password to the login", async () => {
    server.use(
      http.post(
        "*/api/user/confirm_account",
        () => new HttpResponse(null, { status: 204 }),
      ),
    );

    const user = renderConfirmAccount("?identifier=john_doe");
    await submitCode(user, "042917");

    expect(await screen.findByText("Login Page")).toBeInTheDocument();
    expect(useAuthStore.getState().token).toBeNull();
  });

  it("keeps the code and says so when the server fails", async () => {
    respondToCode(500);

    const user = renderConfirmAccount("?identifier=john_doe");
    await submitCode(user, "042917");

    expect(await screen.findByText(/something went wrong/i)).toBeInTheDocument();
    expect(screen.queryByText(/code is wrong or expired/i)).not.toBeInTheDocument();
    expect(screen.getAllByRole("textbox")[0]).toHaveValue("0");
  });

  it("drops a linked code that failed from the URL, so a reload does not guess it again", async () => {
    respondToCode(401);

    renderConfirmAccount("?identifier=john%40example.com&code=042917");

    expect(
      await screen.findByText(/code is wrong or expired/i),
    ).toBeInTheDocument();
    expect(screen.getByRole("status", { name: "Route search" })).toHaveTextContent(
      "?identifier=john%40example.com",
    );
    expect(
      screen.getByRole("status", { name: "Route search" }).textContent,
    ).not.toContain("code=");
  });

  it("clears the cells and says so on a wrong code", async () => {
    respondToCode(401);

    const user = renderConfirmAccount("?identifier=john_doe");
    await submitCode(user, "042917");

    expect(
      await screen.findByText(/code is wrong or expired/i),
    ).toBeInTheDocument();
    for (const input of screen.getAllByRole("textbox")) {
      expect(input).toHaveValue("");
    }
  });

  describe("lockout", () => {
    beforeEach(() => {
      vi.useFakeTimers({ shouldAdvanceTime: true });
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it("shows the remaining time and refuses another code until it ends", async () => {
      const epoch = Math.floor(Date.now() / 1000) + 30;
      respondToCode(429, { "x-account-lockout": String(epoch) });

      const user = renderConfirmAccount("?identifier=john_doe");
      await submitCode(user, "042917");

      expect(
        await screen.findByText(/too many wrong codes/i),
      ).toBeInTheDocument();
      await act(() => vi.advanceTimersByTimeAsync(1000));
      expect(screen.getByText(/\(2\d seconds\)/)).toBeInTheDocument();

      await user.click(screen.getAllByRole("textbox")[0]);
      await user.paste("042917");
      expect(screen.getByRole("button", { name: "Confirm" })).toBeDisabled();

      await act(() => vi.advanceTimersByTimeAsync(30_000));
      expect(screen.getByText(/enter the code again/i)).toBeInTheDocument();
      expect(screen.getByRole("button", { name: "Confirm" })).toBeEnabled();
    });
  });

  describe("notices that go stale", () => {
    it("drops the wrong-code message once a resend goes out", async () => {
      respondToCode(401);

      const user = renderConfirmAccount("?identifier=john_doe");
      await submitCode(user, "042917");
      expect(
        await screen.findByText(/code is wrong or expired/i),
      ).toBeInTheDocument();

      await user.click(screen.getByRole("button", { name: /resend email/i }));

      expect(
        await screen.findByText(/we sent the code again/i),
      ).toBeInTheDocument();
      expect(
        screen.queryByText(/code is wrong or expired/i),
      ).not.toBeInTheDocument();
    });

    it("keeps the lockout message when a resend goes out during it", async () => {
      const epoch = Math.floor(Date.now() / 1000) + 30;
      respondToCode(429, { "x-account-lockout": String(epoch) });

      const user = renderConfirmAccount("?identifier=john_doe");
      await submitCode(user, "042917");
      await screen.findByText(/too many wrong codes/i);
      await user.click(screen.getByRole("button", { name: /resend email/i }));

      expect(
        await screen.findByText(/we sent the code again/i),
      ).toBeInTheDocument();
      expect(screen.getByText(/too many wrong codes/i)).toBeInTheDocument();
    });

    it("keeps a lockout that a code tripped while a resend was on its way", async () => {
      const epoch = Math.floor(Date.now() / 1000) + 30;
      respondToCode(429, { "x-account-lockout": String(epoch) });
      let answerResend!: () => void;
      server.use(
        http.post(
          "*/api/user/resend_email",
          () =>
            new Promise<Response>((resolve) => {
              answerResend = () =>
                resolve(new HttpResponse(null, { status: 200 }));
            }),
        ),
      );

      const user = renderConfirmAccount("?identifier=john_doe");
      await user.click(screen.getByRole("button", { name: /resend email/i }));
      await submitCode(user, "042917");
      await screen.findByText(/too many wrong codes/i);
      answerResend();

      expect(
        await screen.findByText(/we sent the code again/i),
      ).toBeInTheDocument();
      expect(screen.getByText(/too many wrong codes/i)).toBeInTheDocument();
    });

    it("keeps the wrong-code message when the resend fails", async () => {
      respondToCode(401);
      server.use(
        http.post("*/api/user/resend_email", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );

      const user = renderConfirmAccount("?identifier=john_doe");
      await submitCode(user, "042917");
      await screen.findByText(/code is wrong or expired/i);
      await user.click(screen.getByRole("button", { name: /resend email/i }));

      expect(
        await screen.findByText(/failed to resend email/i),
      ).toBeInTheDocument();
      expect(screen.getByText(/code is wrong or expired/i)).toBeInTheDocument();
    });

    it("drops the resend notice once a code is submitted", async () => {
      respondToCode(401);

      const user = renderConfirmAccount("?identifier=john_doe");
      await user.click(screen.getByRole("button", { name: /resend email/i }));
      await screen.findByText(/we sent the code again/i);

      await submitCode(user, "042917");

      expect(
        await screen.findByText(/code is wrong or expired/i),
      ).toBeInTheDocument();
      expect(
        screen.queryByText(/we sent the code again/i),
      ).not.toBeInTheDocument();
    });
  });

  describe("resend", () => {
    it("asks for the email by the identifier and counts the cooldown down", async () => {
      const sent: unknown[] = [];
      resendWith(sent);

      const user = renderConfirmAccount("?identifier=john_doe");
      await user.click(screen.getByRole("button", { name: /resend email/i }));

      expect(
        await screen.findByText(/we sent the code again/i),
      ).toBeInTheDocument();
      expect(sent).toEqual([{ identifier: "john_doe" }]);
      expect(
        screen.getByRole("button", { name: /resend email \(\d+s\)/i }),
      ).toBeDisabled();
    });

    it("sends a fresh code once on arrival from the sign-in, and consumes the request so a reload does not resend", async () => {
      const sent: unknown[] = [];
      resendWith(sent);

      renderConfirmAccount("?identifier=john_doe", { sendCode: true });

      expect(
        await screen.findByRole("button", { name: /resend email \(\d+s\)/i }),
      ).toBeDisabled();
      expect(sent).toEqual([{ identifier: "john_doe" }]);
      expect(screen.getByRole("status", { name: "Route state" })).toHaveTextContent(
        "null",
      );
    });

    it("shows no error when the arrival send is throttled, and waits out the cooldown", async () => {
      server.use(
        http.post("*/api/user/resend_email", () =>
          HttpResponse.json({}, { status: 429 }),
        ),
      );

      renderConfirmAccount("?identifier=john_doe", { sendCode: true });

      expect(
        await screen.findByRole("button", { name: /resend email \(\d+s\)/i }),
      ).toBeDisabled();
      expect(screen.queryByText(/wait a minute/i)).not.toBeInTheDocument();
    });

    it("sends nothing on arrival from the sign-up, which already sent the code", async () => {
      const sent: unknown[] = [];
      resendWith(sent);

      renderConfirmAccount("?identifier=john%40example.com");

      expect(
        await screen.findByRole("button", { name: "Resend Email" }),
      ).toBeEnabled();
      expect(sent).toEqual([]);
    });

    it.each([
      { status: 429, message: /wait a minute/i },
      { status: 500, message: /failed to resend email/i },
    ])(
      "shows $message when the resend answers $status",
      async ({ status, message }) => {
        server.use(
          http.post("*/api/user/resend_email", () =>
            HttpResponse.json({}, { status }),
          ),
        );

        const user = renderConfirmAccount("?identifier=john_doe");
        await user.click(screen.getByRole("button", { name: /resend email/i }));

        expect(await screen.findByText(message)).toBeInTheDocument();
      },
    );
  });
});
