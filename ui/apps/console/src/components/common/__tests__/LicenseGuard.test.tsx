import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import type { ComponentType, ReactNode } from "react";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { getConfig, defaultConfig } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import LicenseGuard from "../LicenseGuard";

const mockGetConfig = vi.mocked(getConfig);

function renderGuard(Wrapper?: ComponentType<{ children: ReactNode }>) {
  const ui = (
    <MemoryRouter initialEntries={["/admin/instance"]}>
      <Routes>
        <Route element={<LicenseGuard />}>
          <Route
            path="/admin/instance"
            element={<div>protected content</div>}
          />
        </Route>
        <Route path="/admin/license" element={<div>license page</div>} />
      </Routes>
    </MemoryRouter>
  );
  return render(Wrapper ? <Wrapper>{ui}</Wrapper> : ui);
}

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "enterprise" });
  useAuthStore.setState({ isAdmin: true });
});

describe("LicenseGuard", () => {
  describe("isLoading — shows PageLoader", () => {
    it("renders a loading indicator while the license check is in progress", () => {
      server.use(http.get("*/admin/api/license", () => new Promise(() => {})));
      renderGuard(createTestWrapper());
      expect(screen.getByText("Checking license...")).toBeInTheDocument();
      expect(screen.queryByText("protected content")).not.toBeInTheDocument();
    });

    it("does not render the Outlet while loading", () => {
      server.use(http.get("*/admin/api/license", () => new Promise(() => {})));
      renderGuard(createTestWrapper());
      expect(screen.queryByText("protected content")).not.toBeInTheDocument();
    });
  });

  describe("isError — redirects to /admin/license", () => {
    it("navigates to the license page when the query errors", async () => {
      server.use(
        http.get("*/admin/api/license", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      renderGuard(createTestWrapper());
      await waitFor(() => {
        expect(screen.getByText("license page")).toBeInTheDocument();
      });
      expect(screen.queryByText("protected content")).not.toBeInTheDocument();
    });
  });

  describe("outside enterprise", () => {
    it.each(["cloud", "community"] as const)(
      "renders the Outlet without calling getLicense on %s",
      async (edition) => {
        mockGetConfig.mockReturnValue({ ...defaultConfig, edition });
        let called = false;
        server.use(
          http.get("*/admin/api/license", () => {
            called = true;
            return HttpResponse.json({}, { status: 400 });
          }),
        );

        renderGuard(createTestWrapper());

        await waitFor(() => {
          expect(screen.getByText("protected content")).toBeInTheDocument();
        });
        expect(called).toBe(false);
      },
    );
  });
});
