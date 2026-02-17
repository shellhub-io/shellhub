import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { useAuthStore } from "../../../stores/authStore";
import ProtectedRoute from "../ProtectedRoute";

beforeEach(() => {
  useAuthStore.setState({
    token: null,
    user: null,
    userId: null,
    email: null,
    tenant: null,
    role: null,
    name: null,
    loading: false,
    error: null,
  });
});

function renderWithRouter(token: string | null) {
  useAuthStore.setState({ token });

  return render(
    <MemoryRouter initialEntries={["/protected"]}>
      <Routes>
        <Route element={<ProtectedRoute />}>
          <Route path="/protected" element={<div>secret content</div>} />
        </Route>
        <Route path="/login" element={<div>login page</div>} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("ProtectedRoute", () => {
  it("renders child route when token exists", () => {
    renderWithRouter("valid-token");
    expect(screen.getByText("secret content")).toBeInTheDocument();
  });

  it("redirects to /login when no token", () => {
    renderWithRouter(null);
    expect(screen.queryByText("secret content")).not.toBeInTheDocument();
    expect(screen.getByText("login page")).toBeInTheDocument();
  });
});
