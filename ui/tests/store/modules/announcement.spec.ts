import { describe, expect, it, beforeEach, afterEach } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { announcementApi } from "@/api/http";
import useAnnouncementStore from "@/store/modules/announcement";

describe("Announcement Store", () => {
  let mockAnnouncementApi: MockAdapter;
  let store: ReturnType<typeof useAnnouncementStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    mockAnnouncementApi = new MockAdapter(announcementApi.getAxios());
    store = useAnnouncementStore();
  });

  afterEach(() => {
    mockAnnouncementApi.reset();
  });

  describe("Initial State", () => {
    it("should have empty currentAnnouncement object", () => {
      expect(store.currentAnnouncement).toEqual({});
    });
  });

  describe("fetchAnnouncements", () => {
    it("should fetch announcements successfully with default pagination", async () => {
      const mockAnnouncements = [
        {
          uuid: "announcement-1",
          title: "System Maintenance",
          date: "2025-12-18T10:00:00Z",
        },
        {
          uuid: "announcement-2",
          title: "New Features Released",
          date: "2025-12-17T15:30:00Z",
        },
      ];

      mockAnnouncementApi
        .onGet("https://cloud.shellhub.io/api/announcements?page=1&per_page=1&order_by=desc")
        .reply(200, mockAnnouncements);

      const result = await store.fetchAnnouncements();

      expect(result).toEqual(mockAnnouncements);
    });

    it("should fetch announcements with custom pagination", async () => {
      const mockAnnouncements = [
        {
          uuid: "announcement-1",
          title: "System Maintenance",
          date: "2025-12-18T10:00:00Z",
        },
        {
          uuid: "announcement-2",
          title: "New Features Released",
          date: "2025-12-17T15:30:00Z",
        },
        {
          uuid: "announcement-3",
          title: "Security Update",
          date: "2025-12-16T12:00:00Z",
        },
      ];

      mockAnnouncementApi
        .onGet("https://cloud.shellhub.io/api/announcements?page=2&per_page=10&order_by=desc")
        .reply(200, mockAnnouncements);

      const result = await store.fetchAnnouncements({
        page: 2,
        perPage: 10,
      });

      expect(result).toEqual(mockAnnouncements);
    });

    it("should fetch announcements with ascending sort order", async () => {
      const mockAnnouncements = [
        {
          uuid: "announcement-1",
          title: "Oldest Announcement",
          date: "2025-12-01T10:00:00Z",
        },
        {
          uuid: "announcement-2",
          title: "Newer Announcement",
          date: "2025-12-15T10:00:00Z",
        },
      ];

      mockAnnouncementApi
        .onGet("https://cloud.shellhub.io/api/announcements?page=1&per_page=5&order_by=asc")
        .reply(200, mockAnnouncements);

      const result = await store.fetchAnnouncements({
        page: 1,
        perPage: 5,
        sortOrder: "asc",
      });

      expect(result).toEqual(mockAnnouncements);
    });

    it("should handle empty announcements response", async () => {
      mockAnnouncementApi
        .onGet("https://cloud.shellhub.io/api/announcements?page=1&per_page=1&order_by=desc")
        .reply(200, []);

      const result = await store.fetchAnnouncements();

      expect(result).toEqual([]);
    });

    it("should handle null data in response", async () => {
      mockAnnouncementApi
        .onGet("https://cloud.shellhub.io/api/announcements?page=1&per_page=1&order_by=desc")
        .reply(200, null);

      const result = await store.fetchAnnouncements();

      expect(result).toEqual([]);
    });

    it("should handle API errors when fetching announcements", async () => {
      mockAnnouncementApi
        .onGet("https://cloud.shellhub.io/api/announcements?page=1&per_page=1&order_by=desc")
        .reply(500, { message: "Internal Server Error" });

      await expect(
        store.fetchAnnouncements(),
      ).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network errors when fetching announcements", async () => {
      mockAnnouncementApi
        .onGet("https://cloud.shellhub.io/api/announcements?page=1&per_page=1&order_by=desc")
        .networkError();

      await expect(
        store.fetchAnnouncements(),
      ).rejects.toThrow();
    });
  });

  describe("fetchById", () => {
    it("should fetch announcement by UUID successfully", async () => {
      const mockAnnouncement = {
        uuid: "announcement-123",
        title: "Important System Update",
        date: "2025-12-18T10:00:00Z",
        content: "This is the full content of the announcement with detailed information.",
      };

      mockAnnouncementApi
        .onGet("https://cloud.shellhub.io/api/announcements/announcement-123")
        .reply(200, mockAnnouncement);

      await store.fetchById("announcement-123");

      expect(store.currentAnnouncement).toEqual(mockAnnouncement);
    });

    it("should update currentAnnouncement state", async () => {
      const firstAnnouncement = {
        uuid: "announcement-1",
        title: "First Announcement",
        date: "2025-12-17T10:00:00Z",
        content: "First content",
      };

      const secondAnnouncement = {
        uuid: "announcement-2",
        title: "Second Announcement",
        date: "2025-12-18T10:00:00Z",
        content: "Second content",
      };

      mockAnnouncementApi
        .onGet("https://cloud.shellhub.io/api/announcements/announcement-1")
        .reply(200, firstAnnouncement);

      await store.fetchById("announcement-1");
      expect(store.currentAnnouncement).toEqual(firstAnnouncement);

      mockAnnouncementApi
        .onGet("https://cloud.shellhub.io/api/announcements/announcement-2")
        .reply(200, secondAnnouncement);

      await store.fetchById("announcement-2");
      expect(store.currentAnnouncement).toEqual(secondAnnouncement);
    });

    it("should reset currentAnnouncement and throw on not found error", async () => {
      // Set initial announcement
      const initialAnnouncement = {
        uuid: "existing-announcement",
        title: "Existing",
        date: "2025-12-17T10:00:00Z",
        content: "Content",
      };

      mockAnnouncementApi
        .onGet("https://cloud.shellhub.io/api/announcements/existing-announcement")
        .reply(200, initialAnnouncement);

      await store.fetchById("existing-announcement");
      expect(store.currentAnnouncement).toEqual(initialAnnouncement);

      // Try to fetch non-existent announcement
      mockAnnouncementApi
        .onGet("https://cloud.shellhub.io/api/announcements/non-existent")
        .reply(404, { message: "Announcement not found" });

      await expect(
        store.fetchById("non-existent"),
      ).rejects.toBeAxiosErrorWithStatus(404);

      expect(store.currentAnnouncement).toEqual({});
    });

    it("should reset currentAnnouncement and throw on permission error", async () => {
      // Set initial announcement
      store.currentAnnouncement = {
        uuid: "test",
        title: "Test",
        date: "2025-12-18T10:00:00Z",
        content: "Test content",
      };

      mockAnnouncementApi
        .onGet("https://cloud.shellhub.io/api/announcements/restricted")
        .reply(403, { message: "Access denied" });

      await expect(
        store.fetchById("restricted"),
      ).rejects.toBeAxiosErrorWithStatus(403);

      expect(store.currentAnnouncement).toEqual({});
    });

    it("should reset currentAnnouncement and throw on server error", async () => {
      // Set initial announcement
      store.currentAnnouncement = {
        uuid: "test",
        title: "Test",
        date: "2025-12-18T10:00:00Z",
        content: "Test content",
      };

      mockAnnouncementApi
        .onGet("https://cloud.shellhub.io/api/announcements/announcement-123")
        .reply(500, { message: "Internal Server Error" });

      await expect(
        store.fetchById("announcement-123"),
      ).rejects.toBeAxiosErrorWithStatus(500);

      expect(store.currentAnnouncement).toEqual({});
    });

    it("should reset currentAnnouncement and throw on network error", async () => {
      // Set initial announcement
      store.currentAnnouncement = {
        uuid: "test",
        title: "Test",
        date: "2025-12-18T10:00:00Z",
        content: "Test content",
      };

      mockAnnouncementApi
        .onGet("https://cloud.shellhub.io/api/announcements/announcement-123")
        .networkError();

      await expect(
        store.fetchById("announcement-123"),
      ).rejects.toThrow();

      expect(store.currentAnnouncement).toEqual({});
    });
  });
});
