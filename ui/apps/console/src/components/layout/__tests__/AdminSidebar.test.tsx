import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, fireEvent } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import React from "react";

import { defaultConfig } from "@/env";
import { useAuthStore } from "@/stores/authStore";

vi.mock("@/hooks/useAdminLicense", () => ({
  useAdminLicense: vi.fn(),
}));

vi.mock("../SidebarShell", () => ({
  default: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="sidebar-shell">{children}</div>
  ),
  NavItemLink: ({
    item,
    disabled,
  }: {
    item: { to: string; label: string; icon: React.ReactNode };
    expanded: boolean;
    disabled?: boolean;
  }) =>
    disabled ? (
      <span aria-disabled="true">{item.label}</span>
    ) : (
      <a href={item.to}>{item.label}</a>
    ),
  navBase: "",
  navDisabled: "",
  navIcon: "",
}));

import { getConfig } from "@/env";
import { useAdminLicense } from "@/hooks/useAdminLicense";
import AdminSidebar from "../AdminSidebar";

const mockGetConfig = vi.mocked(getConfig);
const mockUseAdminLicense = vi.mocked(useAdminLicense);

// ── Helpers ───────────────────────────────────────────────────────────────────

function renderSidebar() {
  return render(
    <MemoryRouter>
      <AdminSidebar expanded={true} pinned={true} onToggle={vi.fn()} />
    </MemoryRouter>,
  );
}

function openSettingsGroup() {
  fireEvent.click(screen.getByRole("button", { name: /settings/i }));
}

// ── Tests ─────────────────────────────────────────────────────────────────────

afterEach(cleanup);

describe("AdminSidebar", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ isAdmin: true } as never);
    mockUseAdminLicense.mockReturnValue({
      isLoading: false,
      isExpired: false,
    } as never);
    mockGetConfig.mockReturnValue({ ...defaultConfig });
  });

  describe("cloud admin (cloud=true, isAdmin=true)", () => {
    beforeEach(() => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
    });

    it("shows the core nav entries", () => {
      renderSidebar();
      expect(screen.getByText("Dashboard")).toBeInTheDocument();
      expect(screen.getByText("Users")).toBeInTheDocument();
      expect(screen.getByText("Devices")).toBeInTheDocument();
      expect(screen.getByText("Sessions")).toBeInTheDocument();
      expect(screen.getByText("Firewall Rules")).toBeInTheDocument();
      expect(screen.getByText("Namespaces")).toBeInTheDocument();
    });

    it("shows Authentication but NOT License in the Settings group", () => {
      renderSidebar();
      openSettingsGroup();
      expect(screen.getByText("Authentication")).toBeInTheDocument();
      expect(screen.queryByText("License")).not.toBeInTheDocument();
    });
  });

  describe("enterprise admin with valid license (cloud=false, isExpired=false)", () => {
    it("shows the core nav entries", () => {
      renderSidebar();
      expect(screen.getByText("Dashboard")).toBeInTheDocument();
      expect(screen.getByText("Users")).toBeInTheDocument();
      expect(screen.getByText("Devices")).toBeInTheDocument();
      expect(screen.getByText("Sessions")).toBeInTheDocument();
      expect(screen.getByText("Firewall Rules")).toBeInTheDocument();
      expect(screen.getByText("Namespaces")).toBeInTheDocument();
    });

    it("shows both Authentication and License in the Settings group", () => {
      renderSidebar();
      openSettingsGroup();
      expect(screen.getByText("Authentication")).toBeInTheDocument();
      expect(screen.getByText("License")).toBeInTheDocument();
    });
  });

  describe("enterprise admin with expired/no license (cloud=false, isExpired=true)", () => {
    beforeEach(() => {
      mockUseAdminLicense.mockReturnValue({
        isLoading: false,
        isExpired: true,
      } as never);
    });

    it("shows the restricted nav with only the License entry", () => {
      renderSidebar();
      openSettingsGroup();
      expect(screen.getByText("License")).toBeInTheDocument();
      expect(screen.queryByText("Authentication")).not.toBeInTheDocument();
      expect(screen.queryByText("Dashboard")).not.toBeInTheDocument();
    });
  });
});
