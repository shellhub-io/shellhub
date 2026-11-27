import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import {
  MemoryRouter,
  Route,
  Routes,
  useLocation,
  useNavigate,
} from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { useAuthStore } from "@/stores/authStore";
import { createTestWrapper } from "@/tests/wrapper";
import MfaResetVerify from "../MfaResetVerify";

function LoginNotice() {
  const { state } = useLocation() as {
    state: { notice?: string; signOut?: boolean } | null;
  };
  const navigate = useNavigate();
  return (
    <div>
      Login{state?.signOut ? " after sign-out" : ""}: {state?.notice}
      <button type="button" onClick={() => void navigate(-1)}>
        Back
      </button>
    </div>
  );
}

function renderVerify(state?: { token?: string; identifier?: string }) {
  return render(
    <MemoryRouter
      initialEntries={["/mfa-recover", { pathname: "/mfa-reset-verify", state }]}
    >
      <Routes>
        <Route path="/mfa-reset-verify" element={<MfaResetVerify />} />
        <Route path="/mfa-recover" element={<div>Recover Page</div>} />
        <Route path="/login" element={<LoginNotice />} />
      </Routes>
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

async function fillCode(
  user: ReturnType<typeof userEvent.setup>,
  field: string,
  code: string,
) {
  for (const [i, char] of [...code].entries()) {
    await user.type(
      screen.getByLabelText(`${field} character ${i + 1} of ${code.length}`),
      char,
    );
  }
}

async function submitCodes(user: ReturnType<typeof userEvent.setup>) {
  await fillCode(user, "Main Email Code", "ABCDE");
  await fillCode(user, "Recovery Email Code", "FGHIJ");
  await user.click(
    screen.getByRole("button", { name: /verify and reset mfa/i }),
  );
}

beforeEach(() => {
  useAuthStore.setState({ mfaToken: "mfa-tok" });
});

describe("MfaResetVerify", () => {
  it("asks for the codes sent for the identifier the request carried", () => {
    renderVerify({ token: "reset-token", identifier: "admin" });

    expect(
      screen.getByRole("heading", { name: /enter verification codes/i }),
    ).toBeInTheDocument();
    expect(screen.getByText(/email addresses for admin/i)).toBeInTheDocument();
  });

  it.each([
    ["no request at all", undefined],
    ["a token without its identifier", { token: "reset-token" }],
  ])("sends the user to the recovery page with %s", async (_, state) => {
    renderVerify(state);

    expect(await screen.findByText("Recover Page")).toBeInTheDocument();
  });

  it("resets MFA with the token, drops the sign-in token and sends the user to sign in", async () => {
    let resetUserId: string | undefined;
    server.use(
      http.put("*/api/user/mfa/reset/:userId", ({ params }) => {
        resetUserId = params.userId as string;
        return HttpResponse.json({});
      }),
    );
    const user = userEvent.setup();
    renderVerify({ token: "reset-token", identifier: "admin" });

    await submitCodes(user);

    expect(
      await screen.findByText(
        "Login after sign-out: MFA is off. Sign in with your password.",
      ),
    ).toBeInTheDocument();
    expect(resetUserId).toBe("reset-token");
    expect(useAuthStore.getState().mfaToken).toBeNull();
  });

  it("does not reopen the used reset when the user goes back from sign-in", async () => {
    server.use(
      http.put("*/api/user/mfa/reset/:userId", () => HttpResponse.json({})),
    );
    const user = userEvent.setup();
    renderVerify({ token: "reset-token", identifier: "admin" });
    await submitCodes(user);
    await screen.findByText(/^Login after sign-out/);

    await user.click(screen.getByRole("button", { name: "Back" }));

    expect(await screen.findByText("Recover Page")).toBeInTheDocument();
  });

  it("says the codes are invalid and clears them when the server refuses the reset", async () => {
    server.use(
      http.put("*/api/user/mfa/reset/:userId", () =>
        HttpResponse.json({}, { status: 400 }),
      ),
    );
    const user = userEvent.setup();
    renderVerify({ token: "reset-token", identifier: "admin" });

    await submitCodes(user);

    expect(
      await screen.findByText(/invalid verification codes/i),
    ).toBeInTheDocument();
    await waitFor(() =>
      expect(
        screen.getByLabelText("Main Email Code character 1 of 5"),
      ).toHaveValue(""),
    );
  });
});
