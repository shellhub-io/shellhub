import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { nextTick } from "vue";
import { store } from "@/store";
import { devicesApi } from "@/api/http";

const initialDevices = {
  data: [
    { uid: "a582b47a42d", name: "Device 1" },
    { uid: "a582b47a42e", name: "Device 2" },
  ],
  headers: {
    "x-total-count": 2,
  },
};
describe("Devices store", () => {
  let mockDevices: MockAdapter;

  beforeEach(() => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    mockDevices = new MockAdapter(devicesApi.getAxios());
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it("Returns devices default variables", () => {
    const defaultState = {
      devices: [],
      quickConnectionList: [],
      device: {},
      numberDevices: 0,
      page: 1,
      perPage: 10,
      filter: "",
      status: "accepted",
      sortStatusField: undefined,
      sortStatusString: "asc",
      deviceChooserStatus: false,
      devicesForUserToChoose: [],
      numberdevicesForUserToChoose: 0,
      devicesSelected: [],
      deviceName: "",
    };

    expect(store.getters["devices/list"]).toEqual(defaultState.devices);
    expect(store.getters["devices/listQuickConnection"]).toEqual(defaultState.quickConnectionList);
    expect(store.getters["devices/get"]).toEqual(defaultState.device);
    expect(store.getters["devices/getNumberDevices"]).toEqual(defaultState.numberDevices);
    expect(store.getters["devices/getPage"]).toEqual(defaultState.page);
    expect(store.getters["devices/getPerPage"]).toEqual(defaultState.perPage);
    expect(store.getters["devices/getFilter"]).toEqual(defaultState.filter);
    expect(store.getters["devices/getStatus"]).toEqual(defaultState.status);
    expect(store.getters["devices/getSortStatusField"]).toEqual(defaultState.sortStatusField);
    expect(store.getters["devices/getSortStatusString"]).toEqual(defaultState.sortStatusString);
    expect(store.getters["devices/getDeviceChooserStatus"]).toEqual(defaultState.deviceChooserStatus);
    expect(store.getters["devices/getDevicesForUserToChoose"]).toEqual(defaultState.devicesForUserToChoose);
    expect(store.getters["devices/getNumberForUserToChoose"]).toEqual(defaultState.numberdevicesForUserToChoose);
    expect(store.getters["devices/getDevicesSelected"]).toEqual(defaultState.devicesSelected);
    expect(store.getters["devices/getDeviceToBeRenamed"]).toEqual(defaultState.deviceName);
  });

  it("Fetches devices and updates state accordingly", async () => {
    const devices = [{ uid: "1", name: "Device 1" }, { uid: "2", name: "Device 2" }];
    const totalCount = 2;

    mockDevices.onGet("http://localhost:3000/api/devices?filter=&page=1&per_page=10&status=accepted")
      .reply(200, devices, { "x-total-count": totalCount });

    await store.dispatch("devices/fetch", {
      page: 1,
      perPage: 10,
      filter: "",
      status: "accepted",
      sortStatusField: undefined,
      sortStatusString: "asc",
    });

    expect(store.getters["devices/list"]).toEqual(devices);
    expect(store.getters["devices/getNumberDevices"]).toEqual(totalCount);
  });

  it("Removes a device from the state", async () => {
    // Mock the API call
    mockDevices.onDelete("http://localhost:3000/api/devices/a582b47a42d").reply(200);

    const storeSpy = vi.spyOn(store, "dispatch");
    // Call the action
    await store.dispatch("devices/remove", "a582b47a42d");

    expect(storeSpy).toBeCalledWith("devices/remove", "a582b47a42d");
  });

  it("Renames a device in the state", async () => {
    const deviceToUpdate = { uid: "a582b47a42d", name: "Device 1" };
    const newName = "Updated Device 1";
    const updatedDevice = { ...deviceToUpdate, name: newName };
    // Set initial state
    store.commit("devices/setDevices", initialDevices);
    const storeSpy = vi.spyOn(store, "dispatch");

    // Mock the API call
    mockDevices.onPut(`http://localhost:3000/api/devices/${deviceToUpdate.uid}`).reply(200);

    // Call the action
    await store.dispatch("devices/rename", updatedDevice);

    // Assert the device was renamed in the state
    expect(storeSpy).toBeCalledWith("devices/rename", updatedDevice);
  });

  it("Gets a device by its UID and updates state", async () => {
    const uid = "a582b47a42d";
    const device = { uid, name: "Device 1" };

    // Mock the API call
    mockDevices.onGet(`http://localhost:3000/api/devices/${uid}`).reply(200, device);

    await store.dispatch("devices/get", uid);

    expect(store.getters["devices/get"]).toEqual(device);
  });

  it("Accepts a device and updates state", async () => {
    const uid = "a582b47a42d";
    const storeSpy = vi.spyOn(store, "dispatch");

    // Mock the API call
    mockDevices.onPatch(`http://localhost:3000/api/devices/${uid}/accept`).reply(200);

    await store.dispatch("devices/accept", uid);

    expect(storeSpy).toBeCalledWith("devices/accept", uid);
  });

  it("Rejects a device and updates state", async () => {
    const uid = "a582b47a42d";
    const storeSpy = vi.spyOn(store, "dispatch");

    // Mock the API call
    mockDevices.onPatch(`http://localhost:3000/api/devices/${uid}/reject`).reply(200);

    await store.dispatch("devices/reject", uid);

    expect(storeSpy).toBeCalledWith("devices/reject", uid);
  });

  it("Sets filter and updates state", async () => {
    const filter = "some_filter";

    await store.dispatch("devices/setFilter", filter);

    expect(store.getters["devices/getFilter"]).toEqual(filter);
  });

  it("Searches for devices and updates state", async () => {
    const devices = [{ uid: "1", name: "Device 1" }, { uid: "2", name: "Device 2" }];
    const totalCount = 2;
    const data = {
      page: 1,
      perPage: 10,
      filter: "some_filter",
    };
    mockDevices.onGet(`http://localhost:3000/api/devices?filter=${data.filter}&page=${data.page}&per_page=${data.perPage}&status=accepted`)
      .reply(200, devices, { "x-total-count": totalCount });

    await store.dispatch("devices/search", data);

    expect(store.getters["devices/list"]).toEqual(devices);
    expect(store.getters["devices/getNumberDevices"]).toEqual(totalCount);
    expect(store.getters["devices/getFilter"]).toEqual(data.filter);
  });

  it("Sets selected devices and updates state", async () => {
    const selectedDevices = [{ uid: "1", name: "Device 1" }, { uid: "2", name: "Device 2" }];

    await store.dispatch("devices/setDevicesSelected", selectedDevices);

    expect(store.getters["devices/getDevicesSelected"]).toEqual(selectedDevices);
  });

  it("Sets sort status and updates state", async () => {
    const sortStatus = {
      sortStatusString: "desc",
    };

    await store.dispatch("devices/setSortStatus", sortStatus);
    await nextTick();
    expect(store.getters["devices/getSortStatusString"]).toEqual(sortStatus.sortStatusString);
  });

  it("Updates device tag", async () => {
    const deviceUid = "a582b47a42d";
    const tags = ["tag1", "tag2"];
    const updateSpy = vi.spyOn(store, "dispatch");
    mockDevices.onPut("http://localhost:3000/api/devices/a582b47a42d/tags").reply(200);

    await store.dispatch("devices/updateDeviceTag", {
      uid: deviceUid,
      tags: { tags },
    });

    expect(updateSpy).toBeCalledWith("devices/updateDeviceTag", {
      uid: deviceUid,
      tags: { tags },
    });
  });
});
