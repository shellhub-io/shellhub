import { describe, it, expect, vi, beforeEach } from "vitest";
import { waitFor } from "@testing-library/react";
import { renderHookWithClient } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import { useAuthStore } from "@/stores/authStore";
import { useAdminSessions } from "../useAdminSessions";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getSessionsAdmin: vi.fn(),
  }),
);

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
  return renderHookWithClient(() => useAdminSessions({ page: 1, perPage: 10 }));
}

describe("useAdminSessions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("when user is not an admin", () => {
    it("returns empty sessions and zero totalCount without fetching", () => {
      useAuthStore.setState({ isAdmin: false });

      const { result } = renderAdminSessions();

      expect(result.current.sessions).toEqual([]);
      expect(result.current.totalCount).toBe(0);
      expect(result.current.isLoading).toBe(false);
      expect(result.current.error).toBeNull();
      expect(sdk.getSessionsAdmin).not.toHaveBeenCalled();
    });
  });

  describe("when user is an admin", () => {
    beforeEach(() => {
      useAuthStore.setState({ isAdmin: true });
    });

    it("returns sessions and totalCount on success", async () => {
      sdk.getSessionsAdmin.mockResolvedValue(
        mockSdkResponse([mockSession], { "X-Total-Count": "1" }),
      );

      const { result } = renderAdminSessions();

      await waitFor(() => expect(result.current.sessions).toHaveLength(1));

      expect(result.current.sessions[0]).toMatchObject({ uid: "session-1" });
      expect(result.current.totalCount).toBe(1);
      expect(result.current.error).toBeNull();
    });

    it("returns empty arrays when the API returns no sessions", async () => {
      sdk.getSessionsAdmin.mockResolvedValue(
        mockSdkResponse([], { "X-Total-Count": "0" }),
      );

      const { result } = renderAdminSessions();

      await waitFor(() => expect(result.current.isLoading).toBe(false));

      expect(result.current.sessions).toEqual([]);
      expect(result.current.totalCount).toBe(0);
    });

    it("is loading while the query is in-flight", () => {
      sdk.getSessionsAdmin.mockReturnValue(new Promise(() => {}));

      const { result } = renderAdminSessions();

      expect(result.current.isLoading).toBe(true);
    });

    it("exposes the raw error on fetch failure", async () => {
      sdk.getSessionsAdmin.mockRejectedValue(new Error("Network timeout"));

      const { result } = renderAdminSessions();

      await waitFor(() => expect(result.current.error).not.toBeNull());
      expect(result.current.error?.message).toBe("Network timeout");
    });
  });
});
