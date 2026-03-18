import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { useConnectivityStore } from "../../../stores/connectivityStore";
import NamespaceGuard from "../NamespaceGuard";

const mockUseNamespaces = vi.fn<() => { namespaces: Array<{ tenant_id: string; name: string }>; isLoading: boolean; error: Error | null; refetch: () => void }>();

vi.mock("../../../hooks/useNamespaces", () => ({
  useNamespaces: () => mockUseNamespaces(),
  useInitRole: () => {},
}));

vi.mock("../CreateNamespace", () => ({
  default: () => <div data-testid="create-namespace" />,
}));

vi.mock("../../layout/UserMenu", () => ({
  default: () => <div data-testid="user-menu" />,
}));

afterEach(cleanup);

beforeEach(() => {
  mockUseNamespaces.mockReturnValue({
    namespaces: [],
    isLoading: false,
    error: null,
    refetch: vi.fn(),
  });

  useConnectivityStore.getState().markUp();
});

function renderGuard(initialPath = "/dashboard") {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <Routes>
        <Route element={<NamespaceGuard />}>
          <Route path="/dashboard" element={<div>dashboard content</div>} />
          <Route path="/profile" element={<div>profile content</div>} />
        </Route>
      </Routes>
    </MemoryRouter>,
  );
}

describe("NamespaceGuard", () => {
  describe("loading state", () => {
    it("shows a loading spinner while namespaces are not yet loaded", () => {
      mockUseNamespaces.mockReturnValue({
        namespaces: [],
        isLoading: true,
        error: null,
        refetch: vi.fn(),
      });
      renderGuard();
      expect(screen.getByText(/loading/i)).toBeInTheDocument();
    });

    it("does not render the outlet while loading", () => {
      mockUseNamespaces.mockReturnValue({
        namespaces: [],
        isLoading: true,
        error: null,
        refetch: vi.fn(),
      });
      renderGuard();
      expect(screen.queryByText("dashboard content")).not.toBeInTheDocument();
    });
  });

  describe("with namespaces", () => {
    it("renders the outlet when namespaces exist", () => {
      mockUseNamespaces.mockReturnValue({
        namespaces: [{ tenant_id: "t1", name: "ns1" }],
        isLoading: false,
        error: null,
        refetch: vi.fn(),
      });
      renderGuard();
      expect(screen.getByText("dashboard content")).toBeInTheDocument();
    });

    it("does not show the create-namespace screen when namespaces exist", () => {
      mockUseNamespaces.mockReturnValue({
        namespaces: [{ tenant_id: "t1", name: "ns1" }],
        isLoading: false,
        error: null,
        refetch: vi.fn(),
      });
      renderGuard();
      expect(screen.queryByTestId("create-namespace")).not.toBeInTheDocument();
    });
  });

  describe("without namespaces — non-profile route", () => {
    it("shows the create-namespace screen", () => {
      renderGuard("/dashboard");
      expect(screen.getByTestId("create-namespace")).toBeInTheDocument();
    });

    it("does not render the outlet", () => {
      renderGuard("/dashboard");
      expect(screen.queryByText("dashboard content")).not.toBeInTheDocument();
    });

    it("renders UserMenu in the minimal header", () => {
      renderGuard("/dashboard");
      expect(screen.getByTestId("user-menu")).toBeInTheDocument();
    });
  });

  describe("without namespaces — /profile route", () => {
    it("renders the outlet instead of the create-namespace screen", () => {
      renderGuard("/profile");
      expect(screen.getByText("profile content")).toBeInTheDocument();
    });

    it("does not show the create-namespace screen", () => {
      renderGuard("/profile");
      expect(screen.queryByTestId("create-namespace")).not.toBeInTheDocument();
    });

    it("does not show the minimal header", () => {
      renderGuard("/profile");
      expect(screen.queryByTestId("user-menu")).not.toBeInTheDocument();
    });
  });
});
