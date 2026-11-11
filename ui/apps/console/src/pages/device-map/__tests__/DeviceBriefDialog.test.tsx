import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import type { Device } from "@/client";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { mockDevice, mockNamespace } from "@/tests/factories";
import { connectDialogHandlers } from "@/tests/handlers";
import { useTerminalStore } from "@/stores/terminalStore";
import TerminalManager from "@/components/terminal/TerminalManager";
import DeviceBriefDialog from "../DeviceBriefDialog";

function renderDialog(device: Device, onClose = vi.fn()) {
  render(
    <>
      <DeviceBriefDialog device={device} onClose={onClose} />
      <TerminalManager />
    </>,
    { wrapper: createTestWrapper({ initialEntries: ["/dashboard/map"] }) },
  );
  return onClose;
}

beforeEach(() => {
  seedAuthStore();
  useTerminalStore.setState({ sessions: [], reconnectTarget: null });
  server.use(
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(mockNamespace({ name: "dev" })),
    ),
    http.get("*/api/namespaces", () => jsonWithTotal([mockNamespace()])),
    ...connectDialogHandlers,
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

  it("closes and opens the connect dialog", async () => {
    const user = userEvent.setup();
    const onClose = renderDialog(mockDevice({ uid: "abc", name: "zurich-1" }));

    await user.click(screen.getByRole("button", { name: "Connect" }));

    expect(onClose).toHaveBeenCalled();
    expect(
      await screen.findByRole("dialog", { name: "Connect" }),
    ).toBeInTheDocument();
  });
});
