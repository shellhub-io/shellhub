import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useSessionRecording } from "../useSessionRecording";

// getSessionRecord with parseAs: "text" returns { data: string } at runtime,
// but the generated type says { data: RecordedSessionResponse }. An untyped
// vi.fn() avoids fighting that mismatch.
const mockGetSessionRecord = vi.hoisted(() => vi.fn());

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return { ...actual, getSessionRecord: mockGetSessionRecord };
});

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
    mockGetSessionRecord.mockResolvedValue({ data: "asciicast-content" });

    const { result } = renderHook(() => useSessionRecording());
    await fetchLogs(result);

    expect(mockGetSessionRecord).toHaveBeenCalledWith({
      path: { uid: "session-1", seat: 0 },
      parseAs: "text",
      throwOnError: true,
    });
  });

  it("sets isLoading true while fetching and false after", async () => {
    let resolve!: (value: unknown) => void;
    mockGetSessionRecord.mockReturnValue(
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
    mockGetSessionRecord.mockResolvedValue({ data: "asciicast-content" });

    const { result } = renderHook(() => useSessionRecording());
    const ok = await fetchLogs(result);

    expect(ok).toBe(true);
    expect(result.current.logs).toBe("asciicast-content");
    expect(result.current.error).toBeNull();
  });

  it("returns false and sets error on fetch failure", async () => {
    mockGetSessionRecord.mockRejectedValue(new Error("Network error"));

    const { result } = renderHook(() => useSessionRecording());
    const ok = await fetchLogs(result);

    expect(ok).toBe(false);
    expect(result.current.logs).toBeNull();
    expect(result.current.error).toBe("Failed to load recording");
    expect(result.current.isLoading).toBe(false);
  });

  it("clears logsError at the start of a new fetchLogs call", async () => {
    mockGetSessionRecord.mockRejectedValueOnce(new Error("first error"));
    mockGetSessionRecord.mockResolvedValue({ data: "log-data" });

    const { result } = renderHook(() => useSessionRecording());

    await fetchLogs(result);
    expect(result.current.error).toBe("Failed to load recording");

    await fetchLogs(result);
    expect(result.current.error).toBeNull();
    expect(result.current.logs).toBe("log-data");
  });

  it("clearLogs resets logs and error without affecting other state", async () => {
    mockGetSessionRecord.mockResolvedValue({ data: "asciicast-content" });

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
