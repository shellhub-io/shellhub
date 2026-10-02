import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { useAuthStore } from "@/stores/authStore";
import AdminDashboard from "../Dashboard";

const fullStats = {
  registered_users: 42,
  registered_devices: 150,
  online_devices: 75,
  active_sessions: 12,
  pending_devices: 5,
  rejected_devices: 3,
};

function renderPage() {
  return render(
    <MemoryRouter>
      <AdminDashboard />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

beforeEach(() => {
  useAuthStore.setState({ isAdmin: true });
  server.use(http.get("*/admin/api/stats", () => HttpResponse.json(fullStats)));
});

describe("AdminDashboard", () => {
  describe("loading state", () => {
    it("renders spinner with role='status'", () => {
      server.use(http.get("*/admin/api/stats", () => new Promise(() => {})));
      renderPage();
      expect(screen.getByRole("status")).toBeInTheDocument();
    });
  });

  describe("error state", () => {
    it("renders error message with role='alert'", async () => {
      server.use(
        http.get("*/admin/api/stats", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      renderPage();
      await waitFor(() => {
        expect(screen.getByRole("alert")).toHaveTextContent(
          "Failed to load dashboard statistics",
        );
      });
      expect(screen.queryByText("Registered Users")).not.toBeInTheDocument();
    });
  });

  describe("success state, all fields present", () => {
    it("renders correct numeric values for each stat", async () => {
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("42")).toBeInTheDocument();
      });
      expect(screen.getByText("150")).toBeInTheDocument();
      expect(screen.getByText("75")).toBeInTheDocument();
      expect(screen.getByText("12")).toBeInTheDocument();
      expect(screen.getByText("5")).toBeInTheDocument();
      expect(screen.getByText("3")).toBeInTheDocument();
    });

    it("links the users count to the users page", async () => {
      renderPage();
      await waitFor(() => {
        expect(
          screen.getByRole("link", { name: /view all users/i }),
        ).toHaveAttribute("href", "/admin/users");
      });
    });

    it("links neither the device nor the session counts", async () => {
      renderPage();
      await screen.findByText("Registered Devices");
      expect(
        screen.queryByRole("link", { name: /devices|sessions/i }),
      ).not.toBeInTheDocument();
    });
  });

  describe("success state, partial stats response", () => {
    it("renders 0 for each missing stat field", async () => {
      server.use(
        http.get("*/admin/api/stats", () =>
          HttpResponse.json({ registered_users: 10 }),
        ),
      );
      renderPage();
      await waitFor(() => {
        expect(screen.getByText("10")).toBeInTheDocument();
      });
      const zeros = screen.getAllByText("0");
      expect(zeros.length).toBeGreaterThanOrEqual(5);
    });
  });
});
