import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Routes, Route, useLocation } from "react-router-dom";
import { setEdition } from "@/tests/edition";
import { ADMIN_INSTANCE_PATH } from "@/utils/adminRoute";
import EnterpriseAdminRouteGuard from "../EnterpriseAdminRouteGuard";

function LocationDisplay() {
  const location = useLocation();
  return <div data-testid="pathname">{location.pathname}</div>;
}

function renderRoute() {
  return render(
    <MemoryRouter initialEntries={["/admin/enterprise-page"]}>
      <Routes>
        <Route element={<EnterpriseAdminRouteGuard />}>
          <Route
            path="/admin/enterprise-page"
            element={<div data-testid="enterprise-page" />}
          />
        </Route>
        <Route path={ADMIN_INSTANCE_PATH} element={<LocationDisplay />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("EnterpriseAdminRouteGuard", () => {
  it("redirects to the instance overview on cloud", () => {
    setEdition("cloud");

    renderRoute();

    expect(screen.getByTestId("pathname")).toHaveTextContent(
      ADMIN_INSTANCE_PATH,
    );
    expect(screen.queryByTestId("enterprise-page")).not.toBeInTheDocument();
  });

  it("renders the page on enterprise", () => {
    setEdition("enterprise");

    renderRoute();

    expect(screen.getByTestId("enterprise-page")).toBeInTheDocument();
  });
});
