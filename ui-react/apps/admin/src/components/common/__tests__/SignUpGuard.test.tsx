// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";

afterEach(cleanup);
import { MemoryRouter, Route, Routes } from "react-router-dom";

vi.mock("../../../env", () => ({ getConfig: vi.fn() }));
import { getConfig } from "../../../env";
import SignUpGuard from "../SignUpGuard";

const mockedGetConfig = vi.mocked(getConfig);

function renderWithRouter(cloud: boolean) {
  mockedGetConfig.mockReturnValue({ cloud } as ReturnType<typeof getConfig>);
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
  it("renders child route when cloud is true", () => {
    renderWithRouter(true);
    expect(screen.getByText("sign up")).toBeInTheDocument();
  });

  it("redirects to /login when cloud is false", () => {
    renderWithRouter(false);
    expect(screen.queryByText("sign up")).not.toBeInTheDocument();
    expect(screen.getByText("login")).toBeInTheDocument();
  });
});
