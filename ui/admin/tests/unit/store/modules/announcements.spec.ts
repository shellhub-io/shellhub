import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { adminApi } from "@/api/http";
import useAnnouncementStore from "@admin/store/modules/announcement";
import { IAdminAnnouncement, IAdminAnnouncementShort, IAdminAnnouncementRequestBody } from "@admin/interfaces/IAnnouncement";
import { buildUrl } from "@tests/utils/url";

const mockAnnouncementShort: IAdminAnnouncementShort = {
  uuid: "52088548-2b99-4f38-ac09-3a8f8988476f",
  title: "This is an announcement",
  date: "2026-01-06T10:00:00.000Z",
};

const mockAnnouncement: IAdminAnnouncement = {
  ...mockAnnouncementShort,
  content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
};

const mockAnnouncementRequestBody: IAdminAnnouncementRequestBody = {
  title: "New announcement",
  content: "## Content here",
};

describe("Admin Announcement Store", () => {
  let announcementStore: ReturnType<typeof useAnnouncementStore>;
  let mockAdminApi: MockAdapter;

  beforeEach(() => {
    setActivePinia(createPinia());
    announcementStore = useAnnouncementStore();
    mockAdminApi = new MockAdapter(adminApi.getAxios());
  });

  afterEach(() => { mockAdminApi.reset(); });

  describe("Initial State", () => {
    it("should have empty announcements array", () => {
      expect(announcementStore.announcements).toEqual([]);
    });

    it("should have empty announcement object", () => {
      expect(announcementStore.announcement).toEqual({});
    });

    it("should have zero announcement count", () => {
      expect(announcementStore.announcementCount).toBe(0);
    });
  });

  describe("createAnnouncement", () => {
    const baseUrl = "http://localhost:3000/admin/api/announcements";

    it("should create announcement successfully and update state", async () => {
      mockAdminApi.onPost(baseUrl, mockAnnouncementRequestBody).reply(201, mockAnnouncement);

      await expect(announcementStore.createAnnouncement(mockAnnouncementRequestBody)).resolves.not.toThrow();

      expect(announcementStore.announcement).toEqual(mockAnnouncement);
    });

    it("should throw on server error when creating announcement", async () => {
      mockAdminApi.onPost(baseUrl, mockAnnouncementRequestBody).reply(500);

      await expect(announcementStore.createAnnouncement(mockAnnouncementRequestBody)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should throw on network error when creating announcement", async () => {
      mockAdminApi.onPost(baseUrl, mockAnnouncementRequestBody).networkError();

      await expect(announcementStore.createAnnouncement(mockAnnouncementRequestBody)).rejects.toThrow("Network Error");
    });
  });

  describe("updateAnnouncement", () => {
    const uuid = "52088548-2b99-4f38-ac09-3a8f8988476f";
    const baseUrl = `http://localhost:3000/admin/api/announcements/${uuid}`;

    it("should update announcement successfully", async () => {
      mockAdminApi.onPut(baseUrl, mockAnnouncementRequestBody).reply(200);

      await expect(announcementStore.updateAnnouncement(uuid, mockAnnouncementRequestBody)).resolves.not.toThrow();
    });

    it("should throw on not found error when updating announcement", async () => {
      mockAdminApi.onPut(baseUrl, mockAnnouncementRequestBody).reply(404, { message: "Announcement not found" });

      await expect(announcementStore.updateAnnouncement(uuid, mockAnnouncementRequestBody)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should throw on network error when updating announcement", async () => {
      mockAdminApi.onPut(baseUrl, mockAnnouncementRequestBody).networkError();

      await expect(announcementStore.updateAnnouncement(uuid, mockAnnouncementRequestBody)).rejects.toThrow("Network Error");
    });
  });

  describe("fetchAnnouncement", () => {
    const uuid = "52088548-2b99-4f38-ac09-3a8f8988476f";
    const baseUrl = `http://localhost:3000/admin/api/announcements/${uuid}`;

    it("should fetch announcement successfully and update state", async () => {
      mockAdminApi.onGet(baseUrl).reply(200, mockAnnouncement);

      await expect(announcementStore.fetchAnnouncement(uuid)).resolves.not.toThrow();

      expect(announcementStore.announcement).toEqual(mockAnnouncement);
    });

    it("should throw on not found error when fetching announcement", async () => {
      mockAdminApi.onGet(baseUrl).reply(404, { message: "Announcement not found" });

      await expect(announcementStore.fetchAnnouncement(uuid)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should throw on network error when fetching announcement", async () => {
      mockAdminApi.onGet(baseUrl).networkError();

      await expect(announcementStore.fetchAnnouncement(uuid)).rejects.toThrow("Network Error");
    });
  });

  describe("fetchAnnouncementList", () => {
    const baseUrl = "http://localhost:3000/admin/api/announcements";

    it("should fetch announcements list successfully with asc ordering", async () => {
      const announcementsList = [mockAnnouncementShort];

      mockAdminApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10", order_by: "asc" }))
        .reply(200, announcementsList, { "x-total-count": "1" });

      await expect(announcementStore.fetchAnnouncementList({ page: 1, perPage: 10, orderBy: "asc" })).resolves.not.toThrow();

      expect(announcementStore.announcements).toEqual(announcementsList);
      expect(announcementStore.announcementCount).toBe(1);
    });

    it("should fetch announcements list successfully with desc ordering", async () => {
      const announcementsList = [
        mockAnnouncementShort,
        { ...mockAnnouncementShort, uuid: "another-uuid", title: "Another announcement" },
      ];

      mockAdminApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10", order_by: "desc" }))
        .reply(200, announcementsList, { "x-total-count": "2" });

      await expect(announcementStore.fetchAnnouncementList({ page: 1, perPage: 10, orderBy: "desc" })).resolves.not.toThrow();

      expect(announcementStore.announcements).toEqual(announcementsList);
      expect(announcementStore.announcementCount).toBe(2);
    });

    it("should fetch empty announcements list successfully", async () => {
      mockAdminApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10", order_by: "asc" }))
        .reply(200, [], { "x-total-count": "0" });

      await expect(announcementStore.fetchAnnouncementList({ page: 1, perPage: 10, orderBy: "asc" })).resolves.not.toThrow();

      expect(announcementStore.announcements).toEqual([]);
      expect(announcementStore.announcementCount).toBe(0);
    });

    it("should throw on server error when fetching announcements list", async () => {
      mockAdminApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10", order_by: "asc" }))
        .reply(500);

      await expect(announcementStore.fetchAnnouncementList({ page: 1, perPage: 10, orderBy: "asc" })).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should throw on network error when fetching announcements list", async () => {
      mockAdminApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10", order_by: "asc" }))
        .networkError();

      await expect(announcementStore.fetchAnnouncementList({ page: 1, perPage: 10, orderBy: "asc" })).rejects.toThrow("Network Error");
    });
  });

  describe("deleteAnnouncement", () => {
    const uuid = "52088548-2b99-4f38-ac09-3a8f8988476f";
    const baseUrl = `http://localhost:3000/admin/api/announcements/${uuid}`;

    it("should delete announcement successfully and update state", async () => {
      mockAdminApi.onDelete(baseUrl).reply(200, mockAnnouncement);

      await expect(announcementStore.deleteAnnouncement(uuid)).resolves.not.toThrow();

      expect(announcementStore.announcement).toEqual(mockAnnouncement);
    });

    it("should throw on not found error when deleting announcement", async () => {
      mockAdminApi.onDelete(baseUrl).reply(403, { message: "Forbidden" });

      await expect(announcementStore.deleteAnnouncement(uuid)).rejects.toBeAxiosErrorWithStatus(403);
    });

    it("should throw on network error when deleting announcement", async () => {
      mockAdminApi.onDelete(baseUrl).networkError();

      await expect(announcementStore.deleteAnnouncement(uuid)).rejects.toThrow("Network Error");
    });
  });
});
