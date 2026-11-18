import { describe, it, expect, beforeEach, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockRecordingMeta, mockSession } from "@/tests/factories";

vi.mock("@/utils/recordings", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/utils/recordings")>()),
  deleteRecording: () => Promise.resolve(),
}));

const { useRecordingsStore } = await import("@/stores/recordingsStore");
const { useTerminalStore } = await import("@/stores/terminalStore");
const { recordingId } = await import("@/utils/recordings");
const { useRemoveRecording } = await import("../useRemoveRecording");

function serveDelete(status = 200, failingSeat?: number) {
  const deleted: string[] = [];
  server.use(
    http.delete("*/api/sessions/:uid/records/:seat", ({ params }) => {
      deleted.push(`${String(params.uid)}/${String(params.seat)}`);
      const failed = String(params.seat) === String(failingSeat);
      return new HttpResponse(null, { status: failed ? 500 : status });
    }),
  );
  return deleted;
}

function serveRecordedAfterEachDelete(...recorded: boolean[]) {
  const answers = [...recorded];
  server.use(
    http.get("*/api/sessions/:uid", () =>
      HttpResponse.json(mockSession({ recorded: answers.shift() ?? false })),
    ),
  );
}

function holdInBrowser() {
  useRecordingsStore.setState({ recordings: [mockRecordingMeta()] });
}

const heldInBrowser = () =>
  useRecordingsStore.getState().recordings.map((r) => r.sessionUid);

const openTabs = () => useTerminalStore.getState().recordings.map((r) => r.id);

function openTab(seat: number) {
  useTerminalStore.getState().openRecording({
    id: recordingId("session-1", seat),
    sessionUid: "session-1",
    seat,
    title: "dev",
    logs: "cast",
    filename: "dev.cast",
    recorded: true,
  });
}

function renderRemove() {
  return renderHook(() => useRemoveRecording(), {
    wrapper: createTestWrapper(),
  }).result;
}

beforeEach(() => {
  useRecordingsStore.setState({ recordings: [] });
  useTerminalStore.setState({ sessions: [], recordings: [] });
  openTab(0);
});

describe("useRemoveRecording", () => {
  it("deletes the copy the browser holds and closes its tab", async () => {
    const deleted = serveDelete();
    holdInBrowser();
    const remove = renderRemove();

    await act(() => remove.current("session-1", false));

    expect(heldInBrowser()).toEqual([]);
    expect(deleted).toEqual([]);
    expect(openTabs()).toEqual([]);
  });

  it("deletes the server's copy when the session was recorded", async () => {
    const deleted = serveDelete();
    const remove = renderRemove();

    await act(() => remove.current("session-1", true));

    expect(deleted).toEqual(["session-1/0"]);
    expect(openTabs()).toEqual([]);
  });

  it("deletes both copies when the browser and the server hold one", async () => {
    const deleted = serveDelete();
    holdInBrowser();
    const remove = renderRemove();

    await act(() => remove.current("session-1", true));

    expect(heldInBrowser()).toEqual([]);
    expect(deleted).toEqual(["session-1/0"]);
  });

  it("rejects, leaving the tab open, when the server's copy cannot be deleted", async () => {
    serveDelete(500);
    const remove = renderRemove();

    await expect(remove.current("session-1", true)).rejects.toThrow();

    expect(openTabs()).toEqual(["session-1"]);
  });

  it("rejects with the browser's copy gone and the tab open when the server's copy remains", async () => {
    serveDelete(500);
    holdInBrowser();
    const remove = renderRemove();

    await expect(remove.current("session-1", true)).rejects.toThrow();

    expect(heldInBrowser()).toEqual([]);
    expect(openTabs()).toEqual(["session-1"]);
  });

  it("deletes every seat it is given while the session stays recorded, closing each tab", async () => {
    const deleted = serveDelete();
    serveRecordedAfterEachDelete(true);
    openTab(1);
    const remove = renderRemove();

    await act(() => remove.current("session-1", true, [0, 1]));

    expect(deleted).toEqual(["session-1/0", "session-1/1"]);
    expect(openTabs()).toEqual([]);
  });

  it("stops once the session has no recording left", async () => {
    const deleted = serveDelete();
    serveRecordedAfterEachDelete(true, false);
    const remove = renderRemove();

    await act(() => remove.current("session-1", true, [0, 1, 2]));

    expect(deleted).toEqual(["session-1/0", "session-1/1"]);
  });

  it("rejects at the first seat that cannot be deleted, skipping the rest", async () => {
    const deleted = serveDelete(500);
    serveRecordedAfterEachDelete(true);
    openTab(1);
    const remove = renderRemove();

    await expect(remove.current("session-1", true, [0, 1])).rejects.toThrow();

    expect(deleted).toEqual(["session-1/0"]);
    expect(openTabs()).toEqual(["session-1", "session-1#1"]);
  });

  it("rejects at a later seat that cannot be deleted, with the earlier seats' tabs closed", async () => {
    const deleted = serveDelete(200, 1);
    serveRecordedAfterEachDelete(true);
    openTab(1);
    const remove = renderRemove();

    await expect(remove.current("session-1", true, [0, 1])).rejects.toThrow();

    expect(deleted).toEqual(["session-1/0", "session-1/1"]);
    expect(openTabs()).toEqual(["session-1#1"]);
  });

  it("deletes only the seat it is given, leaving the other tabs open", async () => {
    const deleted = serveDelete();
    openTab(1);
    const remove = renderRemove();

    await act(() => remove.current("session-1", true, [1]));

    expect(deleted).toEqual(["session-1/1"]);
    expect(openTabs()).toEqual(["session-1"]);
  });

  it("deletes seat zero, both copies, when it is given no seats", async () => {
    const deleted = serveDelete();
    holdInBrowser();
    const remove = renderRemove();

    await act(() => remove.current("session-1", true, []));

    expect(deleted).toEqual(["session-1/0"]);
    expect(heldInBrowser()).toEqual([]);
  });

  it("keeps the browser's copy when seat zero is not among the seats", async () => {
    serveDelete();
    holdInBrowser();
    const remove = renderRemove();

    await act(() => remove.current("session-1", true, [1]));

    expect(heldInBrowser()).toEqual(["session-1"]);
  });
});
