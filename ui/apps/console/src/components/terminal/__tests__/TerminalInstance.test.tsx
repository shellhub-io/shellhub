import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace, mockRecordingMeta } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { server } from "@/tests/msw";
import { defaultHandlers } from "@/tests/handlers";
import { useTerminalStore, type TerminalSession } from "@/stores/terminalStore";
import { useRecordingsStore } from "@/stores/recordingsStore";
import { WS_KIND } from "../terminalErrors";

vi.mock("@xterm/xterm", () => ({ Terminal: buildXtermFake() }));
vi.mock("@xterm/addon-fit", () => ({ FitAddon: buildXtermFake() }));
vi.mock("@xterm/addon-web-links", () => ({ WebLinksAddon: buildXtermFake() }));
vi.mock("@xterm/addon-webgl", () => ({ WebglAddon: buildXtermFake() }));
const recorderCreation = vi.hoisted(() => ({
  created: Promise.resolve<unknown>(undefined),
  finish: (_: unknown) => {},
}));

vi.mock("@/utils/recordings", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/utils/recordings")>()),
  OpfsCastRecorder: { create: () => recorderCreation.created },
  listRecordings: vi.fn(async () => []),
}));

function buildXtermFake() {
  return class {
    cols = 80;
    rows = 24;
    options = {};
    constructor() {
      return new Proxy(this, {
        get: (target, key) =>
          key in target
            ? target[key as keyof typeof target]
            : () => ({ dispose: () => {} }),
      });
    }
  };
}

class FakeWebSocket {
  static readonly OPEN = 1;
  static readonly opened: FakeWebSocket[] = [];
  readyState = FakeWebSocket.OPEN;
  binaryType = "";
  onopen: (() => Promise<void>) | null = null;
  onmessage: ((event: { data: string }) => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: (() => void) | null = null;
  constructor() {
    FakeWebSocket.opened.push(this);
  }

  send() {}
  close() {}
}

class FakeRecorder {
  sessionUid?: string;
  setSessionUid(uid: string) {
    this.sessionUid = uid;
  }

  start() {}
  recordOutput() {}
  recordResize() {}
  async discard() {}
  async finish() {
    return mockRecordingMeta({ sessionUid: this.sessionUid });
  }
}

const { default: TerminalInstance } = await import("../TerminalInstance");

function makeSession(
  overrides: Partial<TerminalSession> = {},
): TerminalSession {
  return {
    id: "session-1",
    deviceUid: "device-uid",
    deviceName: "my-device",
    username: "root",
    password: "secret",
    state: "shown",
    connectionStatus: "connecting",
    ...overrides,
  };
}

function renderTerminal(overrides: Partial<TerminalSession> = {}) {
  return render(<TerminalInstance session={makeSession(overrides)} visible />, {
    wrapper: createTestWrapper({ initialEntries: ["/devices"] }),
  });
}

beforeEach(() => {
  server.use(
    ...defaultHandlers,
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(mockNamespace()),
    ),
  );
  seedAuthStore();
  useTerminalStore.setState({ sessions: [makeSession()] });
  useRecordingsStore.setState({ notice: null });
  recorderCreation.created = new Promise((resolve) => {
    recorderCreation.finish = resolve;
  });
});

afterEach(() => {
  vi.unstubAllGlobals();
  FakeWebSocket.opened.length = 0;
});

describe("TerminalInstance", () => {
  it("tells an observer their role is the reason, and offers no retry", async () => {
    seedAuthStore({ role: "observer" });
    server.use(
      http.post("*/ws/ssh/session", () =>
        HttpResponse.json(null, { status: 403 }),
      ),
    );
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
    server.use(
      http.post("*/ws/ssh/session", () =>
        HttpResponse.json(null, { status: 500 }),
      ),
    );
    renderTerminal();

    const banner = await screen.findByRole("alert");

    expect(banner).toHaveTextContent("Connection failed");
    expect(screen.getByRole("button", { name: "Retry" })).toBeInTheDocument();
  });

  it("keeps the session id on a recording that starts after the session does", async () => {
    vi.stubGlobal("WebSocket", FakeWebSocket);
    server.use(
      http.post("*/ws/ssh/session", () => HttpResponse.json({ token: "t" })),
    );
    renderTerminal({ record: true });
    const ws = await vi.waitUntil(() => FakeWebSocket.opened[0]);

    const opening = ws.onopen?.();
    ws.onmessage?.({
      data: JSON.stringify({ kind: WS_KIND.SESSION, data: "session-uid" }),
    });
    recorderCreation.finish(new FakeRecorder());
    await opening;
    ws.onclose?.();

    await vi.waitFor(() =>
      expect(useRecordingsStore.getState().notice?.sessionUid).toBe(
        "session-uid",
      ),
    );
  });
});
