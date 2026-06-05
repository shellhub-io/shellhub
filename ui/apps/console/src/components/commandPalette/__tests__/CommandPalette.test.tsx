import { render, screen, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, useLocation } from "react-router-dom";
import { describe, it, expect, beforeEach, vi } from "vitest";
import type { NormalizedDevice } from "@/hooks/useDevices";

const { copyMock } = vi.hoisted(() => ({ copyMock: vi.fn() }));

vi.mock("@/hooks/useDevices", () => ({
  useDevices: vi.fn(),
}));

vi.mock("@/stores/authStore", () => ({
  useAuthStore: (sel: (s: { logout: () => void; tenant: string }) => unknown) =>
    sel({ logout: vi.fn(), tenant: "tenant-1" }),
}));

vi.mock("@/hooks/useHasPermission", () => ({
  useHasPermission: vi.fn(),
}));

vi.mock("@/hooks/useNamespaces", () => ({
  useNamespace: () => ({
    namespace: { name: "dev" },
    isLoading: false,
    error: null,
    refetch: vi.fn(),
  }),
}));

vi.mock("@/hooks/useCopy", () => ({
  useCopy: () => ({ copy: copyMock, copied: false }),
}));

import CommandPalette from "@/components/commandPalette/CommandPalette";
import { useDevices } from "@/hooks/useDevices";
import { useHasPermission } from "@/hooks/useHasPermission";
import { useCommandPaletteStore } from "@/stores/commandPaletteStore";
import { useTerminalStore } from "@/stores/terminalStore";

const device = {
  uid: "dev-1",
  name: "web-01",
  status: "accepted",
  online: true,
  identity: { mac: "00:11:22:33:44:55" },
  tags: [],
} as unknown as NormalizedDevice;

/** Reflects the router location so navigation can be asserted. */
function LocationProbe() {
  const location = useLocation();
  return <div data-testid="location">{location.pathname}</div>;
}

function renderPalette() {
  return render(
    <MemoryRouter>
      <CommandPalette />
      <LocationProbe />
    </MemoryRouter>,
  );
}

describe("CommandPalette", () => {
  beforeEach(() => {
    vi.mocked(useDevices).mockReturnValue({
      devices: [device],
      totalCount: 1,
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });
    vi.mocked(useHasPermission).mockReturnValue(true);
    copyMock.mockClear();
    useTerminalStore.setState({ sessions: [], reconnectTarget: null });
    useCommandPaletteStore.setState({ open: true });
  });

  it("shows devices and hides navigation by default", () => {
    renderPalette();

    expect(screen.getByText("web-01")).toBeInTheDocument();
    expect(screen.queryByText("Settings")).not.toBeInTheDocument();
  });

  it("reveals navigation and hides devices in command mode", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.type(screen.getByRole("combobox"), ">");

    expect(screen.getByText("Commands")).toBeInTheDocument();
    expect(screen.getByText("Settings")).toBeInTheDocument();
    expect(screen.getByText("Logout")).toBeInTheDocument();
    expect(screen.queryByText("web-01")).not.toBeInTheDocument();
  });

  it("filters commands by the term after >", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.type(screen.getByRole("combobox"), ">set");

    expect(screen.getByText("Settings")).toBeInTheDocument();
    expect(screen.queryByText("Dashboard")).not.toBeInTheDocument();
  });

  it("requests a connection when a device has no open session", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.click(screen.getByText("web-01"));

    expect(useTerminalStore.getState().reconnectTarget).toEqual({
      deviceUid: "dev-1",
      deviceName: "web-01",
    });
    expect(useCommandPaletteStore.getState().open).toBe(false);
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

    await user.click(screen.getByText("web-01"));

    expect(useTerminalStore.getState().reconnectTarget).toBeNull();
    expect(useTerminalStore.getState().sessions[0].state).toBe("docked");
    expect(useCommandPaletteStore.getState().open).toBe(false);
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

    await user.click(screen.getByText("web-01"));

    expect(useTerminalStore.getState().reconnectTarget).toEqual({
      deviceUid: "dev-1",
      deviceName: "web-01",
    });
    // the unrelated session is left untouched
    expect(useTerminalStore.getState().sessions[0].state).toBe("minimized");
  });

  it("closes the palette after running a command", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.type(screen.getByRole("combobox"), ">");
    await user.click(screen.getByText("Logout"));

    expect(useCommandPaletteStore.getState().open).toBe(false);
  });

  it("rejects connecting to an offline device and keeps the palette open", async () => {
    const user = userEvent.setup();
    vi.mocked(useDevices).mockReturnValue({
      devices: [{ ...device, online: false }],
      totalCount: 1,
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });
    renderPalette();

    await user.click(screen.getByText("web-01"));

    expect(useTerminalStore.getState().reconnectTarget).toBeNull();
    expect(useCommandPaletteStore.getState().open).toBe(true);
    expect(screen.getByRole("alert")).toHaveTextContent(/offline/i);
  });

  it("restores an existing session even when the device is offline", async () => {
    const user = userEvent.setup();
    vi.mocked(useDevices).mockReturnValue({
      devices: [{ ...device, online: false }],
      totalCount: 1,
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });
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

    await user.click(screen.getByText("web-01"));

    // the offline guard only blocks a fresh connect — an open session still restores
    expect(useTerminalStore.getState().sessions[0].state).toBe("docked");
    expect(useTerminalStore.getState().reconnectTarget).toBeNull();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(useCommandPaletteStore.getState().open).toBe(false);
  });

  it("rejects connecting without the device:connect permission", async () => {
    const user = userEvent.setup();
    vi.mocked(useHasPermission).mockReturnValue(false);
    renderPalette();

    await user.click(screen.getByText("web-01"));

    expect(useTerminalStore.getState().reconnectTarget).toBeNull();
    expect(useCommandPaletteStore.getState().open).toBe(true);
    expect(screen.getByRole("alert")).toHaveTextContent(/permission/i);
  });

  it("rejects restoring a session without the device:connect permission", async () => {
    const user = userEvent.setup();
    vi.mocked(useHasPermission).mockReturnValue(false);
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

    // the gate mirrors the Devices page: it covers restore, not just fresh connects
    await user.click(screen.getByText("web-01"));

    expect(useTerminalStore.getState().sessions[0].state).toBe("minimized");
    expect(useCommandPaletteStore.getState().open).toBe(true);
    expect(screen.getByRole("alert")).toHaveTextContent(/permission/i);
  });

  // ─── Phase 4: per-device drill-in action menu ──────────────────────────────

  it("opens the device action menu on ArrowRight without connecting", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.type(screen.getByRole("combobox"), "{ArrowRight}");

    expect(screen.getByText("Connect")).toBeInTheDocument();
    expect(screen.getByText("Copy SSHID")).toBeInTheDocument();
    expect(screen.getByText("Copy ssh command")).toBeInTheDocument();
    expect(screen.getByText("View details")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Back to devices" }),
    ).toBeInTheDocument();
    // → opens the menu; it must not connect
    expect(useTerminalStore.getState().reconnectTarget).toBeNull();
    // the device list is gone (its MAC sublabel no longer rendered)
    expect(screen.queryByText("00:11:22:33:44:55")).not.toBeInTheDocument();
  });

  it("filters the action menu by the typed term", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.type(screen.getByRole("combobox"), "{ArrowRight}");
    await user.type(screen.getByRole("combobox"), "copy");

    expect(screen.getByText("Copy SSHID")).toBeInTheDocument();
    expect(screen.getByText("Copy ssh command")).toBeInTheDocument();
    expect(screen.queryByText("Connect")).not.toBeInTheDocument();
    expect(screen.queryByText("View details")).not.toBeInTheDocument();
  });

  it("does not drill in when ArrowRight is pressed mid-text", async () => {
    const user = userEvent.setup();
    renderPalette();

    // with the caret away from the trailing edge, → must edit text (move the
    // caret), not open the action menu
    const input = screen.getByRole<HTMLInputElement>("combobox");
    await user.type(input, "web");
    input.setSelectionRange(1, 1);
    fireEvent.keyDown(input, { key: "ArrowRight" });

    expect(screen.queryByText("Copy SSHID")).not.toBeInTheDocument();
    expect(screen.getByText("web-01")).toBeInTheDocument();
  });

  it("returns to the device list on ArrowLeft", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.type(screen.getByRole("combobox"), "{ArrowRight}");
    expect(screen.getByText("Copy SSHID")).toBeInTheDocument();

    await user.type(screen.getByRole("combobox"), "{ArrowLeft}");
    expect(screen.queryByText("Copy SSHID")).not.toBeInTheDocument();
    expect(screen.getByText("00:11:22:33:44:55")).toBeInTheDocument();
  });

  it("opens the action menu when the row chevron is clicked", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.click(
      screen.getByRole("button", { name: "Show actions for web-01" }),
    );

    expect(screen.getByText("Copy SSHID")).toBeInTheDocument();
    // clicking the chevron drills in; it must not connect
    expect(useTerminalStore.getState().reconnectTarget).toBeNull();
  });

  it("copies the SSHID and confirms inline, staying open", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.type(screen.getByRole("combobox"), "{ArrowRight}");
    await user.click(screen.getByText("Copy SSHID"));

    expect(copyMock).toHaveBeenCalledWith("dev.web-01@localhost");
    expect(screen.getByText("Copied SSHID to clipboard")).toBeInTheDocument();
    expect(useCommandPaletteStore.getState().open).toBe(true);
  });

  it("copies the ssh command", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.type(screen.getByRole("combobox"), "{ArrowRight}");
    await user.click(screen.getByText("Copy ssh command"));

    expect(copyMock).toHaveBeenCalledWith(
      "ssh <username>@dev.web-01@localhost",
    );
  });

  it("navigates to device details and closes", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.type(screen.getByRole("combobox"), "{ArrowRight}");
    await user.click(screen.getByText("View details"));

    expect(screen.getByTestId("location")).toHaveTextContent("/devices/dev-1");
    expect(useCommandPaletteStore.getState().open).toBe(false);
  });

  it("connects from the action menu", async () => {
    const user = userEvent.setup();
    renderPalette();

    await user.type(screen.getByRole("combobox"), "{ArrowRight}");
    await user.click(screen.getByText("Connect"));

    expect(useTerminalStore.getState().reconnectTarget).toEqual({
      deviceUid: "dev-1",
      deviceName: "web-01",
    });
    expect(useCommandPaletteStore.getState().open).toBe(false);
  });

  it("disables the menu Connect action without permission", async () => {
    const user = userEvent.setup();
    vi.mocked(useHasPermission).mockReturnValue(false);
    renderPalette();

    await user.type(screen.getByRole("combobox"), "{ArrowRight}");

    const connect = screen.getByText("Connect").closest('[role="option"]');
    expect(connect).toHaveAttribute("aria-disabled", "true");
    expect(screen.getByText("Requires connect permission")).toBeInTheDocument();

    // inert: clicking it neither connects nor raises a rejection banner
    await user.click(screen.getByText("Connect"));

    expect(useTerminalStore.getState().reconnectTarget).toBeNull();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(useCommandPaletteStore.getState().open).toBe(true);
  });

  it("disables the menu Connect action for an offline device with no session", async () => {
    const user = userEvent.setup();
    vi.mocked(useDevices).mockReturnValue({
      devices: [{ ...device, online: false }],
      totalCount: 1,
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });
    renderPalette();

    await user.type(screen.getByRole("combobox"), "{ArrowRight}");

    const connect = screen.getByText("Connect").closest('[role="option"]');
    expect(connect).toHaveAttribute("aria-disabled", "true");

    await user.click(screen.getByText("Connect"));

    expect(useTerminalStore.getState().reconnectTarget).toBeNull();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(useCommandPaletteStore.getState().open).toBe(true);
  });

  it("keeps the menu Connect enabled for an offline device with an open session", async () => {
    const user = userEvent.setup();
    vi.mocked(useDevices).mockReturnValue({
      devices: [{ ...device, online: false }],
      totalCount: 1,
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });
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

    await user.type(screen.getByRole("combobox"), "{ArrowRight}");

    const connect = screen.getByText("Connect").closest('[role="option"]');
    expect(connect).not.toHaveAttribute("aria-disabled");

    // an offline device with a live session still restores
    await user.click(screen.getByText("Connect"));

    expect(useTerminalStore.getState().sessions[0].state).toBe("docked");
    expect(useCommandPaletteStore.getState().open).toBe(false);
  });
});
