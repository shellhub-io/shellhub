import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import { describe, expect, it } from "vitest";
import { tagsApi } from "@/api/http";
import useTagsStore from "@/store/modules/tags";

const mockTags = ["tag1", "tag2", "tag3"];

describe("Tags Store", () => {
  const mockTagsApi = new MockAdapter(tagsApi.getAxios());
  setActivePinia(createPinia());
  const tagsStore = useTagsStore();

  it("should have initial state values", () => {
    expect(tagsStore.tags).toEqual([]);
    expect(tagsStore.tagsCount).toEqual(0);
  });

  it("successfully fetches tags", async () => {
    mockTagsApi.onGet("http://localhost:3000/api/tags").reply(200, mockTags, {
      "x-total-count": "3",
    });

    await tagsStore.fetchTags();

    expect(tagsStore.tags).toEqual(mockTags);
    expect(tagsStore.tagsCount).toEqual(3);
  });
});
