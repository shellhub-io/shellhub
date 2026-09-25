import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import React from "react";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { defaultConfig } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import { createTestWrapper } from "@/tests/wrapper";
import { mockLicense } from "@/tests/factories";
import { getConfig } from "@/env";
import AdminSidebar from "../AdminSidebar";

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
      <AdminSidebar expanded />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

describe("AdminSidebar", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ isAdmin: true });
    server.use(
      http.get("*/admin/api/license", () =>
        HttpResponse.json(mockLicense({ expired: false })),
      ),
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

    it("shows Authentication but NOT License under Instance", async () => {
      renderSidebar();
      const instance = await screen.findByRole("group", { name: "Instance" });
      expect(within(instance).getByText("Authentication")).toBeInTheDocument();
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

    it.each([
      ["Accounts", ["Users", "Namespaces"]],
      ["Resources", ["Devices", "Sessions", "Firewall Rules"]],
      ["Instance", ["Authentication", "Instance API Keys", "License"]],
    ])("lists %s pages under their heading", async (title, labels) => {
      renderSidebar();
      const group = await screen.findByRole("group", { name: title });
      for (const label of labels) {
        expect(within(group).getByText(label)).toBeInTheDocument();
      }
    });
  });

  describe("enterprise admin with expired/no license (cloud=false, isExpired=true)", () => {
    beforeEach(() => {
      server.use(
        http.get("*/admin/api/license", () =>
          HttpResponse.json({}, { status: 400 }),
        ),
      );
    });

    it("shows the restricted nav with only the License entry", async () => {
      renderSidebar();
      expect(await screen.findByText("License")).toBeInTheDocument();
      expect(screen.queryByText("Instance")).not.toBeInTheDocument();
      expect(screen.queryByText("Authentication")).not.toBeInTheDocument();
      expect(screen.queryByText("Dashboard")).not.toBeInTheDocument();
    });
  });

  describe("non-admin", () => {
    beforeEach(() => {
      useAuthStore.setState({ isAdmin: false });
    });

    it("shows the License entry disabled", async () => {
      renderSidebar();
      expect(await screen.findByText("License")).toHaveAttribute(
        "aria-disabled",
        "true",
      );
    });
  });
});
