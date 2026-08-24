import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import React from "react";
import { defaultConfig } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import { mockLicense } from "@/tests/factories";
import { makeSdkError } from "@/tests/sdk";
import { getConfig } from "@/env";
import AdminSidebar from "../AdminSidebar";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getLicense: vi.fn(),
  }),
);

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

const mockGetConfig = vi.mocked(getConfig);

function renderSidebar() {
  return render(
    <MemoryRouter>
      <AdminSidebar expanded={true} pinned={true} onToggle={vi.fn()} />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

async function openSettingsGroup() {
  fireEvent.click(await screen.findByRole("button", { name: /settings/i }));
}

describe("AdminSidebar", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ isAdmin: true });
    sdk.getLicense.mockResolvedValue(
      mockSdkResponse(mockLicense({ expired: false })),
    );
    mockGetConfig.mockReturnValue({ ...defaultConfig });
  });

  describe("cloud admin (cloud=true, isAdmin=true)", () => {
    beforeEach(() => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
    });

    it("shows the core nav entries", async () => {
      renderSidebar();
      expect(await screen.findByText("Dashboard")).toBeInTheDocument();
      expect(screen.getByText("Users")).toBeInTheDocument();
      expect(screen.getByText("Devices")).toBeInTheDocument();
      expect(screen.getByText("Sessions")).toBeInTheDocument();
      expect(screen.getByText("Firewall Rules")).toBeInTheDocument();
      expect(screen.getByText("Namespaces")).toBeInTheDocument();
    });

    it("shows Authentication but NOT License in the Settings group", async () => {
      renderSidebar();
      await openSettingsGroup();
      expect(screen.getByText("Authentication")).toBeInTheDocument();
      expect(screen.queryByText("License")).not.toBeInTheDocument();
    });
  });

  describe("enterprise admin with valid license (cloud=false, isExpired=false)", () => {
    it("shows the core nav entries", async () => {
      renderSidebar();
      expect(await screen.findByText("Dashboard")).toBeInTheDocument();
      expect(screen.getByText("Users")).toBeInTheDocument();
      expect(screen.getByText("Devices")).toBeInTheDocument();
      expect(screen.getByText("Sessions")).toBeInTheDocument();
      expect(screen.getByText("Firewall Rules")).toBeInTheDocument();
      expect(screen.getByText("Namespaces")).toBeInTheDocument();
    });

    it("shows both Authentication and License in the Settings group", async () => {
      renderSidebar();
      await openSettingsGroup();
      expect(screen.getByText("Authentication")).toBeInTheDocument();
      expect(screen.getByText("License")).toBeInTheDocument();
    });
  });

  describe("enterprise admin with expired/no license (cloud=false, isExpired=true)", () => {
    beforeEach(() => {
      sdk.getLicense.mockRejectedValue(makeSdkError(400));
    });

    it("shows the restricted nav with only the License entry", async () => {
      renderSidebar();
      await openSettingsGroup();
      expect(screen.getByText("License")).toBeInTheDocument();
      expect(screen.queryByText("Authentication")).not.toBeInTheDocument();
      expect(screen.queryByText("Dashboard")).not.toBeInTheDocument();
    });
  });
});
