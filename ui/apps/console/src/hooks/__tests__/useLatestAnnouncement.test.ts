import { describe, it, expect, vi, beforeEach } from "vitest";
import { waitFor } from "@testing-library/react";
import { renderHookWithClient } from "@/tests/wrapper";
import { getConfig, defaultConfig } from "@/env";
import { mockSdkResponse } from "@/tests/sdk";
import { useLatestAnnouncement } from "../useLatestAnnouncement";
import type { Announcement } from "@/client";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    listAnnouncements: vi.fn(),
    getAnnouncement: vi.fn(),
  }),
);

const mockGetConfig = vi.mocked(getConfig);

function makeAnnouncement(overrides: Partial<Announcement> = {}): Announcement {
  return {
    uuid: "ann-uuid-1",
    title: "Test Announcement",
    content: "## Hello\nSome content",
    date: "2024-06-01T00:00:00Z",
    ...overrides,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig, announcements: true });
  sdk.listAnnouncements.mockReturnValue(new Promise(() => {}));
  sdk.getAnnouncement.mockReturnValue(new Promise(() => {}));
});

describe("useLatestAnnouncement", () => {
  describe("when announcements feature flag is disabled", () => {
    beforeEach(() => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, announcements: false });
    });

    it("returns null announcement immediately", () => {
      const { result } = renderHookWithClient(() => useLatestAnnouncement());

      expect(result.current.announcement).toBeNull();
    });

    it("returns isLoading false", () => {
      const { result } = renderHookWithClient(() => useLatestAnnouncement());

      expect(result.current.isLoading).toBe(false);
    });

    it("does not call the SDK", () => {
      renderHookWithClient(() => useLatestAnnouncement());

      expect(sdk.listAnnouncements).not.toHaveBeenCalled();
    });
  });

  describe("loading state", () => {
    it("returns isLoading true while list query is pending", () => {
      sdk.listAnnouncements.mockReturnValue(new Promise(() => {}));

      const { result } = renderHookWithClient(() => useLatestAnnouncement());

      expect(result.current.isLoading).toBe(true);
    });

    it("returns null announcement while queries are loading", () => {
      const { result } = renderHookWithClient(() => useLatestAnnouncement());

      expect(result.current.announcement).toBeNull();
    });
  });

  describe("when list resolves but is empty", () => {
    it("returns null announcement", async () => {
      sdk.listAnnouncements.mockResolvedValue(mockSdkResponse([]));

      const { result } = renderHookWithClient(() => useLatestAnnouncement());

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.announcement).toBeNull();
    });

    it("does not call getAnnouncement", async () => {
      sdk.listAnnouncements.mockResolvedValue(mockSdkResponse([]));

      const { result } = renderHookWithClient(() => useLatestAnnouncement());

      await waitFor(() => expect(result.current.isLoading).toBe(false));

      expect(sdk.getAnnouncement).not.toHaveBeenCalled();
    });
  });

  describe("when both queries resolve successfully", () => {
    it("returns the full announcement object", async () => {
      const ann = makeAnnouncement({ uuid: "ann-abc", title: "Big Update" });

      sdk.listAnnouncements.mockResolvedValue(
        mockSdkResponse([{ uuid: "ann-abc" }]),
      );
      sdk.getAnnouncement.mockResolvedValue(mockSdkResponse(ann));

      const { result } = renderHookWithClient(() => useLatestAnnouncement());

      await waitFor(() => expect(result.current.announcement).not.toBeNull());
      expect(result.current.announcement).toEqual(ann);
    });

    it("returns isLoading false after both queries settle", async () => {
      const ann = makeAnnouncement();

      sdk.listAnnouncements.mockResolvedValue(
        mockSdkResponse([{ uuid: "ann-uuid-1" }]),
      );
      sdk.getAnnouncement.mockResolvedValue(mockSdkResponse(ann));

      const { result } = renderHookWithClient(() => useLatestAnnouncement());

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.announcement).toEqual(ann);
    });
  });

  describe("when the list query resolves but the detail query is still loading", () => {
    it("returns isLoading true", async () => {
      sdk.listAnnouncements.mockResolvedValue(
        mockSdkResponse([{ uuid: "ann-uuid-1" }]),
      );
      sdk.getAnnouncement.mockReturnValue(new Promise(() => {}));

      const { result } = renderHookWithClient(() => useLatestAnnouncement());

      await waitFor(() => expect(sdk.getAnnouncement).toHaveBeenCalled());

      expect(result.current.isLoading).toBe(true);
    });
  });
});
