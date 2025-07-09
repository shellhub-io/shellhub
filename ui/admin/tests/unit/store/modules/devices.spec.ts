import { describe, expect, it, beforeEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import { ref } from "vue";
import { btoa } from "node:buffer";
import useDevicesStore from "@admin/store/modules/devices";

describe("Devices", () => {
  let devicesStore: ReturnType<typeof useDevicesStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    devicesStore = useDevicesStore();
  });

  const devices = [
    {
      created_at: "2020-05-20T19:58:53.276Z",
      identity: { mac: "00:00:00:00:00:00" },
      info: {
        arch: "x86_64",
        id: "linuxmint",
        platform: "linuxmint",
        pretty_name: "Linux Mint 19.3",
        version: "18.4.2",
      },
      last_seen: "2020-05-20T19:58:53.276Z",
      name: "tests",
      namespace: "dev",
      online: true,
      position: { latitude: 12, longitude: 12 },
      public_key: "xxxxxxxxxxxxxxxx",
      remote_addr: "127.0.0.1",
      status: "accepted",
      tags: ["xxxx", "yyyyy"],
      tenant_id: "00000000",
      uid: "a582b47a42d",
    },
    {
      created_at: "2022-05-20T19:58:53.276Z",
      identity: { mac: "00:00:00:00:00:00" },
      info: {
        arch: "x86_64",
        id: "linuxmint",
        platform: "linuxmint",
        pretty_name: "Linux Mint 19.3",
        version: "18.4.2",
      },
      last_seen: "2020-05-20T19:58:53.276Z",
      name: "ossytem",
      namespace: "ossytem",
      online: true,
      position: { latitude: 12, longitude: 12 },
      public_key: "xxxxxxxxxxxxxxxx",
      remote_addr: "127.0.0.1",
      status: "accepted",
      tags: ["xxxx", "yyyyy"],
      tenant_id: "00000000",
      uid: "a582b47a42d",
    },
  ];

  const numberDevices = 2;
  const filter = ref("");
  const filterToEncodeBase64 = {
    type: "property",
    params: { name: "name", operator: "contains", value: filter.value },
  };
  const encodedFilter = btoa(JSON.stringify(filterToEncodeBase64));
  const data = {
    page: 1,
    perPage: 10,
    filter: "",
    sortStatusField: "",
    sortStatusString: undefined,
  };

  it("returns device default variables", () => {
    expect(devicesStore.list).toEqual([]);
    expect(devicesStore.getDevice).toEqual({});
    expect(devicesStore.getNumberDevices).toBe(0);
    expect(devicesStore.getPage).toBe(1);
    expect(devicesStore.getPerPage).toBe(10);
    expect(devicesStore.getFilter).toBe("");
  });

  it("sets devices and number of devices", () => {
    devicesStore.devices = devices;
    devicesStore.numberDevices = numberDevices;

    expect(devicesStore.list).toEqual(devices);
    expect(devicesStore.getNumberDevices).toBe(numberDevices);
  });

  it("sets page, perPage and filter", () => {
    devicesStore.page = data.page;
    devicesStore.perPage = data.perPage;
    devicesStore.filter = data.filter;

    expect(devicesStore.getPage).toBe(1);
    expect(devicesStore.getPerPage).toBe(10);
    expect(devicesStore.getFilter).toBe("");
  });

  it("sets encoded filter", () => {
    devicesStore.setFilter(encodedFilter);
    expect(devicesStore.getFilter).toBe(encodedFilter);
  });

  it("clears the list of devices", () => {
    devicesStore.devices = devices;
    expect(devicesStore.list.length).toBeGreaterThan(0);

    devicesStore.devices = [];
    expect(devicesStore.list).toEqual([]);
  });
});
