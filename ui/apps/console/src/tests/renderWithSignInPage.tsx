import type { ReactNode } from "react";
import { render } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import Login from "@/pages/Login";
import { createTestWrapper } from "./wrapper";

/**
 * Renders routes, starting at path, beside the real sign-in page on /login, for a test whose flow
 * ends there. The sign-in page loads the instance info, so a test that reaches it serves `/info`.
 */
export function renderWithSignInPage(routes: ReactNode, path = "/") {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        {routes}
        <Route path="/login" element={<Login />} />
      </Routes>
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}
