import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import AdminRoute from "../AdminRoute";

beforeEach(() => {
  useAuthStore.setState({ isAdmin: false });
});

function renderAdminRoute() {
  return render(
    <MemoryRouter initialEntries={["/admin/dashboard"]}>
      <Routes>
        <Route element={<AdminRoute />}>
          <Route path="/admin/dashboard" element={<div>admin content</div>} />
        </Route>
        <Route
          path="/admin/unauthorized"
          element={<div>unauthorized page</div>}
        />
      </Routes>
    </MemoryRouter>,
  );
}

describe("AdminRoute", () => {
  it("renders the Outlet when isAdmin is true", () => {
    useAuthStore.setState({ isAdmin: true });
    renderAdminRoute();

    expect(screen.getByText("admin content")).toBeInTheDocument();
    expect(screen.queryByText("unauthorized page")).not.toBeInTheDocument();
  });

  it("redirects to /admin/unauthorized when isAdmin is false", () => {
    renderAdminRoute();

    expect(screen.getByText("unauthorized page")).toBeInTheDocument();
    expect(screen.queryByText("admin content")).not.toBeInTheDocument();
  });
});
