import { describe, it, expect, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { mockSession } from "@/tests/factories";
import { useTerminalStore } from "@/stores/terminalStore";
import { useSessionRecording } from "../useSessionRecording";

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

async function play(
  result: { current: ReturnType<typeof useSessionRecording> },
  readLocal?: () => Promise<string>,
) {
  let played!: boolean;
  await act(async () => {
    played = await result.current.play(session, readLocal);
  });
  return played;
}

const openRecordings = () =>
  useTerminalStore.getState().recordings.map((r) => [r.id, r.logs, r.shown]);

beforeEach(() => {
  useTerminalStore.setState({ sessions: [], recordings: [] });
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
    const { result } = renderHook(() => useSessionRecording());

    await play(result, () => Promise.resolve("from-browser"));

    expect(requests()).toBe(0);
    expect(openRecordings()).toEqual([["session-1", "from-browser", true]]);
  });

  it("says so, and opens nothing, when the recording cannot be read", async () => {
    const { result } = renderHook(() => useSessionRecording());

    const played = await play(result, () => Promise.reject(new Error("gone")));

    expect(played).toBe(false);
    expect(result.current.error).toBe("Failed to load recording");
    expect(openRecordings()).toEqual([]);
  });

  it("clears the error once a recording plays", async () => {
    const { result } = renderHook(() => useSessionRecording());
    await play(result, () => Promise.reject(new Error("gone")));
    serveRecording("asciicast");

    await play(result);

    expect(result.current.error).toBeNull();
  });

  it("clears an earlier error when it brings an open recording forward", async () => {
    serveRecording("asciicast");
    const { result } = renderHook(() => useSessionRecording());
    await play(result);
    await play(result, () => Promise.reject(new Error("gone")));

    await play(result);

    expect(result.current.error).toBeNull();
  });
});
