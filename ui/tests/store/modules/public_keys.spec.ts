import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { store } from "@/store";
import { sshApi } from "@/api/http";

const publicKeyObject = {
  data: "test-key",
  filter: {
    hostname: ".*",
  },
  name: "example",
  username: ".*",
};

const publicKeyList = [
  {
    data: "test-key",
    fingerprint: "fake-fingerprint",
    created_at: "2020-05-01T00:00:00.000Z",
    tenant_id: "fake-tenant",
    name: "example",
    filter:
    {
      hostname: ".*",
    },
    username: ".*",
  },
];

describe("Public keys store", () => {
  let mockSsh: MockAdapter;

  beforeEach(() => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant");
    mockSsh = new MockAdapter(sshApi.getAxios());
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it("Return publicKeys default variables", () => {
    expect(store.getters["publicKeys/list"]).toEqual([]);
    expect(store.getters["publicKeys/get"]).toEqual({});
    expect(store.getters["publicKeys/getNumberPublicKeys"]).toEqual(0);
    expect(store.getters["publicKeys/getPage"]).toEqual(1);
    expect(store.getters["publicKeys/getPerPage"]).toEqual(10);
  });

  it("Test Create Public Key action", async () => {
    const reqSpy = vi.spyOn(store, "dispatch");

    // Mock the API call for creating public key
    mockSsh.onPost("http://localhost:3000/api/sshkeys/public-keys").reply(200);

    // Trigger the create public key action
    await store.dispatch("publicKeys/post", publicKeyObject);

    // Check if the state has been updated correctly
    expect(reqSpy).toHaveBeenCalled();
  });

  it("Test Get Public Key action", async () => {
    const reqSpy = vi.spyOn(store, "dispatch");

    // Mock the API call for getting public key
    mockSsh.onGet("http://localhost:3000/api/sshkeys/public-keys?filter=&page=1&per_page=10").reply(200, publicKeyList);

    // Trigger the create public key action
    await store.dispatch("publicKeys/fetch", { page: 1, perPage: 10, filter: "" });

    // Check if the state has been updated correctly
    expect(reqSpy).toHaveBeenCalled();
    expect(store.getters["publicKeys/list"]).toEqual(publicKeyList);
  });
});
