import type { ReactNode } from "react";
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
  default: ({ covered }: { covered?: boolean }) => (
    <nav data-testid="sidebar" data-covered={String(!!covered)} />
  ),
}));

vi.mock("../AdminSidebar", () => ({
  default: () => <nav data-testid="admin-sidebar" />,
}));

vi.mock("../TabStrip", () => ({
  default: ({ trailing }: { trailing?: ReactNode }) => (
    <div data-testid="tab-strip">{trailing}</div>
  ),
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

function renderLayout(path = "/") {
  return render(
    <ClipboardProvider>
      <AppLayout />
    </ClipboardProvider>,
    { wrapper: createTestWrapper({ initialEntries: [path] }) },
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

    it("gives way to the admin navigation on admin routes", async () => {
      renderLayout("/admin/dashboard");
      expect(await screen.findByTestId("admin-sidebar")).toBeInTheDocument();
      expect(screen.queryByTestId("sidebar")).not.toBeInTheDocument();
    });
  });

  describe("on the account pages", () => {
    beforeEach(() => {
      server.use(
        http.get("*/api/namespaces", () => jsonWithTotal([mockNamespace()])),
      );
    });

    it("covers the namespace navigation with the page frame", async () => {
      renderLayout("/account/profile");
      expect(await screen.findByTestId("sidebar")).toHaveAttribute(
        "data-covered",
        "true",
      );
    });

    it("moves the account menu beside the tabs", async () => {
      renderLayout("/account/profile");
      expect(
        await screen.findByRole("button", { name: /account menu for/i }),
      ).toBeInTheDocument();
    });

    it("leaves the navigation uncovered elsewhere", async () => {
      renderLayout("/dashboard");
      expect(await screen.findByTestId("sidebar")).toHaveAttribute(
        "data-covered",
        "false",
      );
      expect(
        screen.queryByRole("button", { name: /account menu for/i }),
      ).not.toBeInTheDocument();
    });
  });

  describe("tab strip", () => {
    it("renders regardless of namespaces", async () => {
      renderLayout();
      expect(await screen.findByTestId("tab-strip")).toBeInTheDocument();
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
        await screen.findByTestId("tab-strip");
        expect(
          screen.queryByTestId("device-limit-banner"),
        ).not.toBeInTheDocument();
        expect(screen.queryByTestId("license-banner")).not.toBeInTheDocument();
      },
    );
  });
});
