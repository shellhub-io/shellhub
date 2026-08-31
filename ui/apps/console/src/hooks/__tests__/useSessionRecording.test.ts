import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useSessionRecording } from "../useSessionRecording";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getSessionRecord: vi.fn(),
  }),
);

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
    sdk.getSessionRecord.mockResolvedValue({ data: "asciicast-content" });

    const { result } = renderHook(() => useSessionRecording());
    await fetchLogs(result);

    expect(sdk.getSessionRecord).toHaveBeenCalledWith({
      path: { uid: "session-1", seat: 0 },
      parseAs: "text",
      throwOnError: true,
    });
  });

  it("sets isLoading true while fetching and false after", async () => {
    let resolve!: (value: unknown) => void;
    sdk.getSessionRecord.mockReturnValue(
      new Promise((r) => {
        resolve = r;
      }),
    );

    const { result } = renderHook(() => useSessionRecording());

    act(() => {
      void result.current.fetchLogs("session-1");
    });
    expect(result.current.isLoading).toBe(true);

    await act(async () => {
      resolve({ data: "log-data" });
    });
    expect(result.current.isLoading).toBe(false);
  });

  it("returns true and stores logs on success", async () => {
    sdk.getSessionRecord.mockResolvedValue({ data: "asciicast-content" });

    const { result } = renderHook(() => useSessionRecording());
    const ok = await fetchLogs(result);

    expect(ok).toBe(true);
    expect(result.current.logs).toBe("asciicast-content");
    expect(result.current.error).toBeNull();
  });

  it("returns false and sets error on fetch failure", async () => {
    sdk.getSessionRecord.mockRejectedValue(new Error("Network error"));

    const { result } = renderHook(() => useSessionRecording());
    const ok = await fetchLogs(result);

    expect(ok).toBe(false);
    expect(result.current.logs).toBeNull();
    expect(result.current.error).toBe("Failed to load recording");
    expect(result.current.isLoading).toBe(false);
  });

  it("clears logsError at the start of a new fetchLogs call", async () => {
    sdk.getSessionRecord.mockRejectedValueOnce(new Error("first error"));
    sdk.getSessionRecord.mockResolvedValue({ data: "log-data" });

    const { result } = renderHook(() => useSessionRecording());

    await fetchLogs(result);
    expect(result.current.error).toBe("Failed to load recording");

    await fetchLogs(result);
    expect(result.current.error).toBeNull();
    expect(result.current.logs).toBe("log-data");
  });

  it("clearLogs resets logs and error without affecting other state", async () => {
    sdk.getSessionRecord.mockResolvedValue({ data: "asciicast-content" });

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
