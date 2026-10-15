import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import { setEdition } from "@/tests/edition";
import AdminRoute from "../AdminRoute";

beforeEach(() => {
  useAuthStore.setState({ isAdmin: false });
});

function renderAdminRoute() {
  return render(
    <MemoryRouter initialEntries={["/admin/instance"]}>
      <Routes>
        <Route element={<AdminRoute />}>
          <Route path="/admin/instance" element={<div>admin content</div>} />
        </Route>
        <Route path="/" element={<div>home page</div>} />
        <Route
          path="/admin/unauthorized"
          element={<div>unauthorized page</div>}
        />
      </Routes>
    </MemoryRouter>,
  );
}

describe("AdminRoute", () => {
  describe.each(["enterprise", "cloud"] as const)("on %s", (edition) => {
    beforeEach(() => {
      setEdition(edition);
    });

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

  it.each([true, false])(
    "redirects to the home page on community when isAdmin is %s",
    (isAdmin) => {
      setEdition("community");
      useAuthStore.setState({ isAdmin });
      renderAdminRoute();

      expect(screen.getByText("home page")).toBeInTheDocument();
      expect(screen.queryByText("admin content")).not.toBeInTheDocument();
      expect(screen.queryByText("unauthorized page")).not.toBeInTheDocument();
    },
  );
});
