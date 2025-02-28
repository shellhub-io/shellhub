// tests/store/modules/tags.spec.ts
import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { tagsApi } from "@/api/http";
import useTagsStore from "@/store/modules/tags";
import type { Tags } from "@/interfaces/ITags";

const TENANT = "fake-tenant";
const BASE = "http://localhost:3000";

// eslint-disable-next-line vue/max-len
const makeUrl = (tenant: string, filter: string, page: number, perPage: number) => `${BASE}/api/namespaces/${tenant}/tags?filter=${encodeURIComponent(filter)}&page=${page}&per_page=${perPage}`;

const mockTags: Tags[] = [{ name: "tag1" }, { name: "tag2" }, { name: "tag3" }];

describe("Tags Store", () => {
  let mock: MockAdapter;
  let store: ReturnType<typeof useTagsStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    store = useTagsStore();
    mock = new MockAdapter(tagsApi.getAxios());
  });

  afterEach(() => {
    mock.restore();
  });

  it("should have initial state values", () => {
    expect(store.tags).toEqual([]);
    expect(store.numberTags).toBe(0);
    expect(store.page).toBe(1);
    expect(store.perPage).toBe(10);
    expect(store.filter).toBe("");
    expect(store.getSelected("device")).toEqual([]);
    expect(store.getSelected("container")).toEqual([]);
    expect(store.list).toEqual([]);
    expect(store.getNumberTags).toBe(0);
    expect(store.getPage).toBe(1);
    expect(store.getPerPage).toBe(10);
    expect(store.getFilter).toBe("");
  });

  it("successfully fetches tags (fetch)", async () => {
    const url = makeUrl(TENANT, "", 1, 10);
    mock.onGet(url).reply(200, mockTags, { "x-total-count": String(mockTags.length) });

    await store.fetch({ tenant: TENANT, filter: "", page: 1, perPage: 10 });

    expect(store.list).toEqual(mockTags);
    expect(store.getNumberTags).toBe(3);
    expect(store.getPage).toBe(1);
    expect(store.getPerPage).toBe(10);
    expect(store.getFilter).toBe("");
  });

  it("search updates tags and filter while keeping page/perPage", async () => {
    const initialUrl = makeUrl(TENANT, "", 2, 20);
    mock.onGet(initialUrl).reply(200, mockTags, { "x-total-count": String(mockTags.length) });
    await store.fetch({ tenant: TENANT, filter: "", page: 2, perPage: 20 });

    const filter = "abc";
    const searchUrl = makeUrl(TENANT, filter, 2, 20);
    mock.onGet(searchUrl).reply(200, mockTags, { "x-total-count": "3" });

    await store.search({ tenant: TENANT, filter });

    expect(store.list).toEqual(mockTags);
    expect(store.getNumberTags).toBe(3);
    expect(store.getFilter).toBe(filter);
    // page/perPage unchanged
    expect(store.getPage).toBe(2);
    expect(store.getPerPage).toBe(20);
  });

  it("autocomplete loads using provided page/perPage but does not mutate page/perPage in state", async () => {
    const url1 = makeUrl(TENANT, "", 1, 10);
    mock.onGet(url1).reply(200, mockTags, { "x-total-count": "3" });
    await store.fetch({ tenant: TENANT, filter: "", page: 1, perPage: 10 });

    const autoUrl = makeUrl(TENANT, "", 3, 50);
    mock.onGet(autoUrl).reply(200, mockTags, { "x-total-count": "3" });

    await store.autocomplete({ tenant: TENANT, filter: "", page: 3, perPage: 50 });

    expect(store.list).toEqual(mockTags);
    expect(store.getNumberTags).toBe(3);
    expect(store.getPage).toBe(1);
    expect(store.getPerPage).toBe(10);
  });

  it("fetch clears list on error", async () => {
    const badUrl = makeUrl(TENANT, "", 1, 10);
    mock.onGet(badUrl).reply(500);

    await expect(
      store.fetch({ tenant: TENANT, filter: "", page: 1, perPage: 10 }),
    ).rejects.toBeTruthy();

    expect(store.list).toEqual([]);
    expect(store.getNumberTags).toBe(0);
  });

  it("setSelected toggles and clearSelected resets per variant", () => {
    const t1: Tags = { name: "tag1" };
    const t2: Tags = { name: "tag2" };

    store.setSelected({ variant: "device", tag: t1 });
    store.setSelected({ variant: "device", tag: t2 });
    expect(store.getSelected("device")).toEqual([t1, t2]);

    store.clearSelected("device");
    store.setSelected({ variant: "device", tag: t1 });
    expect(store.getSelected("device")).toEqual([t1]);

    store.clearSelected("device");
    expect(store.getSelected("device")).toEqual([]);
    expect(store.getSelected("container")).toEqual([]);
  });
});
