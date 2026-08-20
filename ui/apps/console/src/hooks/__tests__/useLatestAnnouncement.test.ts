import { describe, it, expect, vi, beforeEach } from "vitest";
import { waitFor } from "@testing-library/react";
import { renderHookWithClient } from "@/tests/wrapper";
import { useLatestAnnouncement } from "../useLatestAnnouncement";
import type { Announcement } from "@/client";

// Mock the generated query-option factories so we control what queryFn runs
vi.mock("@/client", () => ({
  listAnnouncementsOptions: vi.fn(),
  getAnnouncementOptions: vi.fn(),
}));

import { getConfig, defaultConfig } from "@/env";
import { listAnnouncementsOptions, getAnnouncementOptions } from "@/client";

const mockGetConfig = vi.mocked(getConfig);
const mockListAnnouncementsOptions = vi.mocked(listAnnouncementsOptions);
const mockGetAnnouncementOptions = vi.mocked(getAnnouncementOptions);

// ── Helpers ──────────────────────────────────────────────────────────────────

function makeAnnouncement(overrides: Partial<Announcement> = {}): Announcement {
  return {
    uuid: "ann-uuid-1",
    title: "Test Announcement",
    content: "## Hello\nSome content",
    date: "2024-06-01T00:00:00Z",
    ...overrides,
  };
}

/** Returns queryOptions-shaped objects that delegate to the provided fn. */
function makeListOptions(fn: () => unknown) {
  return {
    queryKey: ["listAnnouncements"],
    queryFn: fn,
  };
}

function makeDetailOptions(uuid: string, fn: () => unknown) {
  return {
    queryKey: ["getAnnouncement", uuid],
    queryFn: fn,
  };
}

// ── Setup ─────────────────────────────────────────────────────────────────────

beforeEach(() => {
  vi.clearAllMocks();

  // Default: feature enabled
  mockGetConfig.mockReturnValue({ ...defaultConfig, announcements: true });

  // Default: both option factories return never-resolving promises (loading state)
  mockListAnnouncementsOptions.mockReturnValue(
    makeListOptions(() => new Promise(() => {})) as never,
  );
  mockGetAnnouncementOptions.mockReturnValue(
    makeDetailOptions("", () => new Promise(() => {})) as never,
  );
});

// ── Tests ─────────────────────────────────────────────────────────────────────

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

    it("does not call listAnnouncementsOptions queryFn", () => {
      const listFn = vi.fn().mockResolvedValue([]);
      mockListAnnouncementsOptions.mockReturnValue(
        makeListOptions(listFn) as never,
      );

      renderHookWithClient(() => useLatestAnnouncement());

      expect(listFn).not.toHaveBeenCalled();
    });
  });

  describe("loading state", () => {
    it("returns isLoading true while list query is pending", () => {
      mockListAnnouncementsOptions.mockReturnValue(
        makeListOptions(() => new Promise(() => {})) as never,
      );

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
      mockListAnnouncementsOptions.mockReturnValue(
        makeListOptions(() => Promise.resolve([])) as never,
      );

      const { result } = renderHookWithClient(() => useLatestAnnouncement());

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.announcement).toBeNull();
    });

    it("does not call getAnnouncementOptions queryFn", async () => {
      const detailFn = vi.fn().mockResolvedValue(makeAnnouncement());
      mockListAnnouncementsOptions.mockReturnValue(
        makeListOptions(() => Promise.resolve([])) as never,
      );
      mockGetAnnouncementOptions.mockReturnValue(
        makeDetailOptions("", detailFn) as never,
      );

      const { result } = renderHookWithClient(() => useLatestAnnouncement());

      await waitFor(() => expect(result.current.isLoading).toBe(false));

      expect(detailFn).not.toHaveBeenCalled();
    });
  });

  describe("when both queries resolve successfully", () => {
    it("returns the full announcement object", async () => {
      const ann = makeAnnouncement({ uuid: "ann-abc", title: "Big Update" });

      mockListAnnouncementsOptions.mockReturnValue(
        makeListOptions(() => Promise.resolve([{ uuid: "ann-abc" }])) as never,
      );
      mockGetAnnouncementOptions.mockReturnValue(
        makeDetailOptions("ann-abc", () => Promise.resolve(ann)) as never,
      );

      const { result } = renderHookWithClient(() => useLatestAnnouncement());

      await waitFor(() => expect(result.current.announcement).not.toBeNull());
      expect(result.current.announcement).toEqual(ann);
    });

    it("returns isLoading false after both queries settle", async () => {
      const ann = makeAnnouncement();

      mockListAnnouncementsOptions.mockReturnValue(
        makeListOptions(() =>
          Promise.resolve([{ uuid: "ann-uuid-1" }]),
        ) as never,
      );
      mockGetAnnouncementOptions.mockReturnValue(
        makeDetailOptions("ann-uuid-1", () => Promise.resolve(ann)) as never,
      );

      const { result } = renderHookWithClient(() => useLatestAnnouncement());

      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.announcement).toEqual(ann);
    });
  });

  describe("when the list query resolves but the detail query is still loading", () => {
    it("returns isLoading true", async () => {
      mockListAnnouncementsOptions.mockReturnValue(
        makeListOptions(() =>
          Promise.resolve([{ uuid: "ann-uuid-1" }]),
        ) as never,
      );
      // detail never resolves
      mockGetAnnouncementOptions.mockReturnValue(
        makeDetailOptions("ann-uuid-1", () => new Promise(() => {})) as never,
      );

      const { result } = renderHookWithClient(() => useLatestAnnouncement());

      // Let the list query settle so latestUuid is set and the detail query is enabled
      await waitFor(() =>
        expect(mockGetAnnouncementOptions).toHaveBeenCalledWith(
          expect.objectContaining({ path: { uuid: "ann-uuid-1" } }),
        ),
      );

      expect(result.current.isLoading).toBe(true);
    });
  });
});
