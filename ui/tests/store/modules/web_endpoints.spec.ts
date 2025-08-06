import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { webEndpointsApi } from "@/api/http";
import { store } from "@/store";

const initialWebEndpoints = {
  data: [
    { uid: "abc123", address: "localhost", port: 8080, expires_in: 60000 },
    { uid: "def456", address: "127.0.0.1", port: 8081, expires_in: 60000 },
  ],
  headers: {
    "x-total-count": 2,
  },
};

describe("WebEndpoints store", () => {
  let mock: MockAdapter;

  beforeEach(() => {
    vi.useFakeTimers();
    mock = new MockAdapter(webEndpointsApi.getAxios());
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    mock.restore();
  });

  it("Returns web endpoints default variables", () => {
    expect(store.getters["webEndpoints/listWebEndpoints"]).toEqual([]);
    expect(store.getters["webEndpoints/getFilter"]).toBe("");
    expect(store.getters["webEndpoints/getPage"]).toBe(1);
    expect(store.getters["webEndpoints/getPerPage"]).toBe(10);
    expect(store.getters["webEndpoints/getTotalCount"]).toBe(0);
    expect(store.getters["webEndpoints/getSortBy"]).toBe("uid");
    expect(store.getters["webEndpoints/getOrderBy"]).toBe("asc");
    expect(store.getters["webEndpoints/getShowWebEndpoints"]).toBe(false);
  });

  it("Fetches web endpoints via get() and updates state", async () => {
    mock
      .onGet("http://localhost:3000/api/web-endpoints?filter=&page=1&per_page=10&sort_by=uid&order_by=asc")
      .reply(200, initialWebEndpoints.data, initialWebEndpoints.headers);

    await store.dispatch("webEndpoints/get", {
      page: 1,
      perPage: 10,
      filter: "",
      sortBy: "uid",
      orderBy: "asc",
    });

    expect(store.getters["webEndpoints/listWebEndpoints"]).toEqual(initialWebEndpoints.data);
    expect(store.getters["webEndpoints/getTotalCount"]).toBe(2);
    expect(store.getters["webEndpoints/getShowWebEndpoints"]).toBe(true);
  });

  it("Searches web endpoints and sets filter", async () => {
    mock
      .onGet("http://localhost:3000/api/web-endpoints?filter=test&page=1&per_page=10&sort_by=uid&order_by=asc")
      .reply(200, initialWebEndpoints.data, initialWebEndpoints.headers);

    await store.dispatch("webEndpoints/search", {
      page: 1,
      perPage: 10,
      filter: "test",
    });

    expect(store.getters["webEndpoints/listWebEndpoints"]).toEqual(initialWebEndpoints.data);
    expect(store.getters["webEndpoints/getFilter"]).toBe("test");
  });

  it("Creates a new web endpoint", async () => {
    const payload = { uid: "abc123", host: "localhost", port: 8080, ttl: -1 };
    mock.onPost("http://localhost:3000/api/web-endpoints").reply(200);

    const spy = vi.spyOn(store, "dispatch");
    await store.dispatch("webEndpoints/create", payload);

    expect(spy).toHaveBeenCalledWith("webEndpoints/create", payload);
  });

  it("Deletes a web endpoint", async () => {
    const address = "localhost";
    mock.onDelete(`http://localhost:3000/api/web-endpoints/${address}`).reply(200);

    const spy = vi.spyOn(store, "dispatch");
    await store.dispatch("webEndpoints/delete", { address });

    expect(spy).toHaveBeenCalledWith("webEndpoints/delete", { address });
  });

  it("Updates sort state", async () => {
    const sortPayload = { sortBy: "address", orderBy: "desc" };
    store.commit("webEndpoints/setSortStatus", sortPayload);

    expect(store.getters["webEndpoints/getSortBy"]).toBe("address");
    expect(store.getters["webEndpoints/getOrderBy"]).toBe("desc");
  });
});
