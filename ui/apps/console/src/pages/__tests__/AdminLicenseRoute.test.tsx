import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import {
  MemoryRouter,
  Navigate,
  Routes,
  Route,
  useLocation,
} from "react-router-dom";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

vi.mock("@/env", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/env")>();
  return { ...actual, getConfig: vi.fn(() => actual.getConfig()) };
});

// Stub AdminLicense — its real implementation pulls in many hooks/components
// that are irrelevant to the routing behaviour under test.
vi.mock("../admin/License", () => ({
  default: () => <div data-testid="admin-license-page">Admin License</div>,
}));

import { getConfig, defaultConfig } from "@/env";
import AdminLicense from "../admin/License";

const mockedGetConfig = vi.mocked(getConfig);

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

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
            getConfig().cloud ? (
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

/* ------------------------------------------------------------------ */
/* Setup / teardown                                                    */
/* ------------------------------------------------------------------ */

afterEach(cleanup);

beforeEach(() => {
  vi.clearAllMocks();
});

/* ================================================================== */
/* Tests                                                               */
/* ================================================================== */

describe("AdminLicenseRoute — /admin/license element", () => {
  describe("cloud=true", () => {
    beforeEach(() => {
      mockedGetConfig.mockReturnValue({ ...defaultConfig, cloud: true });
    });

    it("redirects to /admin/dashboard in cloud mode", () => {
      renderRoute();

      // The Navigate component should redirect the router to /admin/dashboard,
      // which renders LocationDisplay. Asserting on its text content confirms
      // that the redirect happened.
      expect(screen.getByTestId("pathname")).toHaveTextContent(
        "/admin/dashboard",
      );
    });
  });

  describe("cloud=false (enterprise / self-hosted)", () => {
    beforeEach(() => {
      mockedGetConfig.mockReturnValue({ ...defaultConfig, cloud: false });
    });

    it("renders the AdminLicense page when not in cloud mode", () => {
      renderRoute();

      expect(screen.getByTestId("admin-license-page")).toBeInTheDocument();
    });
  });
});
