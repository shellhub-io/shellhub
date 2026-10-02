import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { ClipboardProvider } from "@/components/common/ClipboardProvider";
import { getConfig, defaultConfig } from "@/env";
import { useTerminalStore } from "@/stores/terminalStore";
import AppLayout from "../AppLayout";

const viewport = vi.hoisted(() => ({ desktop: true, drawerOpen: false }));

vi.mock("@/hooks/useSidebarLayout", () => ({
  useSidebarLayout: () => ({
    expanded: false,
    pinned: false,
    isOpen: false,
    isDesktop: viewport.desktop,
    isWide: viewport.desktop,
    drawerOpen: viewport.drawerOpen,
    handlers: {
      onMouseEnter: vi.fn(),
      onMouseLeave: vi.fn(),
      onFocus: vi.fn(),
      onBlur: vi.fn(),
      openDrawer: vi.fn(),
      closeDrawer: vi.fn(),
      onDrawerKeyDown: vi.fn(),
    },
  }),
}));

vi.mock("@/components/terminal/TerminalManager", () => ({
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
  viewport.desktop = true;
  viewport.drawerOpen = false;
  useTerminalStore.setState({ sessions: [], recordings: [] });
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
      expect(
        await screen.findByRole("navigation", { name: "Main navigation" }),
      ).toBeInTheDocument();
    });

    it("is hidden when there are no namespaces", async () => {
      renderLayout();
      await waitFor(() => {
        expect(
          screen.queryByRole("navigation", { name: "Main navigation" }),
        ).not.toBeInTheDocument();
      });
    });

    it.each([
      ["a wide", true],
      ["a narrow", false],
    ])(
      "keeps the account menu reachable in the admin console on %s window",
      async (_, desktop) => {
        viewport.desktop = desktop;
        seedAuthStore({ isAdmin: true });
        mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
        renderLayout("/admin/instance");

        const menus = await screen.findAllByRole("button", {
          name: /account menu for/i,
        });
        const reachable = menus.filter((menu) => !menu.closest("[inert]"));
        expect(reachable).toHaveLength(1);
        if (desktop) expect(reachable[0]).toHaveTextContent("admin");
        else expect(reachable[0]).not.toHaveTextContent("admin");
      },
    );

    it("gives way to the admin bar across the top on admin routes", async () => {
      seedAuthStore({ isAdmin: true });
      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
      renderLayout("/admin/instance");
      expect(
        await screen.findByRole("link", { name: "Instance" }),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("navigation", { name: "Main navigation" }),
      ).not.toBeInTheDocument();
    });
  });

  it.each(["/account/profile", "/preferences/appearance"])(
    "keeps the account menu reachable on %s without a namespace",
    async (path) => {
      renderLayout(path);

      expect(
        await screen.findByRole("button", { name: /account menu for/i }),
      ).toBeInTheDocument();
    },
  );

  describe("on the account pages", () => {
    beforeEach(() => {
      server.use(
        http.get("*/api/namespaces", () => jsonWithTotal([mockNamespace()])),
      );
    });

    it("covers the namespace navigation with the page frame", async () => {
      renderLayout("/account/profile");
      expect(
        await screen.findByRole("navigation", { name: "Main navigation" }),
      ).toHaveAttribute("inert");
    });

    it("moves the account menu out from under the frame", async () => {
      renderLayout("/account/profile");
      await screen.findByRole("navigation", { name: "Main navigation" });

      const menus = screen.getAllByRole("button", {
        name: /account menu for/i,
      });
      const reachable = menus.filter((menu) => !menu.closest("[inert]"));
      expect(menus).toHaveLength(2);
      expect(reachable).toHaveLength(1);
    });

    it("leaves the navigation uncovered elsewhere", async () => {
      renderLayout("/dashboard");
      expect(
        await screen.findByRole("navigation", { name: "Main navigation" }),
      ).not.toHaveAttribute("inert");
      expect(
        screen.getAllByRole("button", { name: /account menu for/i }),
      ).toHaveLength(1);
    });
  });

  describe.each([
    [
      "a terminal",
      {
        sessions: [
          {
            id: "session-1",
            deviceUid: "device-uid",
            deviceName: "my-device",
            username: "root",
            password: "",
            state: "shown" as const,
            connectionStatus: "connected" as const,
          },
        ],
      },
    ],
    [
      "a recording",
      {
        recordings: [
          {
            id: "recording-1",
            title: "my-device",
            logs: "",
            filename: "my-device.cast",
            recorded: true,
            shown: true,
          },
        ],
      },
    ],
  ])("with %s shown", (_, windows) => {
    beforeEach(() => {
      server.use(
        http.get("*/api/namespaces", () => jsonWithTotal([mockNamespace()])),
      );
      useTerminalStore.setState(windows);
    });

    it("covers the desktop navigation with the frame", async () => {
      renderLayout("/devices");

      const nav = await screen.findByRole("navigation", {
        name: "Main navigation",
        hidden: true,
      });
      expect(nav.closest("[inert]")).not.toBeNull();
    });

    it("keeps the account menu reachable beside the tabs", async () => {
      renderLayout("/devices");

      const menus = await screen.findAllByRole("button", {
        name: /account menu for/i,
        hidden: true,
      });
      expect(menus.filter((menu) => !menu.closest("[inert]"))).toHaveLength(1);
    });

    it("puts the window away when the logo is clicked on its own page", async () => {
      const user = userEvent.setup();
      renderLayout("/dashboard");

      await user.click(
        await screen.findByRole("link", { name: "ShellHub", hidden: true }),
      );

      await waitFor(() =>
        expect(
          screen.getByRole("navigation", { name: "Main navigation" }),
        ).not.toHaveAttribute("inert"),
      );
    });

    it("puts the window away when the admin logo is clicked on its own page", async () => {
      const user = userEvent.setup();
      seedAuthStore({ isAdmin: true });
      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
      renderLayout("/admin/instance");

      const adminTab = await screen.findByRole("tab", {
        name: /admin console/i,
      });
      expect(adminTab).toHaveAttribute("aria-selected", "false");

      await user.click(
        screen.getByRole("link", { name: "ShellHub", hidden: true }),
      );

      await waitFor(() =>
        expect(adminTab).toHaveAttribute("aria-selected", "true"),
      );
    });

    it("puts the window away when Account is picked on the account's own page", async () => {
      const user = userEvent.setup();
      renderLayout("/account/profile");

      const accountTab = await screen.findByRole("tab", { name: /account/i });
      expect(accountTab).toHaveAttribute("aria-selected", "false");

      const menu = (
        await screen.findAllByRole("button", {
          name: /account menu for/i,
          hidden: true,
        })
      ).find((button) => !button.closest("[inert]"))!;
      await user.click(menu);
      await user.click(await screen.findByRole("button", { name: "Account" }));

      await waitFor(() =>
        expect(accountTab).toHaveAttribute("aria-selected", "true"),
      );
    });

    it("keeps the navigation drawer usable on a narrow window", async () => {
      viewport.desktop = false;
      viewport.drawerOpen = true;
      renderLayout("/devices");

      const nav = await screen.findByRole("navigation", {
        name: "Main navigation",
      });
      expect(nav.closest("[inert]")).toBeNull();
      expect(
        screen.getAllByRole("button", { name: /account menu for/i }),
      ).toHaveLength(1);
    });
  });

  describe("tab strip", () => {
    it("renders regardless of namespaces", async () => {
      renderLayout();
      expect(
        await screen.findByRole("tablist", { name: "Open views" }),
      ).toBeInTheDocument();
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
        await screen.findByRole("tablist", { name: "Open views" });
        expect(
          screen.queryByTestId("device-limit-banner"),
        ).not.toBeInTheDocument();
        expect(screen.queryByTestId("license-banner")).not.toBeInTheDocument();
      },
    );
  });
});
