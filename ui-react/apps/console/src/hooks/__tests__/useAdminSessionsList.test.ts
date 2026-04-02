import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement, type ReactNode } from "react";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

vi.mock("@/stores/authStore", () => ({
  useAuthStore: vi.fn(),
}));

vi.mock("@/api/pagination", () => ({
  paginatedQueryFn: vi.fn(),
}));

vi.mock("@/client/@tanstack/react-query.gen", () => ({
  getSessionsAdminQueryKey: vi.fn(() => ["sessions-admin"]),
}));

vi.mock("@/client", () => ({
  getSessionsAdmin: vi.fn(),
}));

import { useAuthStore } from "@/stores/authStore";
import { paginatedQueryFn } from "@/api/pagination";
import { useAdminSessionsList } from "../useAdminSessionsList";

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

function makeWrapper() {
  const qc = new QueryClient({
    defaultOptions: { queries: { retryDelay: 0 } },
  });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: qc }, children);
}

const mockSession = {
  uid: "session-1",
  device_uid: "device-1",
  username: "root",
  ip_address: "192.168.0.1",
  started_at: "2024-01-01T00:00:00Z",
  last_seen: "2024-01-01T01:00:00Z",
  active: true,
  authenticated: true,
};

/* ------------------------------------------------------------------ */
/* Tests                                                               */
/* ------------------------------------------------------------------ */

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useAdminSessionsList", () => {
  describe("when user is not an admin", () => {
    it("returns empty sessions and zero totalCount without fetching", () => {
      vi.mocked(useAuthStore).mockReturnValue(false);
      // paginatedQueryFn should not be called (query is disabled)
      vi.mocked(paginatedQueryFn).mockReturnValue(() => Promise.resolve({ data: [], totalCount: 0 }));

      const { result } = renderHook(
        () => useAdminSessionsList(1, 10),
        { wrapper: makeWrapper() },
      );

      expect(result.current.sessions).toEqual([]);
      expect(result.current.totalCount).toBe(0);
      expect(result.current.isLoading).toBe(false);
      expect(result.current.error).toBeNull();
    });
  });

  describe("when user is an admin", () => {
    beforeEach(() => {
      vi.mocked(useAuthStore).mockReturnValue(true);
    });

    it("returns sessions and totalCount on success", async () => {
      vi.mocked(paginatedQueryFn).mockReturnValue(() =>
        Promise.resolve({ data: [mockSession], totalCount: 1 }),
      );

      const { result } = renderHook(
        () => useAdminSessionsList(1, 10),
        { wrapper: makeWrapper() },
      );

      await waitFor(() => expect(result.current.sessions).toHaveLength(1));

      expect(result.current.sessions[0]).toMatchObject({ uid: "session-1" });
      expect(result.current.totalCount).toBe(1);
      expect(result.current.error).toBeNull();
    });

    it("returns empty arrays when the API returns no sessions", async () => {
      vi.mocked(paginatedQueryFn).mockReturnValue(() =>
        Promise.resolve({ data: [], totalCount: 0 }),
      );

      const { result } = renderHook(
        () => useAdminSessionsList(1, 10),
        { wrapper: makeWrapper() },
      );

      await waitFor(() => expect(result.current.isLoading).toBe(false));

      expect(result.current.sessions).toEqual([]);
      expect(result.current.totalCount).toBe(0);
    });

    it("is loading while the query is in-flight", () => {
      vi.mocked(paginatedQueryFn).mockReturnValue(
        () => new Promise(() => { /* never resolves */ }),
      );

      const { result } = renderHook(
        () => useAdminSessionsList(1, 10),
        { wrapper: makeWrapper() },
      );

      expect(result.current.isLoading).toBe(true);
    });
  });

  describe("error transformation", () => {
    beforeEach(() => {
      vi.mocked(useAuthStore).mockReturnValue(true);
    });

    it("maps a 403 SdkError to a permission-denied message", async () => {
      const sdkError = Object.assign(new Error(), { status: 403, headers: new Headers() });
      vi.mocked(paginatedQueryFn).mockReturnValue(() => Promise.reject(sdkError));

      const { result } = renderHook(
        () => useAdminSessionsList(1, 10),
        { wrapper: makeWrapper() },
      );

      await waitFor(() => expect(result.current.error).not.toBeNull());
      expect(result.current.error?.message).toBe("You don't have permission to view sessions.");
    });

    it("maps a 500 SdkError to a server-error message", async () => {
      const sdkError = Object.assign(new Error(), { status: 500, headers: new Headers() });
      vi.mocked(paginatedQueryFn).mockReturnValue(() => Promise.reject(sdkError));

      const { result } = renderHook(
        () => useAdminSessionsList(1, 10),
        { wrapper: makeWrapper() },
      );

      await waitFor(() => expect(result.current.error).not.toBeNull());
      expect(result.current.error?.message).toBe("Server error. Please try again later.");
    });

    it("includes the status code for unrecognised SDK errors", async () => {
      const sdkError = Object.assign(new Error(), { status: 422, headers: new Headers() });
      vi.mocked(paginatedQueryFn).mockReturnValue(() => Promise.reject(sdkError));

      const { result } = renderHook(
        () => useAdminSessionsList(1, 10),
        { wrapper: makeWrapper() },
      );

      await waitFor(() => expect(result.current.error).not.toBeNull());
      expect(result.current.error?.message).toBe("Failed to load sessions (422).");
    });

    it("preserves the message of a plain Error", async () => {
      vi.mocked(paginatedQueryFn).mockReturnValue(() =>
        Promise.reject(new Error("Network timeout")),
      );

      const { result } = renderHook(
        () => useAdminSessionsList(1, 10),
        { wrapper: makeWrapper() },
      );

      await waitFor(() => expect(result.current.error).not.toBeNull());
      expect(result.current.error?.message).toBe("Network timeout");
    });

    it("returns a generic message for unknown non-Error throws", async () => {
      // eslint-disable-next-line @typescript-eslint/prefer-promise-reject-errors
      vi.mocked(paginatedQueryFn).mockReturnValue(() => Promise.reject("oops"));

      const { result } = renderHook(
        () => useAdminSessionsList(1, 10),
        { wrapper: makeWrapper() },
      );

      await waitFor(() => expect(result.current.error).not.toBeNull());
      expect(result.current.error?.message).toBe("Failed to load sessions.");
    });
  });
});
