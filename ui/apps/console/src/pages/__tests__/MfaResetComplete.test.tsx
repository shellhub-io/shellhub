import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import MfaResetComplete from "../MfaResetComplete";

function renderPage() {
  return render(
    <MemoryRouter initialEntries={["/reset-mfa?id=user-1"]}>
      <Routes>
        <Route path="/reset-mfa" element={<MfaResetComplete />} />
        <Route path="/dashboard" element={<div>Dashboard</div>} />
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
    await user.click(
      screen.getByRole("button", { name: "Reset MFA and Login" }),
    );

    expect(
      await screen.findByText(
        "Invalid verification codes. Please check and try again.",
      ),
    ).toBeInTheDocument();
    expect(screen.queryByText("Dashboard")).not.toBeInTheDocument();
  });
});
