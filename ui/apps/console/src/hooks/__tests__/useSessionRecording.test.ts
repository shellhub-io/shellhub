import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { mockDevice, mockRecordingMeta, mockSession } from "@/tests/factories";

const readRecording = vi.hoisted(() => vi.fn<() => Promise<string>>());

vi.mock("@/utils/recordings", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/utils/recordings")>()),
  readRecording,
}));

const { useRecordingsStore } = await import("@/stores/recordingsStore");
const { useTerminalStore } = await import("@/stores/terminalStore");
const { useSessionRecording } = await import("../useSessionRecording");

const session = mockSession({ uid: "session-1" });

function serveRecording(body: string) {
  let requests = 0;
  server.use(
    http.get("*/api/sessions/:uid/records/:seat", ({ params }) => {
      requests++;
      return params.seat === "0"
        ? HttpResponse.text(body)
        : HttpResponse.json({}, { status: 404 });
    }),
  );
  return () => requests;
}

function serveNoRecording() {
  server.use(
    http.get("*/api/sessions/:uid/records/:seat", () =>
      HttpResponse.json({}, { status: 404 }),
    ),
  );
}

function holdInBrowser(logs: string) {
  useRecordingsStore.setState({ recordings: [mockRecordingMeta()] });
  readRecording.mockResolvedValue(logs);
}

const { createObjectURL, revokeObjectURL } = URL;

function captureDownloads() {
  const saved: { blob?: Blob; filename?: string } = {};
  URL.createObjectURL = (blob: Blob) => {
    saved.blob = blob;
    return "blob:recording";
  };
  URL.revokeObjectURL = () => {};
  vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(function (
    this: HTMLAnchorElement,
  ) {
    saved.filename = this.download;
  });
  return saved;
}

async function play(result: {
  current: ReturnType<typeof useSessionRecording>;
}) {
  let played!: boolean;
  await act(async () => {
    played = await result.current.play(session);
  });
  return played;
}

const openRecordings = () =>
  useTerminalStore.getState().recordings.map((r) => [r.id, r.logs, r.shown]);

beforeEach(() => {
  useTerminalStore.setState({ sessions: [], recordings: [] });
  useRecordingsStore.setState({ recordings: [] });
  readRecording.mockReset();
});

afterEach(() => {
  vi.restoreAllMocks();
  URL.createObjectURL = createObjectURL;
  URL.revokeObjectURL = revokeObjectURL;
});

describe("useSessionRecording", () => {
  it("opens the recording of seat 0 in its own tab", async () => {
    serveRecording("asciicast");
    const { result } = renderHook(() => useSessionRecording());

    expect(await play(result)).toBe(true);

    expect(openRecordings()).toEqual([["session-1", "asciicast", true]]);
  });

  it("brings an open recording forward without fetching it again", async () => {
    const requests = serveRecording("asciicast");
    const { result } = renderHook(() => useSessionRecording());
    await play(result);
    useTerminalStore.getState().minimizeAll();

    await play(result);

    expect(requests()).toBe(1);
    expect(openRecordings()).toEqual([["session-1", "asciicast", true]]);
  });

  it("plays the copy the browser holds instead of fetching one", async () => {
    const requests = serveRecording("from-server");
    holdInBrowser("from-browser");
    const { result } = renderHook(() => useSessionRecording());

    await play(result);

    expect(requests()).toBe(0);
    expect(openRecordings()).toEqual([["session-1", "from-browser", true]]);
  });

  it("says so, and opens nothing, when the recording cannot be read", async () => {
    serveNoRecording();
    const { result } = renderHook(() => useSessionRecording());

    const played = await play(result);

    expect(played).toBe(false);
    expect(result.current.error).toBe("Failed to load recording");
    expect(openRecordings()).toEqual([]);
  });

  it("clears the error once a recording plays", async () => {
    serveNoRecording();
    const { result } = renderHook(() => useSessionRecording());
    await play(result);
    serveRecording("asciicast");

    await play(result);

    expect(result.current.error).toBeNull();
  });

  it("clears an earlier error when it brings an open recording forward", async () => {
    serveRecording("asciicast");
    const { result } = renderHook(() => useSessionRecording());
    await play(result);
    serveNoRecording();
    await act(() => result.current.download(session));
    expect(result.current.error).toBe("Failed to download recording");

    await play(result);

    expect(result.current.error).toBeNull();
  });

  it("saves the recording as a .cast named after the device and start time", async () => {
    serveRecording("asciicast");
    const saved = captureDownloads();
    const { result } = renderHook(() => useSessionRecording());

    await act(() =>
      result.current.download(
        mockSession({
          uid: "session-1",
          device: mockDevice({ name: "Web 01" }),
          started_at: "2024-01-01T00:00:00Z",
        }),
      ),
    );

    expect(saved.filename).toBe("shellhub-web-01-20240101-000000.cast");
    expect(await saved.blob?.text()).toBe("asciicast");
    expect(result.current.error).toBeNull();
  });

  it("saves the copy the browser holds instead of fetching one", async () => {
    const requests = serveRecording("from-server");
    holdInBrowser("from-browser");
    const saved = captureDownloads();
    const { result } = renderHook(() => useSessionRecording());

    await act(() => result.current.download(session));

    expect(requests()).toBe(0);
    expect(await saved.blob?.text()).toBe("from-browser");
  });

  it("says so when the recording cannot be downloaded", async () => {
    serveNoRecording();
    const { result } = renderHook(() => useSessionRecording());

    await act(() => result.current.download(session));

    expect(result.current.error).toBe("Failed to download recording");
  });
});
