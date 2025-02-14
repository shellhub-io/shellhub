import { describe, expect, it } from "vitest";
import { ref } from "vue";
import { btoa } from "node:buffer";
import { store } from "../../../../src/store";

describe("Devices", () => {
  const devices = [
    {
      created_at: "2020-05-20T19:58:53.276Z",
      identity: { mac: "00:00:00:00:00:00" },
      info: {
        arch: "x86_64",
        id: "linuxmint",
        platform: "linuxmint",
        prettyName: "Linux Mint 19.3",
        version: "18.4.2",
      },
      last_seen: "2020-05-20T19:58:53.276Z",
      name: "tests",
      namespace: "dev",
      online: true,
      position: {
        latitude: 12,
        longitude: 12,
      },
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
        prettyName: "Linux Mint 19.3",
        version: "18.4.2",
      },
      last_seen: "2020-05-20T19:58:53.276Z",
      name: "ossytem",
      namespace: "ossytem",
      online: true,
      position: {
        latitude: 12,
        longitude: 12,
      },
      public_key: "xxxxxxxxxxxxxxxx",
      remote_addr: "127.0.0.1",
      status: "accepted",
      tags: ["xxxx", "yyyyy"],
      tenant_id: "00000000",
      uid: "a582b47a42d",
    },
  ];

  const numberDevices = 2;

  // filter
  const filter = ref("");
  let encodedFilter = "";

  const filterToEncodeBase64 = {
    type: "property",
    params: { name: "name", operator: "contains", value: filter.value },
  };
  encodedFilter = btoa(JSON.stringify(filterToEncodeBase64));
  const data = {
    page: 1,
    perPage: 10,
    filter: null,
    status: "accepted",
    uid: "a582b47a42f",
  };

  it("Return device default variables", () => {
    expect(store.getters["devices/list"]).toEqual([]);
    expect(store.getters["devices/get"]).toEqual({});
    expect(store.getters["devices/numberDevices"]).toEqual(0);
    expect(store.getters["devices/page"]).toEqual(1);
    expect(store.getters["devices/perPage"]).toEqual(10);
    expect(store.getters["devices/filter"]).toEqual("");
  });
  it("Verify initial states change for mutation setDevices", () => {
    store.commit("devices/setDevices", {
      data: devices,
      headers: { "x-total-count": numberDevices },
    });

    expect(store.getters["devices/list"]).toEqual(devices);
    expect(store.getters["devices/numberDevices"]).toEqual(numberDevices);
  });
  it("Verify initial states change for mutation serPagePerpageFilter", () => {
    store.commit("devices/setPagePerpageFilter", data);

    expect(store.getters["devices/page"]).toEqual(1);
    expect(store.getters["devices/perPage"]).toEqual(10);
    expect(store.getters["devices/filter"]).toEqual(null);
  });

  it("Verify changed filter state in setFilter mutation", () => {
    store.commit("devices/setFilterDevices", encodedFilter);
    expect(store.getters["devices/filter"]).toEqual(
      encodedFilter,
    );
  });

  it("Verify empty devices state for clearListDevices mutation", () => {
    store.commit("devices/clearListDevices");
    expect(store.getters["devices/list"]).toEqual([]);
  });
});
