import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { useNamespacesStore } from "../../../stores/namespacesStore";
import AppLayout from "../AppLayout";

vi.mock("../Sidebar", () => ({
  default: () => <nav data-testid="sidebar" />,
}));

vi.mock("../AppBar", () => ({
  default: () => <div data-testid="app-bar" />,
}));

vi.mock("../../terminal/TerminalManager", () => ({
  default: () => null,
}));

vi.mock("../../common/ConnectivityBanner", () => ({
  default: () => null,
}));

afterEach(cleanup);

beforeEach(() => {
  useNamespacesStore.setState({
    namespaces: [],
    currentNamespace: null,
    loading: false,
    loaded: true,
    error: null,
  });
});

function renderLayout() {
  return render(
    <MemoryRouter>
      <AppLayout />
    </MemoryRouter>,
  );
}

describe("AppLayout", () => {
  describe("Sidebar", () => {
    it("renders when namespaces exist", () => {
      useNamespacesStore.setState({
        namespaces: [{ tenant_id: "t1", name: "ns1" }] as never,
        loaded: true,
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
      useNamespacesStore.setState({
        namespaces: [{ tenant_id: "t1", name: "ns1" }] as never,
        loaded: true,
      });
      renderLayout();
      expect(screen.getByTestId("app-bar")).toBeInTheDocument();
      expect(screen.getByTestId("sidebar")).toBeInTheDocument();
    });
  });
});
