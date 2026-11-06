import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import type { Device } from "@/client";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { mockDevice, mockNamespace } from "@/tests/factories";
import { ClipboardProvider } from "@/components/common/ClipboardProvider";
import { useTerminalStore } from "@/stores/terminalStore";
import DeviceBriefDialog from "../DeviceBriefDialog";

function renderDialog(device: Device, onClose = vi.fn()) {
  server.use(
    http.get("*/api/devices/:uid", () => HttpResponse.json(device)),
  );
  render(
    <ClipboardProvider>
      <DeviceBriefDialog device={device} onClose={onClose} />
    </ClipboardProvider>,
    { wrapper: createTestWrapper({ initialEntries: ["/dashboard/map"] }) },
  );
  return onClose;
}

beforeEach(() => {
  seedAuthStore();
  useTerminalStore.setState({ sessions: [] });
  server.use(
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(mockNamespace({ name: "dev" })),
    ),
    http.get("*/api/ssh-identities", () => HttpResponse.json([])),
  );
});

describe("DeviceBriefDialog", () => {
  it("shows the device's SSHID", async () => {
    renderDialog(mockDevice({ name: "zurich-1" }));

    expect(
      await screen.findByText("dev.zurich-1@localhost"),
    ).toBeInTheDocument();
  });

  it("links to the device's details", () => {
    renderDialog(mockDevice({ uid: "abc" }));

    expect(
      screen.getByRole("link", { name: /open device details/i }),
    ).toHaveAttribute("href", "/devices/abc");
  });

  it("disables Connect while the device is offline", () => {
    renderDialog(mockDevice({ online: false }));

    expect(screen.getByRole("button", { name: "Connect" })).toBeDisabled();
  });

  it("swaps itself for the connect dialog when the device has no open terminal", async () => {
    const user = userEvent.setup();
    renderDialog(mockDevice({ name: "zurich-1" }));

    await user.click(screen.getByRole("button", { name: "Connect" }));

    expect(
      await screen.findByRole("dialog", { name: "Connect" }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("dialog", { name: "zurich-1" }),
    ).not.toBeInTheDocument();
  });

  it("closes once the connect dialog is cancelled", async () => {
    const user = userEvent.setup();
    const onClose = renderDialog(mockDevice());
    await user.click(screen.getByRole("button", { name: "Connect" }));
    await screen.findByRole("dialog", { name: "Connect" });

    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(onClose).toHaveBeenCalled();
  });

  it("brings back the device's open terminal instead of connecting again", async () => {
    const user = userEvent.setup();
    useTerminalStore.setState({
      sessions: [
        {
          id: "term-1",
          deviceUid: "abc",
          deviceName: "zurich-1",
          username: "root",
          password: "",
          state: "minimized",
          connectionStatus: "connected",
        },
      ],
    });
    const onClose = renderDialog(mockDevice({ uid: "abc", name: "zurich-1" }));

    await user.click(screen.getByRole("button", { name: "Connect" }));

    expect(useTerminalStore.getState().sessions[0].state).toBe("shown");
    expect(onClose).toHaveBeenCalled();
    expect(
      screen.queryByRole("dialog", { name: "Connect" }),
    ).not.toBeInTheDocument();
  });
});
