import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { describe, it, expect, beforeEach, vi } from "vitest";
import type { NormalizedDevice } from "@/hooks/useDevices";

vi.mock("@/hooks/useDevices", () => ({
  useDevices: vi.fn(),
}));

vi.mock("@/stores/authStore", () => ({
  useAuthStore: (sel: (s: { logout: () => void }) => unknown) =>
    sel({ logout: vi.fn() }),
}));

vi.mock("@/hooks/useHasPermission", () => ({
  useHasPermission: vi.fn(),
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

function renderPalette() {
  return render(
    <MemoryRouter>
      <CommandPalette />
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
});
