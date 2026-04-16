import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import AdminRoute from "../AdminRoute";

// Mock the API client so fetchUser never makes a real HTTP request.
// authStore.fetchUser calls getUserInfo from @/client (re-exported from ../client).
vi.mock("@/client", async (importOriginal) => {
  const original = await importOriginal<typeof import("@/client")>();
  return {
    ...original,
    getUserInfo: vi.fn(),
  };
});

import { getUserInfo } from "@/client";

const mockGetUserInfo = vi.mocked(getUserInfo);

const baseState = {
  token: null,
  user: null,
  userId: null,
  email: null,
  username: null,
  recoveryEmail: null,
  tenant: null,
  role: null,
  isAdmin: false,
  name: null,
  loading: false,
  error: null,
  mfaEnabled: false,
  mfaToken: null,
  mfaRecoveryExpiry: null,
  mfaResetUserId: null,
  mfaResetIdentifier: null,
};

beforeEach(() => {
  useAuthStore.setState(baseState);
  vi.clearAllMocks();
});

function renderAdminRoute() {
  return render(
    <MemoryRouter initialEntries={["/admin/dashboard"]}>
      <Routes>
        <Route element={<AdminRoute />}>
          <Route path="/admin/dashboard" element={<div>admin content</div>} />
        </Route>
        <Route path="/admin/unauthorized" element={<div>unauthorized page</div>} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("AdminRoute", () => {
  describe("loading state", () => {
    it("renders a spinner while verifying admin status", async () => {
      // fetchUser never settles during this assertion window
      let resolveFetch!: () => void;
      mockGetUserInfo.mockReturnValue(
        new Promise((resolve) => {
          resolveFetch = () => resolve({ data: { admin: true } } as never);
        }) as never,
      );

      renderAdminRoute();

      // The spinner is the only element in the loading branch; confirm admin
      // content is not yet visible.
      expect(screen.queryByText("admin content")).not.toBeInTheDocument();
      expect(screen.queryByText("unauthorized page")).not.toBeInTheDocument();

      // Unblock so the component can finish, avoiding act() warnings.
      resolveFetch();
      await waitFor(() =>
        expect(screen.queryByText("admin content")).toBeInTheDocument(),
      );
    });
  });

  describe("admin user", () => {
    it("renders the Outlet when fetchUser resolves and isAdmin is true", async () => {
      mockGetUserInfo.mockResolvedValue({
        data: { admin: true, user: "root", id: "1", email: "a@b.com", tenant: "t", name: "Root" },
      } as never);

      renderAdminRoute();

      await waitFor(() =>
        expect(screen.getByText("admin content")).toBeInTheDocument(),
      );
      expect(screen.queryByText("unauthorized page")).not.toBeInTheDocument();
    });

    it("does not redirect to /admin/unauthorized when the user is admin", async () => {
      mockGetUserInfo.mockResolvedValue({
        data: { admin: true, user: "root", id: "1", email: "a@b.com", tenant: "t", name: "Root" },
      } as never);

      renderAdminRoute();

      await waitFor(() =>
        expect(screen.getByText("admin content")).toBeInTheDocument(),
      );
    });
  });

  describe("non-admin user", () => {
    it("redirects to /admin/unauthorized when fetchUser resolves and isAdmin is false", async () => {
      mockGetUserInfo.mockResolvedValue({
        data: { admin: false, user: "user1", id: "2", email: "u@b.com", tenant: "t", name: "User" },
      } as never);

      renderAdminRoute();

      await waitFor(() =>
        expect(screen.getByText("unauthorized page")).toBeInTheDocument(),
      );
      expect(screen.queryByText("admin content")).not.toBeInTheDocument();
    });
  });

  describe("fetchUser failure", () => {
    it("redirects to /admin/unauthorized when fetchUser rejects", async () => {
      // authStore.fetchUser catches the error and sets isAdmin: false
      mockGetUserInfo.mockRejectedValue(new Error("network error"));

      renderAdminRoute();

      await waitFor(() =>
        expect(screen.getByText("unauthorized page")).toBeInTheDocument(),
      );
      expect(screen.queryByText("admin content")).not.toBeInTheDocument();
    });
  });
});
