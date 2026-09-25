import { describe, it, expect, beforeEach } from "vitest";
import { act, renderHook } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { useTerminalStore } from "@/stores/terminalStore";
import { useDeleteSessionRecording } from "../useSessionMutations";

beforeEach(() => {
  useTerminalStore.setState({ sessions: [], recordings: [] });
  server.use(
    http.delete(
      "*/api/sessions/:uid/records/:seat",
      () => new HttpResponse(null, { status: 200 }),
    ),
  );
});

describe("useDeleteSessionRecording", () => {
  it("closes the tab playing the recording it deletes", async () => {
    useTerminalStore.getState().openRecording({
      id: "session-1",
      title: "dev",
      logs: "cast",
    });
    const { result } = renderHook(() => useDeleteSessionRecording(), {
      wrapper: createTestWrapper(),
    });

    await act(async () => {
      await result.current.mutateAsync("session-1");
    });

    expect(useTerminalStore.getState().recordings).toEqual([]);
  });
});
