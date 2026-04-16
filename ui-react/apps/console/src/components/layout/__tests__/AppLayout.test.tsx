import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import AppLayout from "../AppLayout";

const mockUseNamespaces =
  vi.fn<
    () => {
      namespaces: Array<{ tenant_id: string; name: string }>;
      isLoading: boolean;
      error: Error | null;
      refetch: () => void;
    }
  >();

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

vi.mock("@/common/ConnectivityBanner", () => ({
  default: () => null,
}));

afterEach(cleanup);

beforeEach(() => {
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
        <AppLayout />
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
});
