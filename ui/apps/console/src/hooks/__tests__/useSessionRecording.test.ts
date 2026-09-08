import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { useSessionRecording } from "../useSessionRecording";

async function fetchLogs(
  result: { current: ReturnType<typeof useSessionRecording> },
  uid = "session-1",
) {
  let ok!: boolean;
  await act(async () => {
    ok = await result.current.fetchLogs(uid);
  });
  return ok;
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useSessionRecording", () => {
  it("starts with null logs and no loading or error state", () => {
    const { result } = renderHook(() => useSessionRecording());

    expect(result.current.logs).toBeNull();
    expect(result.current.isLoading).toBe(false);
    expect(result.current.error).toBeNull();
  });

  it("reads the recording of seat 0 as text", async () => {
    let capturedUrl = "";
    server.use(
      http.get("*/api/sessions/:uid/records/:seat", ({ request }) => {
        capturedUrl = new URL(request.url).pathname;
        return new HttpResponse("asciicast-content", {
          headers: { "Content-Type": "text/plain" },
        });
      }),
    );

    const { result } = renderHook(() => useSessionRecording());
    await fetchLogs(result);

    expect(capturedUrl).toBe("/api/sessions/session-1/records/0");
  });

  it("sets isLoading true while fetching and false after", async () => {
    let resolveHandler!: (r: Response) => void;
    const handlerReady = new Promise<void>((ready) => {
      server.use(
        http.get(
          "*/api/sessions/:uid/records/:seat",
          () =>
            new Promise<Response>((resolve) => {
              resolveHandler = resolve;
              ready();
            }),
        ),
      );
    });

    const { result } = renderHook(() => useSessionRecording());

    act(() => {
      void result.current.fetchLogs("session-1");
    });
    await handlerReady;
    expect(result.current.isLoading).toBe(true);

    await act(async () => {
      resolveHandler(
        new HttpResponse("log-data", {
          headers: { "Content-Type": "text/plain" },
        }),
      );
    });
    expect(result.current.isLoading).toBe(false);
  });

  it("returns true and stores logs on success", async () => {
    server.use(
      http.get(
        "*/api/sessions/:uid/records/:seat",
        () =>
          new HttpResponse("asciicast-content", {
            headers: { "Content-Type": "text/plain" },
          }),
      ),
    );

    const { result } = renderHook(() => useSessionRecording());
    const ok = await fetchLogs(result);

    expect(ok).toBe(true);
    expect(result.current.logs).toBe("asciicast-content");
    expect(result.current.error).toBeNull();
  });

  it("returns false and sets error on fetch failure", async () => {
    server.use(
      http.get("*/api/sessions/:uid/records/:seat", () =>
        HttpResponse.json({}, { status: 500 }),
      ),
    );

    const { result } = renderHook(() => useSessionRecording());
    const ok = await fetchLogs(result);

    expect(ok).toBe(false);
    expect(result.current.logs).toBeNull();
    expect(result.current.error).toBe("Failed to load recording");
    expect(result.current.isLoading).toBe(false);
  });

  it("clears logsError at the start of a new fetchLogs call", async () => {
    server.use(
      http.get("*/api/sessions/:uid/records/:seat", () =>
        HttpResponse.json({}, { status: 500 }),
      ),
    );

    const { result } = renderHook(() => useSessionRecording());

    await fetchLogs(result);
    expect(result.current.error).toBe("Failed to load recording");

    server.use(
      http.get(
        "*/api/sessions/:uid/records/:seat",
        () =>
          new HttpResponse("log-data", {
            headers: { "Content-Type": "text/plain" },
          }),
      ),
    );

    await fetchLogs(result);
    expect(result.current.error).toBeNull();
    expect(result.current.logs).toBe("log-data");
  });

  it("clearLogs resets logs and error without affecting other state", async () => {
    server.use(
      http.get(
        "*/api/sessions/:uid/records/:seat",
        () =>
          new HttpResponse("asciicast-content", {
            headers: { "Content-Type": "text/plain" },
          }),
      ),
    );

    const { result } = renderHook(() => useSessionRecording());
    await fetchLogs(result);
    expect(result.current.logs).toBe("asciicast-content");

    act(() => {
      result.current.clearLogs();
    });

    expect(result.current.logs).toBeNull();
    expect(result.current.error).toBeNull();
    expect(result.current.isLoading).toBe(false);
  });
});
