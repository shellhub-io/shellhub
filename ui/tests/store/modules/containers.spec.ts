import MockAdapter from "axios-mock-adapter";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { containersApi } from "@/api/http";
import { store } from "@/store";

const initialContainers = {
  data: [
    { uid: "a582b47a42d", name: "Device 1" },
    { uid: "a582b47a42e", name: "Device 2" },
  ],
  headers: {
    "x-total-count": 2,
  },
};

describe("Containers store", () => {
  let mockContainers: MockAdapter;

  beforeEach(() => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    mockContainers = new MockAdapter(containersApi.getAxios());
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it("Returns containers default variables", () => {
    const defaultState = {
      containers: [],
      container: {},
      numberContainers: 0,
      showContainers: false,
      page: 1,
      perPage: 10,
      filter: "",
      status: "accepted",
      sortStatusField: undefined,
      sortStatusString: "asc",
    };

    expect(store.getters["container/list"]).toEqual(defaultState.containers);
    expect(store.getters["container/get"]).toEqual(defaultState.container);
    expect(store.getters["container/getShowContainers"]).toEqual(defaultState.showContainers);
    expect(store.getters["container/getNumberContainers"]).toEqual(defaultState.numberContainers);
    expect(store.getters["container/getPage"]).toEqual(defaultState.page);
    expect(store.getters["container/getPerPage"]).toEqual(defaultState.perPage);
    expect(store.getters["container/getFilter"]).toEqual(defaultState.filter);
    expect(store.getters["container/getStatus"]).toEqual(defaultState.status);
    expect(store.getters["container/getSortStatusField"]).toEqual(defaultState.sortStatusField);
    expect(store.getters["container/getSortStatusString"]).toEqual(defaultState.sortStatusString);
  });

  it("Fetches containers and updates state accordingly", async () => {
    const devices = [{ uid: "1", name: "Device 1" }, { uid: "2", name: "Device 2" }];
    const totalCount = 2;

    mockContainers.onGet("http://localhost:3000/api/containers?filter=&page=1&per_page=10&status=accepted")
      .reply(200, devices, { "x-total-count": totalCount });

    await store.dispatch("container/fetch", {
      page: 1,
      perPage: 10,
      filter: "",
      status: "accepted",
      sortStatusField: undefined,
      sortStatusString: "asc",
    });

    expect(store.getters["container/list"]).toEqual(devices);
    expect(store.getters["container/getNumberContainers"]).toEqual(totalCount);
  });

  it("Removes a container from the state", async () => {
    // Mock the API call
    mockContainers.onDelete("http://localhost:3000/api/containers/a582b47a42d").reply(200);

    const storeSpy = vi.spyOn(store, "dispatch");
    // Call the action
    await store.dispatch("container/remove", "a582b47a42d");

    expect(storeSpy).toBeCalledWith("container/remove", "a582b47a42d");
  });

  it("Renames a device in the state", async () => {
    const containerToUpdate = { uid: "a582b47a42d", name: "Device 1" };
    const newName = "Updated Device 1";
    const updatedContainer = { ...containerToUpdate, name: newName };
    // Set initial state
    store.commit("devices/setDevices", initialContainers);
    const storeSpy = vi.spyOn(store, "dispatch");

    // Mock the API call
    mockContainers.onPut(`http://localhost:3000/api/containers/${containerToUpdate.uid}`).reply(200);

    // Call the action
    await store.dispatch("container/rename", updatedContainer);

    // Assert the device was renamed in the state
    expect(storeSpy).toBeCalledWith("container/rename", updatedContainer);
  });

  it("Gets a device by its UID and updates state", async () => {
    const uid = "a582b47a42d";
    const container = { uid, name: "Device 1" };

    // Mock the API call
    mockContainers.onGet(`http://localhost:3000/api/containers/${uid}`).reply(200, container);

    await store.dispatch("container/get", uid);

    expect(store.getters["container/get"]).toEqual(container);
  });

  it("Accepts a device and updates state", async () => {
    const uid = "a582b47a42d";
    const storeSpy = vi.spyOn(store, "dispatch");

    // Mock the API call
    mockContainers.onPatch(`http://localhost:3000/api/containers/${uid}/accept`).reply(200);

    await store.dispatch("container/accept", uid);

    expect(storeSpy).toBeCalledWith("container/accept", uid);
  });

  it("Rejects a device and updates state", async () => {
    const uid = "a582b47a42d";
    const storeSpy = vi.spyOn(store, "dispatch");

    // Mock the API call
    mockContainers.onPatch(`http://localhost:3000/api/containers/${uid}/reject`).reply(200);

    await store.dispatch("container/reject", uid);

    expect(storeSpy).toBeCalledWith("container/reject", uid);
  });

  it("Sets filter and updates state", async () => {
    const filter = "some_filter";

    await store.dispatch("container/setFilter", filter);

    expect(store.getters["container/getFilter"]).toEqual(filter);
  });

  it("Searches for devices and updates state", async () => {
    const devices = [{ uid: "1", name: "Device 1" }, { uid: "2", name: "Device 2" }];
    const totalCount = 2;
    const data = {
      page: 1,
      perPage: 10,
      filter: "some_filter",
    };

    // eslint-disable-next-line vue/max-len
    mockContainers.onGet(`http://localhost:3000/api/containers?filter=${data.filter}&page=${data.page}&per_page=${data.perPage}&status=accepted`)
      .reply(200, devices, { "x-total-count": totalCount });

    await store.dispatch("container/search", data);

    expect(store.getters["container/list"]).toEqual(devices);
    expect(store.getters["container/getNumberContainers"]).toEqual(totalCount);
    expect(store.getters["container/getFilter"]).toEqual(data.filter);
  });
});
