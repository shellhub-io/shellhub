import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import AdminSessions from "../Sessions";
import { makeSdkError, paginatedResponse } from "@/tests/sdk";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSession } from "@/tests/factories";
import { LocationProbe } from "@/tests/LocationProbe";
import { useAuthStore } from "@/stores/authStore";

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getSessionsAdmin: vi.fn(),
  }),
);

function renderPage(initialEntries: string[] = ["/"]) {
  let lastSearch = "";
  const result = render(
    <MemoryRouter initialEntries={initialEntries}>
      <AdminSessions />
      <LocationProbe
        onLocation={(s) => {
          lastSearch = s;
        }}
      />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
  return { ...result, getSearch: () => lastSearch };
}

beforeEach(() => {
  vi.clearAllMocks();
  useAuthStore.setState({ isAdmin: true });
  sdk.getSessionsAdmin.mockResolvedValue(paginatedResponse([]));
});

describe("AdminSessions", () => {
  describe("loading state", () => {
    it("shows a loading spinner while fetching", () => {
      sdk.getSessionsAdmin.mockReturnValue(new Promise(() => {}));
      renderPage();
      expect(screen.getByText(/loading sessions/i)).toBeInTheDocument();
    });

    it("does not render session rows while loading", () => {
      sdk.getSessionsAdmin.mockReturnValue(new Promise(() => {}));
      renderPage();
      expect(screen.queryByText("root")).not.toBeInTheDocument();
    });
  });

  describe("empty state", () => {
    it("shows 'No sessions found' when there are no sessions", async () => {
      renderPage();
      expect(await screen.findByText("No sessions found")).toBeInTheDocument();
    });
  });

  describe("error state", () => {
    it("renders the error banner with role='alert'", async () => {
      sdk.getSessionsAdmin.mockRejectedValue(makeSdkError(500));
      renderPage();
      expect(await screen.findByRole("alert")).toBeInTheDocument();
    });

    it("displays the console's own copy for the status in the banner", async () => {
      sdk.getSessionsAdmin.mockRejectedValue(makeSdkError(403));
      renderPage();
      const alert = await screen.findByRole("alert");
      expect(alert).toHaveTextContent("You do not have permission to do this.");
    });

    it("does not show the error banner when there is no error", async () => {
      renderPage();
      await screen.findByText("No sessions found");
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });

  describe("session rows", () => {
    it("renders one row per session", async () => {
      sdk.getSessionsAdmin.mockResolvedValue(
        paginatedResponse([
          mockSession({ uid: "session-1", username: "root" }),
          mockSession({ uid: "session-2", username: "admin" }),
        ]),
      );
      renderPage();
      expect(await screen.findByText("root")).toBeInTheDocument();
      expect(screen.getByText("admin")).toBeInTheDocument();
    });

    it("renders the device name via DeviceChip", async () => {
      sdk.getSessionsAdmin.mockResolvedValue(
        paginatedResponse([mockSession()]),
      );
      renderPage();
      expect(await screen.findByText("my-device")).toBeInTheDocument();
    });

    it("renders the truncated session uid", async () => {
      sdk.getSessionsAdmin.mockResolvedValue(
        paginatedResponse([mockSession({ uid: "abcdef1234567890" })]),
      );
      renderPage();
      expect(await screen.findByText("abcdef1234")).toBeInTheDocument();
    });

    it("renders the IP address", async () => {
      sdk.getSessionsAdmin.mockResolvedValue(
        paginatedResponse([mockSession({ ip_address: "10.0.0.1" })]),
      );
      renderPage();
      expect(await screen.findByText("10.0.0.1")).toBeInTheDocument();
    });

    it("navigates to session detail when a row is clicked", async () => {
      const user = userEvent.setup();
      sdk.getSessionsAdmin.mockResolvedValue(
        paginatedResponse([mockSession({ uid: "session-abc" })]),
      );
      renderPage();

      await user.click(await screen.findByText("root"));

      expect(mockNavigate).toHaveBeenCalledWith("/admin/sessions/session-abc");
    });
  });

  describe("active indicator", () => {
    it("renders a green dot for active sessions", async () => {
      sdk.getSessionsAdmin.mockResolvedValue(
        paginatedResponse([mockSession({ active: true })]),
      );
      renderPage();
      await screen.findByText("root");
      const dot = document.querySelector(".bg-accent-green");
      expect(dot).toBeInTheDocument();
    });

    it("renders a muted dot for inactive sessions", async () => {
      sdk.getSessionsAdmin.mockResolvedValue(
        paginatedResponse([mockSession({ active: false })]),
      );
      renderPage();
      await screen.findByText("root");
      const dot = document.querySelector(".bg-text-muted\\/40");
      expect(dot).toBeInTheDocument();
    });
  });

  describe("authentication indicator", () => {
    it("renders the 'Authenticated' shield for authenticated sessions", async () => {
      sdk.getSessionsAdmin.mockResolvedValue(
        paginatedResponse([mockSession({ authenticated: true })]),
      );
      renderPage();
      expect(await screen.findByTitle("Authenticated")).toBeInTheDocument();
    });

    it("renders the 'Not authenticated' shield for unauthenticated sessions", async () => {
      sdk.getSessionsAdmin.mockResolvedValue(
        paginatedResponse([mockSession({ authenticated: false })]),
      );
      renderPage();
      await screen.findByText("root");
      expect(screen.getAllByTitle("Not authenticated").length).toBeGreaterThan(
        0,
      );
    });

    it("shows the warning icon in the username cell for unauthenticated sessions", async () => {
      sdk.getSessionsAdmin.mockResolvedValue(
        paginatedResponse([mockSession({ authenticated: false })]),
      );
      renderPage();
      await screen.findByText("root");
      expect(
        screen.getAllByTitle("Not authenticated").length,
      ).toBeGreaterThanOrEqual(2);
    });
  });

  describe("device fallback", () => {
    it("shows the truncated device_uid when device object is missing", async () => {
      sdk.getSessionsAdmin.mockResolvedValue(
        paginatedResponse([
          mockSession({ device: null, device_uid: "abcd1234efgh" }),
        ]),
      );
      renderPage();
      expect(await screen.findByText("abcd1234")).toBeInTheDocument();
    });
  });

  describe("pagination", () => {
    it("renders pagination when totalCount > perPage", async () => {
      sdk.getSessionsAdmin.mockResolvedValue(
        paginatedResponse(
          Array.from({ length: 10 }, (_, i) =>
            mockSession({ uid: `session-${i}`, username: `user-${i}` }),
          ),
          25,
        ),
      );
      renderPage();
      expect(await screen.findByText(/25/)).toBeInTheDocument();
    });
  });

  describe("URL hydration", () => {
    it("passes page=3 to the SDK when URL has ?page=3", async () => {
      renderPage(["/?page=3"]);
      await waitFor(() => {
        expect(sdk.getSessionsAdmin).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({ page: 3 }),
          }),
        );
      });
    });

    it("passes page=1 to the SDK when URL has no page param", async () => {
      renderPage(["/"]);
      await waitFor(() => {
        expect(sdk.getSessionsAdmin).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({ page: 1 }),
          }),
        );
      });
    });
  });

  describe("URL writes", () => {
    it("writes ?page=2 to the URL when the user clicks Next page", async () => {
      const user = userEvent.setup();
      sdk.getSessionsAdmin.mockResolvedValue(
        paginatedResponse(
          Array.from({ length: 10 }, (_, i) =>
            mockSession({ uid: `s-${i}`, username: `u-${i}` }),
          ),
          30,
        ),
      );
      const { getSearch } = renderPage();

      await screen.findByText("u-0");

      await user.click(screen.getByRole("button", { name: "Next page" }));

      await waitFor(() => {
        const sp = new URLSearchParams(getSearch());
        expect(sp.get("page")).toBe("2");
      });
    });

    it("omits ?page from the URL when on the default page 1", () => {
      const { getSearch } = renderPage(["/"]);
      const sp = new URLSearchParams(getSearch());
      expect(sp.get("page")).toBeNull();
    });
  });
});
