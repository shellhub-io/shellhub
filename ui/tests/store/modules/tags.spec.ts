import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import { describe, expect, it } from "vitest";
import { tagsApi } from "@/api/http";
import useTagsStore from "@/store/modules/tags";
import type { ITag } from "@/interfaces/ITags";

const BASE = "http://localhost:3000";

const makeUrl = (page: number, perPage: number, filter?: string) => {
  const params = new URLSearchParams();
  if (filter) params.append("filter", filter);
  params.append("page", String(page));
  params.append("per_page", String(perPage));
  return `${BASE}/api/tags?${params.toString()}`;
};

const mockTags = [
  { name: "tag1", tenant_id: "tenant1", created_at: "2025-01-01", updated_at: "2025-01-01" },
  { name: "tag2", tenant_id: "tenant1", created_at: "2025-01-01", updated_at: "2025-01-01" },
  { name: "tag3", tenant_id: "tenant1", created_at: "2025-01-01", updated_at: "2025-01-01" },
] as ITag[];

describe("Tags Store", () => {
  setActivePinia(createPinia());
  const tagsStore = useTagsStore();
  const mockTagsApi = new MockAdapter(tagsApi.getAxios());

  // afterEach(() => { mockTagsApi.restore(); });

  it("should have initial state values", () => {
    expect(tagsStore.tags).toEqual([]);
    expect(tagsStore.tagCount).toBe(0);
    expect(tagsStore.showTags).toBe(false);
    expect(tagsStore.selectedTags).toEqual([]);
  });

  it("successfully fetches tags", async () => {
    const url = makeUrl(1, 10);
    mockTagsApi.onGet(url).reply(200, mockTags, { "x-total-count": String(mockTags.length) });

    await tagsStore.fetchTagList({ page: 1, perPage: 10 });

    expect(tagsStore.tags).toEqual(mockTags);
    expect(tagsStore.tagCount).toBe(3);
  });

  it("fetchTagList with filter parameter", async () => {
    const filter = "abc";
    const url = makeUrl(1, 10, filter);
    mockTagsApi.onGet(url).reply(200, mockTags, { "x-total-count": "3" });

    await tagsStore.fetchTagList({ filter, page: 1, perPage: 10 });

    expect(tagsStore.tags).toEqual(mockTags);
    expect(tagsStore.tagCount).toBe(3);
  });

  it("fetchTagList clears list on error", async () => {
    const url = makeUrl(1, 10);
    mockTagsApi.onGet(url).reply(500);

    await expect(
      tagsStore.fetchTagList({ page: 1, perPage: 10 }),
    ).rejects.toBeTruthy();

    expect(tagsStore.tags).toEqual([]);
    expect(tagsStore.tagCount).toBe(0);
  });

  it("setTagListVisibility sets showTags when tags exist", async () => {
    const url = makeUrl(1, 1);
    mockTagsApi.onGet(url).reply(200, [mockTags[0]], { "x-total-count": "3" });

    await tagsStore.setTagListVisibility();

    expect(tagsStore.showTags).toBe(true);
  });

  it("toggleSelectedTag adds and removes tags", () => {
    const t1 = { name: "tag1" } as ITag;
    const t2 = { name: "tag2" } as ITag;

    tagsStore.toggleSelectedTag(t1);
    tagsStore.toggleSelectedTag(t2);
    expect(tagsStore.selectedTags).toEqual([t1, t2]);

    tagsStore.toggleSelectedTag(t1);
    expect(tagsStore.selectedTags).toEqual([t2]);

    tagsStore.toggleSelectedTag(t2);
    expect(tagsStore.selectedTags).toEqual([]);
  });

  it("can clear selectedTags directly", () => {
    const t1 = { name: "tag1" } as ITag;
    tagsStore.toggleSelectedTag(t1);
    expect(tagsStore.selectedTags).toEqual([t1]);

    tagsStore.selectedTags = [];
    expect(tagsStore.selectedTags).toEqual([]);
  });

  it("successfully creates a tag", async () => {
    mockTagsApi.onPost(`${BASE}/api/tags`).reply(201);

    await tagsStore.createTag("new-tag");

    expect(mockTagsApi.history.post.length).toBe(1);
  });

  it("successfully updates a tag", async () => {
    mockTagsApi.onPatch(`${BASE}/api/tags/old-tag`).reply(200);

    await tagsStore.updateTag("old-tag", { name: "new-tag" });

    expect(mockTagsApi.history.patch.length).toBe(1);
  });

  it("successfully deletes a tag", async () => {
    mockTagsApi.onDelete(`${BASE}/api/tags/tag-to-delete`).reply(200);

    await tagsStore.deleteTag("tag-to-delete");

    expect(mockTagsApi.history.delete.length).toBe(1);
  });

  it("successfully adds tag to device", async () => {
    mockTagsApi.resetHistory();
    mockTagsApi.onPost(`${BASE}/api/devices/device-123/tags/tag1`).reply(200);

    await tagsStore.addTagToDevice("device-123", "tag1");

    expect(mockTagsApi.history.post.length).toBe(1);
  });

  it("successfully removes tag from device", async () => {
    mockTagsApi.resetHistory();
    mockTagsApi.onDelete(`${BASE}/api/devices/device-123/tags/tag1`).reply(200);

    await tagsStore.removeTagFromDevice("device-123", "tag1");

    expect(mockTagsApi.history.delete.length).toBe(1);
  });
});
