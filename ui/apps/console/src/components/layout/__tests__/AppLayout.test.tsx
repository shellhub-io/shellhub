import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { ClipboardProvider } from "@/components/common/ClipboardProvider";
import { getConfig, defaultConfig } from "@/env";
import AppLayout from "../AppLayout";

vi.mock("@/hooks/useSidebarLayout", () => ({
  useSidebarLayout: () => ({
    expanded: false,
    pinned: false,
    isOpen: false,
    isDesktop: true,
    drawerOpen: false,
    handlers: {
      onMouseEnter: vi.fn(),
      onMouseLeave: vi.fn(),
      onFocus: vi.fn(),
      onBlur: vi.fn(),
      onToggle: vi.fn(),
      openDrawer: vi.fn(),
      closeDrawer: vi.fn(),
      onDrawerKeyDown: vi.fn(),
    },
  }),
}));

vi.mock("../Sidebar", () => ({
  default: () => <nav data-testid="sidebar" />,
}));

vi.mock("../AppBar", () => ({
  default: () => <div data-testid="app-bar" />,
}));

vi.mock("@/terminal/TerminalManager", () => ({
  default: () => null,
}));

vi.mock("@/components/common/ConnectivityBanner", () => ({
  default: () => <div data-testid="connectivity-banner" />,
}));

vi.mock("@/components/common/DeviceLimitBanner", () => ({
  default: () => <div data-testid="device-limit-banner" />,
}));

vi.mock("@/components/common/LicenseBanner", () => ({
  default: () => <div data-testid="license-banner" />,
}));

const mockGetConfig = vi.mocked(getConfig);

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig });
  seedAuthStore();
  server.use(
    http.get("*/api/namespaces", () => jsonWithTotal([])),
    http.get("*/api/namespaces/:tenant", () => HttpResponse.json(null)),
    http.get("*/api/auth/token/:tenant", () =>
      HttpResponse.json({ token: "jwt-token", role: "owner" }),
    ),
  );
});

function renderLayout() {
  return render(
    <ClipboardProvider>
      <AppLayout />
    </ClipboardProvider>,
    { wrapper: createTestWrapper({ initialEntries: ["/"] }) },
  );
}

describe("AppLayout", () => {
  describe("Sidebar", () => {
    it("renders when namespaces exist", async () => {
      server.use(
        http.get("*/api/namespaces", () => jsonWithTotal([mockNamespace()])),
      );
      renderLayout();
      expect(await screen.findByTestId("sidebar")).toBeInTheDocument();
    });

    it("is hidden when there are no namespaces", async () => {
      renderLayout();
      await waitFor(() => {
        expect(screen.queryByTestId("sidebar")).not.toBeInTheDocument();
      });
    });
  });

  describe("AppBar", () => {
    it("renders regardless of namespaces", async () => {
      renderLayout();
      expect(await screen.findByTestId("app-bar")).toBeInTheDocument();
    });
  });

  describe("skip link", () => {
    it("points at the main landmark, which is focusable", async () => {
      renderLayout();
      const link = await screen.findByRole("link", {
        name: /skip to main content/i,
      });
      expect(link).toHaveAttribute("href", "#main-content");

      const main = screen.getByRole("main");
      expect(main).toHaveAttribute("id", "main-content");
      expect(main).toHaveAttribute("tabindex", "-1");
    });

    it("renders the skip link before the main content in the DOM", async () => {
      renderLayout();
      const link = await screen.findByRole("link", {
        name: /skip to main content/i,
      });
      const main = screen.getByRole("main");
      expect(
        link.compareDocumentPosition(main) & Node.DOCUMENT_POSITION_FOLLOWING,
      ).toBeTruthy();
    });
  });

  describe("enterprise banners", () => {
    it("mounts LicenseBanner and DeviceLimitBanner in an enterprise instance", async () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
      renderLayout();
      expect(
        await screen.findByTestId("device-limit-banner"),
      ).toBeInTheDocument();
      expect(screen.getByTestId("license-banner")).toBeInTheDocument();
    });

    it.each(["community", "cloud"] as const)(
      "does not mount the enterprise banners on a %s instance",
      async (edition) => {
        mockGetConfig.mockReturnValue({ ...defaultConfig, edition });
        renderLayout();
        await screen.findByTestId("app-bar");
        expect(
          screen.queryByTestId("device-limit-banner"),
        ).not.toBeInTheDocument();
        expect(screen.queryByTestId("license-banner")).not.toBeInTheDocument();
      },
    );
  });
});
