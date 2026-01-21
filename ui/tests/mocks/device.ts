import { IDevice } from "@/interfaces/IDevice";

/**
 * Mock device data for testing.
 * Provides a complete device object with all required fields.
 */
export const mockDevice: IDevice = {
  uid: "a582b47a42d",
  name: "39-5e-2a",
  identity: {
    mac: "00:00:00:00:00:00",
  },
  info: {
    id: "linuxmint",
    pretty_name: "Linux Mint 19.3",
    version: "",
    arch: "x86_64",
    platform: "linux",
  },
  public_key: "----- PUBLIC KEY -----",
  tenant_id: "fake-tenant-data",
  last_seen: "2020-05-20T18:58:53.276Z",
  created_at: "2020-05-20T18:00:00.000Z",
  status_updated_at: "2020-05-20T18:58:53.276Z",
  online: false,
  namespace: "user",
  status: "accepted",
  remote_addr: "127.0.0.1",
  position: { latitude: 0, longitude: 0 },
  tags: [
    {
      tenant_id: "fake-tenant-data",
      name: "test-tag",
      created_at: "",
      updated_at: "",
    },
  ],
};

/**
 * Mock device for session details.
 * Provides a device with Manjaro Linux configuration.
 */
export const mockDeviceForSession: IDevice = {
  uid: "1",
  name: "00-00-00-00-00-01",
  identity: {
    mac: "00-00-00-00-00-01",
  },
  info: {
    id: "manjaro",
    pretty_name: "Manjaro Linux",
    version: "latest",
    arch: "amd64",
    platform: "docker",
  },
  public_key: "",
  tenant_id: "fake-tenant-data",
  last_seen: "2025-01-02T00:00:00.000Z",
  online: true,
  namespace: "dev",
  status: "accepted",
  status_updated_at: "0",
  created_at: "2025-01-01T00:00:00.000Z",
  remote_addr: "192.168.0.1",
  position: { latitude: 0, longitude: 0 },
  tags: [],
};
