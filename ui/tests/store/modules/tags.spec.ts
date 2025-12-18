import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { tagsApi } from "@/api/http";
import useTagsStore from "@/store/modules/tags";
import type { ITag } from "@/interfaces/ITags";
import { buildUrl } from "../../utils/url";

const mockTagBase: ITag = {
  name: "production",
  tenant_id: "tenant-id-123",
  created_at: "2025-01-01T00:00:00.000Z",
  updated_at: "2025-01-01T00:00:00.000Z",
};

describe("Tags Store", () => {
  let tagsStore: ReturnType<typeof useTagsStore>;
  let mockTagsApi: MockAdapter;

  beforeEach(() => {
    setActivePinia(createPinia());
    tagsStore = useTagsStore();
    mockTagsApi = new MockAdapter(tagsApi.getAxios());
  });

  afterEach(() => { mockTagsApi.reset(); });

  describe("Initial State", () => {
    it("should have empty tags array", () => {
      expect(tagsStore.tags).toEqual([]);
    });

    it("should have zero tag count", () => {
      expect(tagsStore.tagCount).toBe(0);
    });

    it("should have showTags as false", () => {
      expect(tagsStore.showTags).toBe(false);
    });

    it("should have empty selected tags array", () => {
      expect(tagsStore.selectedTags).toEqual([]);
    });
  });

  describe("fetchTagList", () => {
    const baseUrl = "http://localhost:3000/api/tags";

    it("should fetch tags successfully with default pagination", async () => {
      const tagList = [mockTagBase];

      mockTagsApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .reply(200, tagList, { "x-total-count": "1" });

      await expect(tagsStore.fetchTagList()).resolves.not.toThrow();

      expect(tagsStore.tags).toEqual(tagList);
      expect(tagsStore.tagCount).toBe(1);
    });

    it("should fetch tags successfully with custom pagination", async () => {
      const tagList = [mockTagBase];

      mockTagsApi
        .onGet(buildUrl(baseUrl, { page: "2", per_page: "20" }))
        .reply(200, tagList, { "x-total-count": "5" });

      await expect(tagsStore.fetchTagList({ page: 2, perPage: 20 })).resolves.not.toThrow();

      expect(tagsStore.tags).toEqual(tagList);
      expect(tagsStore.tagCount).toBe(5);
    });

    it("should fetch tags successfully with filter", async () => {
      const tagList = [mockTagBase];
      const filter = "prod";

      mockTagsApi
        .onGet(buildUrl(baseUrl, { filter: filter, page: "1", per_page: "10" }))
        .reply(200, tagList, { "x-total-count": "1" });

      await expect(tagsStore.fetchTagList({ filter, page: 1, perPage: 10 })).resolves.not.toThrow();

      expect(tagsStore.tags).toEqual(tagList);
    });

    it("should reset state and throw on not found error when fetching tags", async () => {
      mockTagsApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .reply(404, { message: "Tags not found" });

      await expect(tagsStore.fetchTagList()).rejects.toBeAxiosErrorWithStatus(404);

      expect(tagsStore.tags).toEqual([]);
      expect(tagsStore.tagCount).toBe(0);
    });

    it("should reset state and throw on server error when fetching tags", async () => {
      mockTagsApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .reply(500);

      await expect(tagsStore.fetchTagList()).rejects.toBeAxiosErrorWithStatus(500);

      expect(tagsStore.tags).toEqual([]);
      expect(tagsStore.tagCount).toBe(0);
    });

    it("should reset state and throw on network error when fetching tags", async () => {
      mockTagsApi
        .onGet(buildUrl(baseUrl, { page: "1", per_page: "10" }))
        .networkError();

      await expect(tagsStore.fetchTagList()).rejects.toThrow("Network Error");

      expect(tagsStore.tags).toEqual([]);
      expect(tagsStore.tagCount).toBe(0);
    });
  });

  describe("setTagListVisibility", () => {
    const visibilityUrl = "http://localhost:3000/api/tags?page=1&per_page=1";

    it("should set showTags to true when tags exist", async () => {
      mockTagsApi
        .onGet(visibilityUrl)
        .reply(200, [mockTagBase], { "x-total-count": "5" });

      await expect(tagsStore.setTagListVisibility()).resolves.not.toThrow();

      expect(tagsStore.showTags).toBe(true);
    });

    it("should not set showTags when no tags exist", async () => {
      mockTagsApi
        .onGet(visibilityUrl)
        .reply(200, [], { "x-total-count": "0" });

      await expect(tagsStore.setTagListVisibility()).resolves.not.toThrow();

      expect(tagsStore.showTags).toBe(false);
    });

    it("should handle server error when checking tag visibility", async () => {
      mockTagsApi
        .onGet(visibilityUrl)
        .reply(500);

      await expect(tagsStore.setTagListVisibility()).rejects.toBeAxiosErrorWithStatus(500);
    });
  });

  describe("toggleSelectedTag", () => {
    it("should add tag to selected tags", () => {
      const tag = mockTagBase;

      tagsStore.toggleSelectedTag(tag);

      expect(tagsStore.selectedTags).toEqual([tag]);
    });

    it("should remove tag from selected tags when already selected", () => {
      const tag = mockTagBase;

      tagsStore.toggleSelectedTag(tag);
      expect(tagsStore.selectedTags).toEqual([tag]);

      tagsStore.toggleSelectedTag(tag);
      expect(tagsStore.selectedTags).toEqual([]);
    });

    it("should toggle multiple tags correctly", () => {
      const tag1 = mockTagBase;
      const tag2 = { ...mockTagBase, name: "staging" };
      const tag3 = { ...mockTagBase, name: "development" };

      tagsStore.toggleSelectedTag(tag1);
      tagsStore.toggleSelectedTag(tag2);
      tagsStore.toggleSelectedTag(tag3);

      expect(tagsStore.selectedTags).toEqual([tag1, tag2, tag3]);

      tagsStore.toggleSelectedTag(tag2);
      expect(tagsStore.selectedTags).toEqual([tag1, tag3]);

      tagsStore.toggleSelectedTag(tag1);
      expect(tagsStore.selectedTags).toEqual([tag3]);

      tagsStore.toggleSelectedTag(tag3);
      expect(tagsStore.selectedTags).toEqual([]);
    });

    it("should handle tag as string", () => {
      const tag = "production" as unknown as ITag;

      tagsStore.toggleSelectedTag(tag);

      expect(tagsStore.selectedTags).toContain(tag);
    });

    it("should clear selected tags directly", () => {
      tagsStore.toggleSelectedTag(mockTagBase);
      expect(tagsStore.selectedTags).toHaveLength(1);

      tagsStore.selectedTags = [];

      expect(tagsStore.selectedTags).toEqual([]);
    });
  });

  describe("createTag", () => {
    const createUrl = "http://localhost:3000/api/tags";

    it("should create tag successfully", async () => {
      const tagName = "new-tag";

      mockTagsApi
        .onPost(createUrl)
        .reply(201);

      await expect(tagsStore.createTag(tagName)).resolves.not.toThrow();
    });

    it("should handle validation error when creating tag", async () => {
      const tagName = "invalid tag name";

      mockTagsApi
        .onPost(createUrl)
        .reply(400, { message: "Invalid tag name" });

      await expect(tagsStore.createTag(tagName)).rejects.toBeAxiosErrorWithStatus(400);
    });

    it("should handle server error when creating tag", async () => {
      const tagName = "new-tag";

      mockTagsApi
        .onPost(createUrl)
        .reply(500);

      await expect(tagsStore.createTag(tagName)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when creating tag", async () => {
      const tagName = "new-tag";

      mockTagsApi
        .onPost(createUrl)
        .networkError();

      await expect(tagsStore.createTag(tagName)).rejects.toThrow("Network Error");
    });
  });

  describe("updateTag", () => {
    it("should update tag successfully", async () => {
      const currentName = "old-tag";
      const newName = { name: "new-tag" };

      mockTagsApi
        .onPatch(`http://localhost:3000/api/tags/${currentName}`)
        .reply(200);

      await expect(tagsStore.updateTag(currentName, newName)).resolves.not.toThrow();
    });

    it("should handle not found error when updating tag", async () => {
      const currentName = "non-existent-tag";
      const newName = { name: "new-tag" };

      mockTagsApi
        .onPatch(`http://localhost:3000/api/tags/${currentName}`)
        .reply(404, { message: "Tag not found" });

      await expect(tagsStore.updateTag(currentName, newName)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when updating tag", async () => {
      const currentName = "old-tag";
      const newName = { name: "new-tag" };

      mockTagsApi
        .onPatch(`http://localhost:3000/api/tags/${currentName}`)
        .reply(500);

      await expect(tagsStore.updateTag(currentName, newName)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when updating tag", async () => {
      const currentName = "old-tag";
      const newName = { name: "new-tag" };

      mockTagsApi
        .onPatch(`http://localhost:3000/api/tags/${currentName}`)
        .networkError();

      await expect(tagsStore.updateTag(currentName, newName)).rejects.toThrow("Network Error");
    });
  });

  describe("deleteTag", () => {
    it("should delete tag successfully", async () => {
      const tagName = "tag-to-delete";

      mockTagsApi
        .onDelete(`http://localhost:3000/api/tags/${tagName}`)
        .reply(200);

      await expect(tagsStore.deleteTag(tagName)).resolves.not.toThrow();
    });

    it("should handle not found error when deleting tag", async () => {
      const tagName = "non-existent-tag";

      mockTagsApi
        .onDelete(`http://localhost:3000/api/tags/${tagName}`)
        .reply(404, { message: "Tag not found" });

      await expect(tagsStore.deleteTag(tagName)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when deleting tag", async () => {
      const tagName = "tag-to-delete";

      mockTagsApi
        .onDelete(`http://localhost:3000/api/tags/${tagName}`)
        .reply(500);

      await expect(tagsStore.deleteTag(tagName)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when deleting tag", async () => {
      const tagName = "tag-to-delete";

      mockTagsApi
        .onDelete(`http://localhost:3000/api/tags/${tagName}`)
        .networkError();

      await expect(tagsStore.deleteTag(tagName)).rejects.toThrow("Network Error");
    });
  });

  describe("addTagToDevice", () => {
    const addTagUrl = (deviceUid: string, tagName: string) => `http://localhost:3000/api/devices/${deviceUid}/tags/${tagName}`;

    it("should add tag to device successfully", async () => {
      const deviceUid = "device-uid-123";
      const tagName = "production";

      mockTagsApi
        .onPost(addTagUrl(deviceUid, tagName))
        .reply(200);

      await expect(tagsStore.addTagToDevice(deviceUid, tagName)).resolves.not.toThrow();
    });

    it("should handle not found error when adding tag to device", async () => {
      const deviceUid = "non-existent-device";
      const tagName = "production";

      mockTagsApi
        .onPost(addTagUrl(deviceUid, tagName))
        .reply(404, { message: "Device not found" });

      await expect(tagsStore.addTagToDevice(deviceUid, tagName)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when adding tag to device", async () => {
      const deviceUid = "device-uid-123";
      const tagName = "production";

      mockTagsApi
        .onPost(addTagUrl(deviceUid, tagName))
        .reply(500);

      await expect(tagsStore.addTagToDevice(deviceUid, tagName)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when adding tag to device", async () => {
      const deviceUid = "device-uid-123";
      const tagName = "production";

      mockTagsApi
        .onPost(addTagUrl(deviceUid, tagName))
        .networkError();

      await expect(tagsStore.addTagToDevice(deviceUid, tagName)).rejects.toThrow("Network Error");
    });
  });

  describe("removeTagFromDevice", () => {
    const removeTagUrl = (deviceUid: string, tagName: string) => `http://localhost:3000/api/devices/${deviceUid}/tags/${tagName}`;

    it("should remove tag from device successfully", async () => {
      const deviceUid = "device-uid-123";
      const tagName = "production";

      mockTagsApi
        .onDelete(removeTagUrl(deviceUid, tagName))
        .reply(200);

      await expect(tagsStore.removeTagFromDevice(deviceUid, tagName)).resolves.not.toThrow();
    });

    it("should handle not found error when removing tag from device", async () => {
      const deviceUid = "non-existent-device";
      const tagName = "production";

      mockTagsApi
        .onDelete(removeTagUrl(deviceUid, tagName))
        .reply(404, { message: "Device or tag not found" });

      await expect(tagsStore.removeTagFromDevice(deviceUid, tagName)).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when removing tag from device", async () => {
      const deviceUid = "device-uid-123";
      const tagName = "production";

      mockTagsApi
        .onDelete(removeTagUrl(deviceUid, tagName))
        .reply(500);

      await expect(tagsStore.removeTagFromDevice(deviceUid, tagName)).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when removing tag from device", async () => {
      const deviceUid = "device-uid-123";
      const tagName = "production";

      mockTagsApi
        .onDelete(removeTagUrl(deviceUid, tagName))
        .networkError();

      await expect(tagsStore.removeTagFromDevice(deviceUid, tagName)).rejects.toThrow("Network Error");
    });
  });
});
