import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, beforeEach, vi } from "vitest";
import type { Device } from "@/client";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestQueryClient, createTestWrapper } from "@/tests/wrapper";
import {
  mockDevice as mockDeviceFactory,
  mockNamespace,
} from "@/tests/factories";
import { setEdition } from "@/tests/edition";
import { seedAuthStore, VALID_JWT } from "@/tests/seedAuthStore";
import { useAuthStore } from "@/stores/authStore";
import type { TerminalSession } from "@/stores/terminalStore";

const { copyMock } = vi.hoisted(() => ({ copyMock: vi.fn() }));

vi.mock("@/hooks/useCopy", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/hooks/useCopy")>()),
  useCopy: () => copyMock,
}));

import CommandPalette from "@/components/commandPalette/CommandPalette";
import { useCommandPaletteStore } from "@/stores/commandPaletteStore";
import { useTerminalStore } from "@/stores/terminalStore";
import { useRecentDevicesStore } from "@/stores/recentDevicesStore";
import { useLocation } from "react-router-dom";

const device = mockDeviceFactory({
  uid: "dev-1",
  name: "web-01",
  status: "accepted",
  online: true,
  identity: { mac: "00:11:22:33:44:55" },
  tags: [],
});

const device2: Device = {
  ...device,
  uid: "dev-2",
  name: "db-01",
  identity: { mac: "aa:bb:cc:dd:ee:ff" },
};

const session = {
  id: "s1",
  deviceUid: "dev-1",
  deviceName: "web-01",
  username: "root",
  password: "",
  state: "minimized",
  connectionStatus: "connected",
} satisfies TerminalSession;

const HOUR_AGO = new Date(Date.now() - 60 * 60 * 1000).toISOString();

function LocationProbe() {
  const location = useLocation();
  return <div data-testid="location">{location.pathname}</div>;
}

const logoutSpy = vi.fn();

function setDevices(devices: Device[]) {
  server.use(http.get("*/api/devices", () => jsonWithTotal(devices)));
}

function renderPalette(queryClient = createTestQueryClient()) {
  return render(
    <>
      <CommandPalette />
      <LocationProbe />
    </>,
    { wrapper: createTestWrapper({ initialEntries: ["/"], queryClient }) },
  );
}

describe("CommandPalette", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setEdition("community");
    seedAuthStore();
    useAuthStore.setState({ logout: logoutSpy });
    server.use(
      http.get("*/api/devices", () => jsonWithTotal([device])),
      http.get("*/api/namespaces", () =>
        jsonWithTotal([mockNamespace({ name: "dev" })]),
      ),
      http.get("*/api/namespaces/:tenant", () =>
        HttpResponse.json(mockNamespace({ name: "dev" })),
      ),
      http.get("*/api/auth/token/:tenant", () =>
        HttpResponse.json({ token: VALID_JWT, role: "owner" }),
      ),
    );
    copyMock.mockClear();
    useTerminalStore.setState({ sessions: [], reconnectTarget: null });
    useRecentDevicesStore.setState({ byTenant: {} });
    useCommandPaletteStore.setState({ view: "palette" });
  });

  it("leaves the admin licence alone outside the admin console", async () => {
    const licenceRequested = vi.fn();
    useAuthStore.setState({ isAdmin: true });
    server.use(
      http.get("*/admin/api/license", () => {
        licenceRequested();
        return HttpResponse.json({});
      }),
    );
    renderPalette();

    expect(await screen.findByText("web-01")).toBeInTheDocument();
    expect(licenceRequested).not.toHaveBeenCalled();
  });

  it("shows devices and hides navigation by default", async () => {
    renderPalette();

    expect(await screen.findByText("web-01")).toBeInTheDocument();
    expect(screen.queryByText("Settings")).not.toBeInTheDocument();
  });

  it("reveals navigation and hides devices in command mode", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.type(await screen.findByRole("combobox"), ">");

    expect(screen.getByText("Commands")).toBeInTheDocument();
    expect(screen.getByText("Settings")).toBeInTheDocument();
    expect(screen.getByText("Logout")).toBeInTheDocument();
    expect(screen.queryByText("web-01")).not.toBeInTheDocument();
  });

  it("filters commands by the term after >", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.type(await screen.findByRole("combobox"), ">set");

    expect(screen.getByText("Settings")).toBeInTheDocument();
    expect(screen.queryByText("Dashboard")).not.toBeInTheDocument();
  });

  it("requests a connection when a device has no open session", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.click(await screen.findByText("web-01"));

    expect(useTerminalStore.getState().reconnectTarget).toEqual({
      deviceUid: "dev-1",
      deviceName: "web-01",
    });
    expect(useCommandPaletteStore.getState().view).toBeNull();
  });

  it("restores an existing session instead of connecting", async () => {
    const user = userEvent.setup();
    useTerminalStore.setState({
      sessions: [
        {
          id: "s1",
          deviceUid: "dev-1",
          deviceName: "web-01",
          username: "root",
          password: "",
          state: "minimized",
          connectionStatus: "connected",
        },
      ],
      reconnectTarget: null,
    });
    renderPalette();

    await user.click(await screen.findByText("web-01"));

    expect(useTerminalStore.getState().reconnectTarget).toBeNull();
    expect(useTerminalStore.getState().sessions[0].state).toBe("shown");
    expect(useCommandPaletteStore.getState().view).toBeNull();
  });

  it("connects rather than restoring a session for a different device", async () => {
    const user = userEvent.setup();
    useTerminalStore.setState({
      sessions: [
        {
          id: "s-other",
          deviceUid: "other-dev",
          deviceName: "db-01",
          username: "root",
          password: "",
          state: "minimized",
          connectionStatus: "connected",
        },
      ],
      reconnectTarget: null,
    });
    renderPalette();

    await user.click(await screen.findByText("web-01"));

    expect(useTerminalStore.getState().reconnectTarget).toEqual({
      deviceUid: "dev-1",
      deviceName: "web-01",
    });
    expect(useTerminalStore.getState().sessions[0].state).toBe("minimized");
  });

  it("closes the palette after running a command", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.type(await screen.findByRole("combobox"), ">");
    await user.click(screen.getByText("Logout"));

    expect(useCommandPaletteStore.getState().view).toBeNull();
  });

  it("rejects connecting to an offline device and keeps the palette open", async () => {
    const user = userEvent.setup();
    setDevices([{ ...device, online: false }]);
    renderPalette();

    await user.click(await screen.findByText("web-01"));

    expect(useTerminalStore.getState().reconnectTarget).toBeNull();
    expect(useCommandPaletteStore.getState().view).toBe("palette");
    expect(screen.getByRole("alert")).toHaveTextContent(/offline/i);
  });

  it("restores an existing session even when the device is offline", async () => {
    const user = userEvent.setup();
    setDevices([{ ...device, online: false }]);
    useTerminalStore.setState({
      sessions: [
        {
          id: "s1",
          deviceUid: "dev-1",
          deviceName: "web-01",
          username: "root",
          password: "",
          state: "minimized",
          connectionStatus: "connected",
        },
      ],
      reconnectTarget: null,
    });
    renderPalette();

    await user.click(await screen.findByText("web-01"));

    expect(useTerminalStore.getState().sessions[0].state).toBe("shown");
    expect(useTerminalStore.getState().reconnectTarget).toBeNull();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(useCommandPaletteStore.getState().view).toBeNull();
  });

  it("rejects connecting without the device:connect permission", async () => {
    const user = userEvent.setup();
    useAuthStore.setState({ role: null });
    renderPalette();

    await user.click(await screen.findByText("web-01"));

    expect(useTerminalStore.getState().reconnectTarget).toBeNull();
    expect(useCommandPaletteStore.getState().view).toBe("palette");
    expect(screen.getByRole("alert")).toHaveTextContent(/permission/i);
  });

  it("rejects restoring a session without the device:connect permission", async () => {
    const user = userEvent.setup();
    useAuthStore.setState({ role: null });
    useTerminalStore.setState({
      sessions: [
        {
          id: "s1",
          deviceUid: "dev-1",
          deviceName: "web-01",
          username: "root",
          password: "",
          state: "minimized",
          connectionStatus: "connected",
        },
      ],
      reconnectTarget: null,
    });
    renderPalette();

    await user.click(await screen.findByText("web-01"));

    expect(useTerminalStore.getState().sessions[0].state).toBe("minimized");
    expect(useCommandPaletteStore.getState().view).toBe("palette");
    expect(screen.getByRole("alert")).toHaveTextContent(/permission/i);
  });

  it("lists open terminal sessions above devices", async () => {
    useTerminalStore.setState({ sessions: [session], reconnectTarget: null });
    renderPalette();

    expect(await screen.findByText("Terminal Sessions")).toBeInTheDocument();
    await screen.findByText("00:11:22:33:44:55");
    const options = screen.getAllByRole("option");
    expect(options[0]).toHaveTextContent("root@web-01");
    expect(options[1]).toHaveTextContent("00:11:22:33:44:55");
  });

  it("restores a session from its top row", async () => {
    const user = userEvent.setup();
    useTerminalStore.setState({ sessions: [session], reconnectTarget: null });
    renderPalette();

    await user.click(await screen.findByText("root@web-01"));

    expect(useTerminalStore.getState().sessions[0].state).toBe("shown");
    expect(useCommandPaletteStore.getState().view).toBeNull();
  });

  it("restores the leading session on Enter by default", async () => {
    const user = userEvent.setup();
    useTerminalStore.setState({ sessions: [session], reconnectTarget: null });
    renderPalette();

    await user.type(await screen.findByRole("combobox"), "{Enter}");

    expect(useTerminalStore.getState().sessions[0].state).toBe("shown");
    expect(useCommandPaletteStore.getState().view).toBeNull();
  });

  it("lists recent devices between sessions and the full device list", async () => {
    setDevices([device, device2]);
    useTerminalStore.setState({ sessions: [session], reconnectTarget: null });
    useRecentDevicesStore.setState({
      byTenant: {
        "tenant-456": [{ uid: "dev-2", name: "db-01", connectedAt: HOUR_AGO }],
      },
    });
    renderPalette();

    expect(await screen.findByText("Recent")).toBeInTheDocument();
    const options = screen.getAllByRole("option");
    expect(options[0]).toHaveTextContent("root@web-01");
    expect(options[1]).toHaveTextContent("db-01");
    expect(screen.getByText(/1 hour ago/)).toBeInTheDocument();
  });

  it("hides a device with an open session from the Recent section", async () => {
    useTerminalStore.setState({ sessions: [session], reconnectTarget: null });
    useRecentDevicesStore.setState({
      byTenant: {
        "tenant-456": [{ uid: "dev-1", name: "web-01", connectedAt: HOUR_AGO }],
      },
    });
    renderPalette();

    expect(await screen.findByText("web-01")).toBeInTheDocument();
    expect(screen.queryByText("Recent")).not.toBeInTheDocument();
  });

  it("drops a recent device that is no longer in the device list", async () => {
    useRecentDevicesStore.setState({
      byTenant: {
        "tenant-456": [
          { uid: "ghost", name: "old-box", connectedAt: HOUR_AGO },
        ],
      },
    });
    renderPalette();

    expect(await screen.findByText("web-01")).toBeInTheDocument();
    expect(screen.queryByText("Recent")).not.toBeInTheDocument();
    expect(screen.queryByText("old-box")).not.toBeInTheDocument();
  });

  it("connects from a recent device row", async () => {
    const user = userEvent.setup();
    setDevices([device, device2]);
    useRecentDevicesStore.setState({
      byTenant: {
        "tenant-456": [{ uid: "dev-2", name: "db-01", connectedAt: HOUR_AGO }],
      },
    });
    renderPalette();

    const recentRow = (await screen.findByText(/1 hour ago/)).closest(
      '[role="option"]',
    );
    expect(recentRow).not.toBeNull();
    await user.click(recentRow as HTMLElement);

    expect(useTerminalStore.getState().reconnectTarget).toEqual({
      deviceUid: "dev-2",
      deviceName: "db-01",
    });
    expect(useCommandPaletteStore.getState().view).toBeNull();
  });

  it("lists a recent device once, not again under Devices", async () => {
    setDevices([device, device2]);
    useRecentDevicesStore.setState({
      byTenant: {
        "tenant-456": [{ uid: "dev-2", name: "db-01", connectedAt: HOUR_AGO }],
      },
    });
    renderPalette();

    expect(await screen.findByText(/1 hour ago/)).toBeInTheDocument();
    expect(screen.getAllByText("db-01")).toHaveLength(1);
    expect(screen.queryByText("aa:bb:cc:dd:ee:ff")).not.toBeInTheDocument();
  });

  it("shakes an offline recent row it can't connect to", async () => {
    const user = userEvent.setup();
    setDevices([device, { ...device2, online: false }]);
    useRecentDevicesStore.setState({
      byTenant: {
        "tenant-456": [{ uid: "dev-2", name: "db-01", connectedAt: HOUR_AGO }],
      },
    });
    renderPalette();

    const recentRow = await screen.findByRole("option", { name: /db-01/ });
    await user.click(recentRow);

    expect(recentRow.className).toContain("animate-shake");
    expect(screen.getByRole("alert")).toHaveTextContent(/offline/i);
  });

  it("leaves a device's actions, not the palette, on a click outside", async () => {
    const user = userEvent.setup();
    renderPalette();

    await screen.findByText("web-01");
    await user.type(screen.getByRole("combobox"), "{ArrowRight}");
    expect(screen.getByText("Copy SSHID")).toBeInTheDocument();
    await user.click(screen.getByRole("dialog", { name: "Command palette" }));

    expect(screen.queryByText("Copy SSHID")).not.toBeInTheDocument();
    expect(useCommandPaletteStore.getState().view).toBe("palette");
  });

  it("lists devices A to Z", async () => {
    setDevices([device, device2]);
    renderPalette();

    await screen.findByText("web-01");
    const options = screen.getAllByRole("option");
    expect(options[0]).toHaveTextContent("db-01");
    expect(options[1]).toHaveTextContent("web-01");
  });

  it("lists no namespaces", async () => {
    renderPalette();

    expect(await screen.findByText("web-01")).toBeInTheDocument();
    expect(screen.queryByText("dev")).not.toBeInTheDocument();
    expect(screen.queryByText("Namespaces")).not.toBeInTheDocument();
  });

  it("moves the highlight down across sections, tracking aria-activedescendant", async () => {
    const user = userEvent.setup();
    setDevices([device, device2]);
    useTerminalStore.setState({ sessions: [session], reconnectTarget: null });
    useRecentDevicesStore.setState({
      byTenant: {
        "tenant-456": [{ uid: "dev-2", name: "db-01", connectedAt: HOUR_AGO }],
      },
    });
    renderPalette();

    const input = await screen.findByRole("combobox");
    expect(input).toHaveAttribute("aria-activedescendant", "cmdk-opt-term-s1");
    await user.type(input, "{ArrowDown}");
    expect(input).toHaveAttribute(
      "aria-activedescendant",
      "cmdk-opt-recent-dev-2",
    );
    await user.type(input, "{ArrowDown}");
    expect(input).toHaveAttribute(
      "aria-activedescendant",
      "cmdk-opt-device-dev-1",
    );
  });

  it("wraps around at the list ends with ArrowUp/ArrowDown", async () => {
    const user = userEvent.setup();
    setDevices([device, device2]);
    renderPalette();

    const input = await screen.findByRole("combobox");
    await waitFor(() => {
      expect(input).toHaveAttribute(
        "aria-activedescendant",
        "cmdk-opt-device-dev-2",
      );
    });
    await user.type(input, "{ArrowUp}");
    expect(input).toHaveAttribute(
      "aria-activedescendant",
      "cmdk-opt-device-dev-1",
    );
    await user.type(input, "{ArrowDown}");
    expect(input).toHaveAttribute(
      "aria-activedescendant",
      "cmdk-opt-device-dev-2",
    );
  });

  it("jumps to the first and last option with Home and End", async () => {
    const user = userEvent.setup();
    setDevices([device, device2]);
    renderPalette();

    const input = await screen.findByRole("combobox");
    await waitFor(() => {
      expect(input).toHaveAttribute(
        "aria-activedescendant",
        "cmdk-opt-device-dev-2",
      );
    });
    await user.type(input, "{End}");
    expect(input).toHaveAttribute(
      "aria-activedescendant",
      "cmdk-opt-device-dev-1",
    );
    await user.type(input, "{Home}");
    expect(input).toHaveAttribute(
      "aria-activedescendant",
      "cmdk-opt-device-dev-2",
    );
  });

  it("selects the highlighted option after navigating", async () => {
    const user = userEvent.setup();
    setDevices([device, device2]);
    renderPalette();

    const input = await screen.findByRole("combobox");
    await waitFor(() => {
      expect(input).toHaveAttribute(
        "aria-activedescendant",
        "cmdk-opt-device-dev-2",
      );
    });
    await user.type(input, "{ArrowDown}");
    await user.type(input, "{Enter}");

    expect(useTerminalStore.getState().reconnectTarget).toEqual({
      deviceUid: "dev-1",
      deviceName: "web-01",
    });
    expect(useCommandPaletteStore.getState().view).toBeNull();
  });

  it("opens the device action menu on ArrowRight without connecting", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.type(await screen.findByRole("combobox"), "{ArrowRight}");

    expect(screen.getByText("Connect")).toBeInTheDocument();
    expect(screen.getByText("Copy SSHID")).toBeInTheDocument();
    expect(screen.getByText("Copy ssh command")).toBeInTheDocument();
    expect(screen.getByText("View details")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Back to devices" }),
    ).toBeInTheDocument();
    expect(useTerminalStore.getState().reconnectTarget).toBeNull();
    expect(screen.queryByText("00:11:22:33:44:55")).not.toBeInTheDocument();
  });

  it("filters the action menu by the typed term", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.type(await screen.findByRole("combobox"), "{ArrowRight}");
    await user.type(screen.getByRole("combobox"), "copy");

    expect(screen.getByText("Copy SSHID")).toBeInTheDocument();
    expect(screen.getByText("Copy ssh command")).toBeInTheDocument();
    expect(screen.queryByText("Connect")).not.toBeInTheDocument();
    expect(screen.queryByText("View details")).not.toBeInTheDocument();
  });

  it("does not drill in when ArrowRight is pressed mid-text", async () => {
    const user = userEvent.setup();
    renderPalette();

    const input = await screen.findByRole<HTMLInputElement>("combobox");
    await user.type(input, "web");
    input.setSelectionRange(1, 1);
    fireEvent.keyDown(input, { key: "ArrowRight" });

    expect(screen.queryByText("Copy SSHID")).not.toBeInTheDocument();
    expect(screen.getByText("web-01")).toBeInTheDocument();
  });

  it("returns to the device list on ArrowLeft", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.type(await screen.findByRole("combobox"), "{ArrowRight}");
    expect(screen.getByText("Copy SSHID")).toBeInTheDocument();

    await user.type(screen.getByRole("combobox"), "{ArrowLeft}");
    expect(screen.queryByText("Copy SSHID")).not.toBeInTheDocument();
    expect(screen.getByText("00:11:22:33:44:55")).toBeInTheDocument();
  });

  it("opens the action menu when the row chevron is clicked", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.click(
      await screen.findByRole("button", { name: "Show actions for web-01" }),
    );

    expect(screen.getByText("Copy SSHID")).toBeInTheDocument();
    expect(useTerminalStore.getState().reconnectTarget).toBeNull();
  });

  it("copies the SSHID and confirms inline, staying open", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.type(await screen.findByRole("combobox"), "{ArrowRight}");
    await user.click(screen.getByText("Copy SSHID"));

    expect(copyMock).toHaveBeenCalledWith("dev.web-01@localhost");
    expect(screen.getByText("Copied SSHID to clipboard")).toBeInTheDocument();
    expect(useCommandPaletteStore.getState().view).toBe("palette");
  });

  it("copies the ssh command", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.type(await screen.findByRole("combobox"), "{ArrowRight}");
    await user.click(screen.getByText("Copy ssh command"));

    expect(copyMock).toHaveBeenCalledWith("ssh <login>@dev.web-01@localhost");
  });

  it("navigates to device details and closes", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.type(await screen.findByRole("combobox"), "{ArrowRight}");
    await user.click(screen.getByText("View details"));

    expect(screen.getByTestId("location")).toHaveTextContent("/devices/dev-1");
    expect(useCommandPaletteStore.getState().view).toBeNull();
  });

  it("connects from the action menu", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.type(await screen.findByRole("combobox"), "{ArrowRight}");
    await user.click(screen.getByText("Connect"));

    expect(useTerminalStore.getState().reconnectTarget).toEqual({
      deviceUid: "dev-1",
      deviceName: "web-01",
    });
    expect(useCommandPaletteStore.getState().view).toBeNull();
  });

  it("disables the menu Connect action without permission", async () => {
    const user = userEvent.setup();
    useAuthStore.setState({ role: null });
    renderPalette();

    await user.type(await screen.findByRole("combobox"), "{ArrowRight}");

    const connect = screen.getByText("Connect").closest('[role="option"]');
    expect(connect).toHaveAttribute("aria-disabled", "true");
    expect(screen.getByText("Requires connect permission")).toBeInTheDocument();

    await user.click(screen.getByText("Connect"));

    expect(useTerminalStore.getState().reconnectTarget).toBeNull();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(useCommandPaletteStore.getState().view).toBe("palette");
  });

  it("disables the menu Connect action for an offline device with no session", async () => {
    const user = userEvent.setup();
    setDevices([{ ...device, online: false }]);
    renderPalette();

    await user.type(await screen.findByRole("combobox"), "{ArrowRight}");

    const connect = screen.getByText("Connect").closest('[role="option"]');
    expect(connect).toHaveAttribute("aria-disabled", "true");

    await user.click(screen.getByText("Connect"));

    expect(useTerminalStore.getState().reconnectTarget).toBeNull();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(useCommandPaletteStore.getState().view).toBe("palette");
  });

  it("keeps the menu Connect enabled for an offline device with an open session", async () => {
    const user = userEvent.setup();
    setDevices([{ ...device, online: false }]);
    useTerminalStore.setState({
      sessions: [
        {
          id: "s1",
          deviceUid: "dev-1",
          deviceName: "web-01",
          username: "root",
          password: "",
          state: "minimized",
          connectionStatus: "connected",
        },
      ],
      reconnectTarget: null,
    });
    renderPalette();

    await user.click(
      await screen.findByRole("button", { name: "Show actions for web-01" }),
    );

    const connect = screen.getByText("Connect").closest('[role="option"]');
    expect(connect).not.toHaveAttribute("aria-disabled");

    await user.click(screen.getByText("Connect"));

    expect(useTerminalStore.getState().sessions[0].state).toBe("shown");
    expect(useCommandPaletteStore.getState().view).toBeNull();
  });

  describe("opened from the new-tab picker", () => {
    beforeEach(() => {
      useCommandPaletteStore.setState({ view: "connect" });
    });

    it("searches the current namespace's devices, without terminals", async () => {
      useTerminalStore.setState({ sessions: [session], reconnectTarget: null });
      renderPalette();

      expect(
        await screen.findByRole("dialog", { name: "Connect to a device" }),
      ).toBeInTheDocument();
      expect(await screen.findByText("in dev")).toBeInTheDocument();
      expect(screen.getByText("web-01")).toBeInTheDocument();
      expect(screen.queryByText("root@web-01")).not.toBeInTheDocument();
    });

    it("treats > as a search term, not command mode", async () => {
      const user = userEvent.setup();
      renderPalette();

      await user.type(await screen.findByRole("combobox"), ">");

      expect(screen.queryByText("Dashboard")).not.toBeInTheDocument();
      expect(screen.getByText("No device in dev matches")).toBeInTheDocument();
    });

    it("returns to the picker on Escape", async () => {
      renderPalette();

      await screen.findByText("web-01");
      fireEvent(
        screen.getByRole("dialog", { name: "Connect to a device" }),
        new Event("cancel", { cancelable: true }),
      );

      expect(useCommandPaletteStore.getState().view).toBe("picker");
    });

    it("closes outright on a click outside", async () => {
      const user = userEvent.setup();
      renderPalette();

      await user.click(
        await screen.findByRole("dialog", { name: "Connect to a device" }),
      );

      expect(useCommandPaletteStore.getState().view).toBeNull();
    });

    it("closes outright on a click outside a device's actions", async () => {
      const user = userEvent.setup();
      renderPalette();

      await screen.findByText("web-01");
      await user.type(screen.getByRole("combobox"), "{ArrowRight}");
      expect(screen.getByText("Copy SSHID")).toBeInTheDocument();
      await user.click(
        screen.getByRole("dialog", { name: "Connect to a device" }),
      );

      expect(useCommandPaletteStore.getState().view).toBeNull();
    });

    it("returns to the picker from the back button", async () => {
      const user = userEvent.setup();
      renderPalette();

      await user.click(
        await screen.findByRole("button", { name: "Back to new tab" }),
      );

      expect(useCommandPaletteStore.getState().view).toBe("picker");
    });

    it("returns to the picker on ArrowLeft with an empty query", async () => {
      const user = userEvent.setup();
      renderPalette();

      await user.type(await screen.findByRole("combobox"), "{ArrowLeft}");

      expect(useCommandPaletteStore.getState().view).toBe("picker");
    });

    it("leaves a device's actions before returning to the picker", async () => {
      const user = userEvent.setup();
      renderPalette();

      await screen.findByText("web-01");
      await user.type(screen.getByRole("combobox"), "{ArrowRight}");
      expect(screen.getByText("Copy SSHID")).toBeInTheDocument();
      fireEvent(
        screen.getByRole("dialog", { name: "Connect to a device" }),
        new Event("cancel", { cancelable: true }),
      );

      expect(screen.queryByText("Copy SSHID")).not.toBeInTheDocument();
      expect(useCommandPaletteStore.getState().view).toBe("connect");
    });
  });
});
