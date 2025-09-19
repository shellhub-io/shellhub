import { describe, expect, it } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useDevicesStore from "@admin/store/modules/devices";
import { IDevice } from "@/interfaces/IDevice";

describe("Devices", () => {
  setActivePinia(createPinia());
  const devicesStore = useDevicesStore();

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
      name: "ossystems",
      namespace: "ossystems",
      online: true,
      position: { latitude: 12, longitude: 12 },
      public_key: "xxxxxxxxxxxxxxxx",
      remote_addr: "127.0.0.1",
      status: "accepted",
      tags: [{ name: "xxxx" }, { name: "yyyyy" }],
      tenant_id: "00000000",
      uid: "a582b47a42d",
    },
  ];

  const deviceCount = 2;

  it("returns devices store default variables", () => {
    expect(devicesStore.devices).toEqual([]);
    expect(devicesStore.deviceCount).toBe(0);
  });

  it("sets devices and number of devices", () => {
    devicesStore.devices = devices as IDevice[];
    devicesStore.deviceCount = deviceCount;

    expect(devicesStore.devices).toEqual(devices);
    expect(devicesStore.deviceCount).toBe(deviceCount);
  });
});
