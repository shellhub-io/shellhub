import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse, makeSdkError } from "@/tests/sdk";
import { mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { useTerminalStore, type TerminalSession } from "@/stores/terminalStore";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    createWebSshSession: vi.fn(),
    getNamespace: vi.fn(),
  }),
);

const { default: TerminalInstance } = await import("../TerminalInstance");

function makeSession(overrides: Partial<TerminalSession> = {}): TerminalSession {
  return {
    id: "session-1",
    deviceUid: "device-uid",
    deviceName: "my-device",
    username: "root",
    password: "secret",
    state: "docked",
    connectionStatus: "connecting",
    ...overrides,
  };
}

function renderTerminal() {
  return render(<TerminalInstance session={makeSession()} visible />, {
    wrapper: createTestWrapper({ initialEntries: ["/devices"] }),
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
  useTerminalStore.setState({ sessions: [makeSession()] });
  sdk.getNamespace.mockResolvedValue(mockSdkResponse(mockNamespace()));
});

describe("TerminalInstance", () => {
  it("tells an observer their role is the reason, and offers no retry", async () => {
    seedAuthStore({ role: "observer" });
    sdk.createWebSshSession.mockRejectedValue(makeSdkError(403));
    renderTerminal();

    const banner = await screen.findByRole("alert");

    expect(banner).toHaveTextContent(
      "Your role does not allow connecting to devices.",
    );
    expect(
      screen.queryByRole("button", { name: "Retry" }),
    ).not.toBeInTheDocument();
  });

  it("offers a retry when the session fails for any other reason", async () => {
    sdk.createWebSshSession.mockRejectedValue(makeSdkError(500));
    renderTerminal();

    const banner = await screen.findByRole("alert");

    expect(banner).toHaveTextContent("Connection failed");
    expect(screen.getByRole("button", { name: "Retry" })).toBeInTheDocument();
  });
});
