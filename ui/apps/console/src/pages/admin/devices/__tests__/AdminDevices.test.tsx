import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import AdminDevices from "../index";
import { createTestWrapper } from "@/tests/wrapper";
import { mockDevice } from "@/tests/factories";
import { useAuthStore } from "@/stores/authStore";

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

let lastRequestUrl: URL | null;

function setDevices(
  devices: ReturnType<typeof mockDevice>[],
  total?: number,
) {
  server.use(
    http.get("*/admin/api/devices", ({ request }) => {
      lastRequestUrl = new URL(request.url);
      return jsonWithTotal(devices, total ?? devices.length);
    }),
  );
}

function renderPage(initialEntries: string[] = ["/"]) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <AdminDevices />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

describe("AdminDevices", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    lastRequestUrl = null;
    useAuthStore.setState({ isAdmin: true });
    setDevices([]);
  });

  describe("rendering", () => {
    it("renders the page heading", () => {
      renderPage();
      expect(
        screen.getByRole("heading", { name: "Devices" }),
      ).toBeInTheDocument();
    });

    it("renders the search input with correct aria-label", () => {
      renderPage();
      expect(
        screen.getByRole("searchbox", { name: "Search devices by hostname" }),
      ).toBeInTheDocument();
    });

    it("renders all status filter tabs", () => {
      renderPage();
      expect(screen.getByRole("tab", { name: "All" })).toBeInTheDocument();
      expect(screen.getByRole("tab", { name: "Accepted" })).toBeInTheDocument();
      expect(screen.getByRole("tab", { name: "Pending" })).toBeInTheDocument();
      expect(screen.getByRole("tab", { name: "Rejected" })).toBeInTheDocument();
    });
  });

  describe("loading state", () => {
    it('renders the loading spinner with "Loading devices..." text', () => {
      server.use(
        http.get("*/admin/api/devices", () => new Promise(() => {})),
      );
      renderPage();
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText("Loading devices...")).toBeInTheDocument();
    });
  });

  describe("empty state", () => {
    it('renders "No devices found" when the device list is empty', async () => {
      renderPage();
      expect(await screen.findByText("No devices found")).toBeInTheDocument();
    });
  });

  describe("device rows", () => {
    it("renders a row for each returned device", async () => {
      setDevices([
        mockDevice({ uid: "uid-1", name: "device-alpha" }),
        mockDevice({ uid: "uid-2", name: "device-beta" }),
      ]);
      renderPage();
      expect(await screen.findByText("device-alpha")).toBeInTheDocument();
      expect(screen.getByText("device-beta")).toBeInTheDocument();
    });

    it("renders the status chip for each device", async () => {
      setDevices([mockDevice({ status: "pending" })]);
      renderPage();
      await screen.findByText("my-device");
      expect(screen.getAllByText("Pending").length).toBeGreaterThanOrEqual(2);
    });

    it("navigates to the device detail page when a row is clicked", async () => {
      const user = userEvent.setup();
      setDevices([mockDevice({ uid: "uid-abc", name: "clickable-device" })]);
      renderPage();

      await user.click(await screen.findByText("clickable-device"));
      expect(mockNavigate).toHaveBeenCalledWith("/admin/devices/uid-abc");
    });
  });

  describe("error state", () => {
    it("renders an error alert when the API returns an error", async () => {
      server.use(
        http.get("*/admin/api/devices", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      renderPage();
      expect(await screen.findByRole("alert")).toBeInTheDocument();
      expect(
        screen.getByText("Something went wrong on our side. Try again."),
      ).toBeInTheDocument();
    });
  });

  describe("status tab interaction", () => {
    it("re-renders without crashing after clicking a status tab", async () => {
      const user = userEvent.setup();
      renderPage();
      await user.click(screen.getByRole("tab", { name: "Accepted" }));
      expect(
        screen.getByRole("searchbox", { name: "Search devices by hostname" }),
      ).toBeInTheDocument();
    });
  });

  describe("URL hydration — controls reflect URL params on mount", () => {
    it("passes sortBy/orderBy hydrated from URL to the API", async () => {
      renderPage(["/?sortField=name&sortOrder=asc"]);
      await screen.findByText("No devices found");
      expect(lastRequestUrl!.searchParams.get("sort_by")).toBe("name");
      expect(lastRequestUrl!.searchParams.get("order_by")).toBe("asc");
    });

    it("passes status hydrated from URL to the API", async () => {
      renderPage(["/?status=accepted"]);
      await screen.findByText("No devices found");
      expect(lastRequestUrl!.searchParams.get("status")).toBe("accepted");
    });

    it("marks the matching status tab as selected when status is in the URL", () => {
      renderPage(["/?status=pending"]);
      expect(screen.getByRole("tab", { name: "Pending" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });

    it("passes page hydrated from URL to the API", async () => {
      renderPage(["/?page=3"]);
      await screen.findByText("No devices found");
      expect(lastRequestUrl!.searchParams.get("page")).toBe("3");
    });

    it("uses defaults when URL params are absent (last_seen/desc, page 1, no status)", async () => {
      renderPage(["/"]);
      await screen.findByText("No devices found");
      expect(lastRequestUrl!.searchParams.get("sort_by")).toBe("last_seen");
      expect(lastRequestUrl!.searchParams.get("order_by")).toBe("desc");
      expect(lastRequestUrl!.searchParams.get("page")).toBe("1");
    });

    it("rejects an invalid status value and falls back to no status filter (All tab selected)", async () => {
      renderPage(["/?status=invalid-status"]);
      await screen.findByText("No devices found");
      expect(lastRequestUrl!.searchParams.get("status")).toBeNull();
      expect(screen.getByRole("tab", { name: "All" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });
  });

  describe("URL writes — interactions update URL and reset page", () => {
    it("clicking a status tab writes status to URL and resets page to 1", async () => {
      const user = userEvent.setup();
      renderPage(["/?page=2"]);
      await screen.findByText("No devices found");

      await user.click(screen.getByRole("tab", { name: "Accepted" }));

      await screen.findByText("No devices found");
      expect(lastRequestUrl!.searchParams.get("status")).toBe("accepted");
      expect(lastRequestUrl!.searchParams.get("page")).toBe("1");
    });

    it("clicking a sort column header writes sort to API and resets page", async () => {
      const user = userEvent.setup();
      renderPage(["/?page=3"]);
      await screen.findByText("No devices found");

      await user.click(
        screen.getByRole("button", { name: /sort by hostname/i }),
      );

      await screen.findByText("No devices found");
      expect(lastRequestUrl!.searchParams.get("sort_by")).toBe("name");
      expect(lastRequestUrl!.searchParams.get("order_by")).toBe("asc");
      expect(lastRequestUrl!.searchParams.get("page")).toBe("1");
    });

    it("clicking the same sort column again toggles order from asc to desc", async () => {
      const user = userEvent.setup();
      renderPage(["/?sortField=name&sortOrder=asc"]);
      await screen.findByText("No devices found");

      await user.click(
        screen.getByRole("button", { name: /sort by hostname/i }),
      );

      await screen.findByText("No devices found");
      expect(lastRequestUrl!.searchParams.get("sort_by")).toBe("name");
      expect(lastRequestUrl!.searchParams.get("order_by")).toBe("desc");
    });
  });

  describe("URL writes — default params are omitted from the URL", () => {
    it("API receives page=1 when on the default page", async () => {
      renderPage(["/"]);
      await screen.findByText("No devices found");
      expect(lastRequestUrl!.searchParams.get("page")).toBe("1");
    });

    it("API receives no status when All tab is selected (default)", async () => {
      renderPage(["/"]);
      await screen.findByText("No devices found");
      expect(lastRequestUrl!.searchParams.get("status")).toBeNull();
    });
  });
});
