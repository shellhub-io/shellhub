// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";

afterEach(cleanup);
import { MemoryRouter, Route, Routes } from "react-router-dom";

import { getConfig, defaultConfig, Edition } from "@/env";
import SignUpGuard from "../SignUpGuard";

const mockedGetConfig = vi.mocked(getConfig);

function renderWithRouter(edition: Edition) {
  mockedGetConfig.mockReturnValue({
    ...defaultConfig,
    edition,
  });
  return render(
    <MemoryRouter initialEntries={["/sign-up"]}>
      <Routes>
        <Route element={<SignUpGuard />}>
          <Route path="/sign-up" element={<div>sign up</div>} />
        </Route>
        <Route path="/login" element={<div>login</div>} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("SignUpGuard", () => {
  it("renders child route when edition is cloud", () => {
    renderWithRouter("cloud");
    expect(screen.getByText("sign up")).toBeInTheDocument();
  });

  it.each(["community", "enterprise"] as const)("redirects to /login when edition is %s", (edition) => {
    renderWithRouter(edition);
    expect(screen.queryByText("sign up")).not.toBeInTheDocument();
    expect(screen.getByText("login")).toBeInTheDocument();
  });
});
