import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useSessionRecording } from "../useSessionRecording";
import apiClient from "../../api/client";

vi.mock("../../api/client", () => ({
  default: {
    get: vi.fn(),
    delete: vi.fn(),
  },
}));

const mockGet = vi.mocked(apiClient.get);

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

  it("sets isLoading true while fetching and false after", async () => {
    let resolve!: (value: unknown) => void;
    mockGet.mockReturnValue(new Promise((r) => { resolve = r; }) as never);

    const { result } = renderHook(() => useSessionRecording());

    act(() => { void result.current.fetchLogs("session-1"); });
    expect(result.current.isLoading).toBe(true);

    await act(async () => { resolve({ data: "log-data" }); });
    expect(result.current.isLoading).toBe(false);
  });

  it("returns true and stores logs on success", async () => {
    mockGet.mockResolvedValue({ data: "asciicast-content" });

    const { result } = renderHook(() => useSessionRecording());
    let ok!: boolean;
    await act(async () => { ok = await result.current.fetchLogs("session-1"); });

    expect(ok).toBe(true);
    expect(result.current.logs).toBe("asciicast-content");
    expect(result.current.error).toBeNull();
  });

  it("returns false and sets error on fetch failure", async () => {
    mockGet.mockRejectedValue(new Error("Network error"));

    const { result } = renderHook(() => useSessionRecording());
    let ok!: boolean;
    await act(async () => { ok = await result.current.fetchLogs("session-1"); });

    expect(ok).toBe(false);
    expect(result.current.logs).toBeNull();
    expect(result.current.error).toBe("Failed to load recording");
    expect(result.current.isLoading).toBe(false);
  });

  it("clears logsError at the start of a new fetchLogs call", async () => {
    mockGet.mockRejectedValueOnce(new Error("first error"));
    mockGet.mockResolvedValue({ data: "log-data" });

    const { result } = renderHook(() => useSessionRecording());

    await act(async () => { await result.current.fetchLogs("session-1"); });
    expect(result.current.error).toBe("Failed to load recording");

    await act(async () => { await result.current.fetchLogs("session-1"); });
    expect(result.current.error).toBeNull();
    expect(result.current.logs).toBe("log-data");
  });

  it("clearLogs resets logs and error without affecting other state", async () => {
    mockGet.mockResolvedValue({ data: "asciicast-content" });

    const { result } = renderHook(() => useSessionRecording());
    await act(async () => { await result.current.fetchLogs("session-1"); });
    expect(result.current.logs).toBe("asciicast-content");

    act(() => { result.current.clearLogs(); });

    expect(result.current.logs).toBeNull();
    expect(result.current.error).toBeNull();
    expect(result.current.isLoading).toBe(false);
  });
});
