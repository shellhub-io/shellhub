import { StrictMode } from "react";
import { render, screen, fireEvent, act } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import AppBar from "../AppBar";
import { useTerminalStore, type TerminalSession } from "@/stores/terminalStore";

vi.mock("../NamespaceSelector", () => ({
  default: () => <div data-testid="namespace-selector" />,
}));
vi.mock("../UserMenu", () => ({
  default: () => <div data-testid="user-menu" />,
}));
vi.mock("../SupportButton", () => ({
  default: () => <div data-testid="support-button" />,
}));
vi.mock("../../terminal/TerminalControls", () => ({
  TerminalInfo: ({ session }: { session: TerminalSession }) => (
    <div data-testid="terminal-info">{session.deviceName}</div>
  ),
  TerminalActions: ({ session }: { session: TerminalSession }) => (
    <div data-testid="terminal-actions">{session.deviceName}</div>
  ),
}));
vi.mock("@/stores/terminalThemeStore", () => ({
  useTerminalThemeStore: (selector: (s: unknown) => unknown) =>
    selector({ theme: { colors: { background: "#000000" } } }),
}));

function makeSession(
  overrides: Partial<TerminalSession> = {},
): TerminalSession {
  return {
    id: "t1",
    deviceUid: "uid-1",
    deviceName: "device-1",
    username: "root",
    password: "",
    state: "docked",
    connectionStatus: "connected",
    ...overrides,
  };
}

function renderAppBar() {
  return render(
    <MemoryRouter>
      <AppBar />
    </MemoryRouter>,
  );
}

function contentWrapper(): HTMLElement {
  const child =
    screen.queryByTestId("terminal-info") ??
    screen.getByTestId("namespace-selector");
  return child.parentElement as HTMLElement;
}

beforeEach(() => {
  useTerminalStore.setState({ sessions: [] });
  vi.spyOn(global, "requestAnimationFrame").mockImplementation(
    (cb: FrameRequestCallback) => {
      cb(0);
      return 0;
    },
  );
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe("AppBar", () => {
  it("shows the namespace selector when there is no active session", () => {
    renderAppBar();

    expect(screen.getByTestId("namespace-selector")).toBeInTheDocument();
    expect(screen.queryByTestId("terminal-info")).not.toBeInTheDocument();
    expect(contentWrapper().className).toContain("opacity-100");
  });

  it("crossfades into terminal mode when a session becomes active", () => {
    renderAppBar();

    act(() => {
      useTerminalStore.setState({
        sessions: [makeSession({ id: "t1", deviceName: "device-1" })],
      });
    });

    const wrapper = contentWrapper();
    expect(screen.getByTestId("namespace-selector")).toBeInTheDocument();
    expect(wrapper.className).toContain("opacity-0");

    act(() => {
      fireEvent.transitionEnd(wrapper);
    });

    expect(screen.getByTestId("terminal-info")).toHaveTextContent("device-1");
    expect(contentWrapper().className).toContain("opacity-100");
  });

  it("swaps same-mode sessions instantly without a fade (no flash)", () => {
    useTerminalStore.setState({
      sessions: [makeSession({ id: "t1", deviceName: "device-1" })],
    });
    renderAppBar();
    expect(screen.getByTestId("terminal-info")).toHaveTextContent("device-1");

    act(() => {
      useTerminalStore.setState({
        sessions: [makeSession({ id: "t2", deviceName: "device-2" })],
      });
    });

    expect(screen.getByTestId("terminal-info")).toHaveTextContent("device-2");
    const wrapper = contentWrapper();
    expect(wrapper.className).toContain("opacity-100");
    expect(wrapper.className).not.toContain("opacity-0");
  });

  it("recovers to namespace mode if the session vanishes mid fade-out (safety net)", () => {
    renderAppBar();

    act(() => {
      useTerminalStore.setState({
        sessions: [makeSession({ id: "t1", deviceName: "device-1" })],
      });
    });
    expect(contentWrapper().className).toContain("opacity-0");

    act(() => {
      useTerminalStore.setState({ sessions: [] });
    });

    expect(screen.getByTestId("namespace-selector")).toBeInTheDocument();
    expect(screen.queryByTestId("terminal-info")).not.toBeInTheDocument();
    expect(contentWrapper().className).toContain("opacity-100");
  });

  it("renders correctly under StrictMode (idempotent render-phase swap)", () => {
    useTerminalStore.setState({
      sessions: [makeSession({ id: "t1", deviceName: "device-1" })],
    });
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

    render(
      <StrictMode>
        <MemoryRouter>
          <AppBar />
        </MemoryRouter>
      </StrictMode>,
    );

    expect(screen.getByTestId("terminal-info")).toHaveTextContent("device-1");
    expect(errorSpy).not.toHaveBeenCalled();
  });
});
