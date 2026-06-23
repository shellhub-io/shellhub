import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import AppLayout from "../AppLayout";
import { ClipboardProvider } from "@/components/common/ClipboardProvider";
import { getConfig, defaultConfig } from "@/env";

const mockGetConfig = vi.mocked(getConfig);

const mockUseNamespaces = vi.fn<
  () => {
    namespaces: Array<{ tenant_id: string; name: string }>;
    isLoading: boolean;
    error: Error | null;
    refetch: () => void;
  }
>();

vi.mock("@/env", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/env")>();
  return { ...actual, getConfig: vi.fn() };
});

vi.mock("@/hooks/useNamespaces", () => ({
  useNamespaces: () => mockUseNamespaces(),
  useNamespace: () => ({ tenant_id: "t1", name: "ns1" }),
  useInitRole: () => {},
}));

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

afterEach(cleanup);

beforeEach(() => {
  mockGetConfig.mockReturnValue({ ...defaultConfig });
  mockUseNamespaces.mockReturnValue({
    namespaces: [],
    isLoading: false,
    error: null,
    refetch: vi.fn(),
  });
});

function renderLayout() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <ClipboardProvider>
          <AppLayout />
        </ClipboardProvider>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("AppLayout", () => {
  describe("Sidebar", () => {
    it("renders when namespaces exist", () => {
      mockUseNamespaces.mockReturnValue({
        namespaces: [{ tenant_id: "t1", name: "ns1" }],
        isLoading: false,
        error: null,
        refetch: vi.fn(),
      });
      renderLayout();
      expect(screen.getByTestId("sidebar")).toBeInTheDocument();
    });

    it("is hidden when there are no namespaces", () => {
      renderLayout();
      expect(screen.queryByTestId("sidebar")).not.toBeInTheDocument();
    });
  });

  describe("AppBar", () => {
    it("renders regardless of namespaces", () => {
      renderLayout();
      expect(screen.getByTestId("app-bar")).toBeInTheDocument();
    });

    it("renders alongside the sidebar when namespaces exist", () => {
      mockUseNamespaces.mockReturnValue({
        namespaces: [{ tenant_id: "t1", name: "ns1" }],
        isLoading: false,
        error: null,
        refetch: vi.fn(),
      });
      renderLayout();
      expect(screen.getByTestId("app-bar")).toBeInTheDocument();
      expect(screen.getByTestId("sidebar")).toBeInTheDocument();
    });
  });

  describe("ConnectivityBanner", () => {
    it("is always mounted", () => {
      renderLayout();
      expect(screen.getByTestId("connectivity-banner")).toBeInTheDocument();
    });
  });

  describe("skip link", () => {
    it("renders the skip link pointing at main content", () => {
      renderLayout();
      const link = screen.getByRole("link", { name: /skip to main content/i });
      expect(link).toBeInTheDocument();
      expect(link).toHaveAttribute("href", "#main-content");
    });

    it("exposes the main landmark with id and tabindex", () => {
      renderLayout();
      const main = screen.getByRole("main");
      expect(main).toHaveAttribute("id", "main-content");
      expect(main).toHaveAttribute("tabindex", "-1");
    });

    it("renders the skip link before the main content in the DOM", () => {
      renderLayout();
      const link = screen.getByRole("link", { name: /skip to main content/i });
      const main = screen.getByRole("main");
      expect(
        link.compareDocumentPosition(main) & Node.DOCUMENT_POSITION_FOLLOWING,
      ).toBeTruthy();
    });

    it("renders the skip link when the sidebar is visible", () => {
      mockUseNamespaces.mockReturnValue({
        namespaces: [{ tenant_id: "t1", name: "ns1" }],
        isLoading: false,
        error: null,
        refetch: vi.fn(),
      });
      renderLayout();
      expect(
        screen.getByRole("link", { name: /skip to main content/i }),
      ).toBeInTheDocument();
    });
  });

  describe("enterprise banners", () => {
    it("mounts LicenseBanner and DeviceLimitBanner when enterprise and not cloud", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        enterprise: true,
        cloud: false,
      });
      renderLayout();
      expect(screen.getByTestId("device-limit-banner")).toBeInTheDocument();
      expect(screen.getByTestId("license-banner")).toBeInTheDocument();
    });

    it("does not mount the enterprise banners on a community instance", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        enterprise: false,
        cloud: false,
      });
      renderLayout();
      expect(
        screen.queryByTestId("device-limit-banner"),
      ).not.toBeInTheDocument();
      expect(screen.queryByTestId("license-banner")).not.toBeInTheDocument();
    });

    it("does not mount the enterprise banners when cloud is true (even if enterprise is true)", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        enterprise: true,
        cloud: true,
      });
      renderLayout();
      expect(
        screen.queryByTestId("device-limit-banner"),
      ).not.toBeInTheDocument();
      expect(screen.queryByTestId("license-banner")).not.toBeInTheDocument();
    });
  });
});
