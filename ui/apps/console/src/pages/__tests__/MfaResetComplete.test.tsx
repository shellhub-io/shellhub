import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes, useLocation } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { useAuthStore } from "@/stores/authStore";
import MfaResetComplete from "../MfaResetComplete";

function LoginNotice() {
  const { state } = useLocation() as {
    state: { notice?: string; signOut?: boolean } | null;
  };
  return (
    <div>
      Login{state?.signOut ? " after sign-out" : ""}: {state?.notice}
    </div>
  );
}

function renderPage() {
  return render(
    <MemoryRouter initialEntries={["/reset-mfa?id=user-1"]}>
      <Routes>
        <Route path="/reset-mfa" element={<MfaResetComplete />} />
        <Route path="/login" element={<LoginNotice />} />
      </Routes>
    </MemoryRouter>,
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

describe("MfaResetComplete", () => {
  it("says the codes are invalid when the server refuses the reset", async () => {
    server.use(
      http.put("*/api/user/mfa/reset/:userId", () =>
        HttpResponse.json({}, { status: 403 }),
      ),
    );
    const user = userEvent.setup();
    renderPage();

    await fillCode(user, "Main Email Code", "ABCDE");
    await fillCode(user, "Recovery Email Code", "FGHIJ");
    await user.click(screen.getByRole("button", { name: "Reset MFA" }));

    expect(
      await screen.findByText(
        "Invalid verification codes. Please check and try again.",
      ),
    ).toBeInTheDocument();
    expect(screen.queryByText(/^Login/)).not.toBeInTheDocument();
  });

  it("resets MFA and sends the user to sign in with the password", async () => {
    server.use(
      http.put(
        "*/api/user/mfa/reset/:userId",
        () => new HttpResponse(null, { status: 200 }),
      ),
    );
    useAuthStore.setState({ mfaToken: "pending-mfa-token" });
    const user = userEvent.setup();
    renderPage();

    await fillCode(user, "Main Email Code", "ABCDE");
    await fillCode(user, "Recovery Email Code", "FGHIJ");
    await user.click(screen.getByRole("button", { name: "Reset MFA" }));

    expect(
      await screen.findByText(
        "Login after sign-out: MFA is off. Sign in with your password.",
      ),
    ).toBeInTheDocument();
    expect(useAuthStore.getState().mfaToken).toBeNull();
  });
});
