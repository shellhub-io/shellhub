import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import {
  MemoryRouter,
  Navigate,
  Routes,
  Route,
  useLocation,
} from "react-router-dom";

vi.mock("../admin/License", () => ({
  default: () => <div data-testid="admin-license-page">Admin License</div>,
}));

import { getConfig, isCloud, defaultConfig } from "@/env";
import AdminLicense from "../admin/License";

const mockedGetConfig = vi.mocked(getConfig);

/** Captures the current pathname so we can assert post-redirect location. */
function LocationDisplay() {
  const location = useLocation();
  return <div data-testid="pathname">{location.pathname}</div>;
}

function renderRoute() {
  return render(
    <MemoryRouter initialEntries={["/admin/license"]}>
      <Routes>
        <Route
          path="/admin/license"
          element={
            isCloud() ? (
              <Navigate to="/admin/dashboard" replace />
            ) : (
              <AdminLicense />
            )
          }
        />
        <Route path="/admin/dashboard" element={<LocationDisplay />} />
      </Routes>
    </MemoryRouter>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("AdminLicenseRoute — /admin/license element", () => {
  describe("Cloud Edition", () => {
    beforeEach(() => {
      mockedGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
    });

    it("redirects to /admin/dashboard in cloud mode", () => {
      renderRoute();

      expect(screen.getByTestId("pathname")).toHaveTextContent(
        "/admin/dashboard",
      );
    });
  });

  describe("Enterprise Edition", () => {
    beforeEach(() => {
      mockedGetConfig.mockReturnValue({ ...defaultConfig, edition: "enterprise" });
    });

    it("renders the AdminLicense page when not in cloud mode", () => {
      renderRoute();

      expect(screen.getByTestId("admin-license-page")).toBeInTheDocument();
    });
  });
});
