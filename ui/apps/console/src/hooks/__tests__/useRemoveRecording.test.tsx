import { describe, it, expect, beforeEach, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockRecordingMeta } from "@/tests/factories";

vi.mock("@/utils/recordings", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/utils/recordings")>()),
  deleteRecording: () => Promise.resolve(),
}));

const { useRecordingsStore } = await import("@/stores/recordingsStore");
const { useTerminalStore } = await import("@/stores/terminalStore");
const { useRemoveRecording } = await import("../useRemoveRecording");

function serveDelete(status = 200) {
  const deleted: string[] = [];
  server.use(
    http.delete("*/api/sessions/:uid/records/:seat", ({ params }) => {
      deleted.push(String(params.uid));
      return new HttpResponse(null, { status });
    }),
  );
  return deleted;
}

function holdInBrowser() {
  useRecordingsStore.setState({ recordings: [mockRecordingMeta()] });
}

const heldInBrowser = () =>
  useRecordingsStore.getState().recordings.map((r) => r.sessionUid);

const openTabs = () => useTerminalStore.getState().recordings.map((r) => r.id);

function renderRemove() {
  return renderHook(() => useRemoveRecording(), {
    wrapper: createTestWrapper(),
  }).result;
}

beforeEach(() => {
  useRecordingsStore.setState({ recordings: [] });
  useTerminalStore.setState({ sessions: [], recordings: [] });
  useTerminalStore.getState().openRecording({
    id: "session-1",
    title: "dev",
    logs: "cast",
    filename: "dev.cast",
    recorded: true,
  });
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

    expect(deleted).toEqual(["session-1"]);
    expect(openTabs()).toEqual([]);
  });

  it("deletes both copies when the browser and the server hold one", async () => {
    const deleted = serveDelete();
    holdInBrowser();
    const remove = renderRemove();

    await act(() => remove.current("session-1", true));

    expect(heldInBrowser()).toEqual([]);
    expect(deleted).toEqual(["session-1"]);
  });

  it("rejects, leaving the tab open, when the server's copy cannot be deleted", async () => {
    serveDelete(500);
    const remove = renderRemove();

    await expect(remove.current("session-1", true)).rejects.toThrow();

    expect(openTabs()).toEqual(["session-1"]);
  });
});
