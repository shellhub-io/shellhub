import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement, type ReactNode } from "react";

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
import { useAdminSessions } from "../useAdminSessions";

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

function renderAdminSessions() {
  return renderHook(
    () => useAdminSessions({ page: 1, perPage: 10 }),
    { wrapper: makeWrapper() },
  );
}

describe("useAdminSessions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("when user is not an admin", () => {
    it("returns empty sessions and zero totalCount without fetching", () => {
      vi.mocked(useAuthStore).mockReturnValue(false);
      // paginatedQueryFn should not be called (query is disabled)
      vi.mocked(paginatedQueryFn).mockReturnValue(() => Promise.resolve({ data: [], totalCount: 0 }));

      const { result } = renderAdminSessions();

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

      const { result } = renderAdminSessions();

      await waitFor(() => expect(result.current.sessions).toHaveLength(1));

      expect(result.current.sessions[0]).toMatchObject({ uid: "session-1" });
      expect(result.current.totalCount).toBe(1);
      expect(result.current.error).toBeNull();
    });

    it("returns empty arrays when the API returns no sessions", async () => {
      vi.mocked(paginatedQueryFn).mockReturnValue(() =>
        Promise.resolve({ data: [], totalCount: 0 }),
      );

      const { result } = renderAdminSessions();

      await waitFor(() => expect(result.current.isLoading).toBe(false));

      expect(result.current.sessions).toEqual([]);
      expect(result.current.totalCount).toBe(0);
    });

    it("is loading while the query is in-flight", () => {
      vi.mocked(paginatedQueryFn).mockReturnValue(
        () => new Promise(() => { /* never resolves */ }),
      );

      const { result } = renderAdminSessions();

      expect(result.current.isLoading).toBe(true);
    });

    it("exposes the raw error on fetch failure", async () => {
      vi.mocked(paginatedQueryFn).mockReturnValue(() =>
        Promise.reject(new Error("Network timeout")),
      );

      const { result } = renderAdminSessions();

      await waitFor(() => expect(result.current.error).not.toBeNull());
      expect(result.current.error?.message).toBe("Network timeout");
    });
  });
});
