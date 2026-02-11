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
    platform: "linux" as const,
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

/**
 * Mock device with online status and multiple tags.
 * Provides an Ubuntu device with production tags.
 */
export const mockDeviceOnlineWithTags: IDevice = {
  uid: "device-1",
  name: "device-one",
  namespace: "user-ns",
  identity: { mac: "00:00:00:00:00:01" },
  info: {
    id: "ubuntu",
    pretty_name: "Ubuntu 22.04",
    version: "22.04",
    arch: "x86_64",
    platform: "linux",
  },
  public_key: "---- PUBLIC KEY ----",
  tenant_id: "tenant-1",
  last_seen: "2025-01-15T10:00:00.000Z",
  online: true,
  status: "accepted",
  status_updated_at: "2025-01-15T10:00:00.000Z",
  tags: [
    { name: "production", tenant_id: "tenant-1", created_at: "2025-01-01T00:00:00.000Z", updated_at: "2025-01-01T00:00:00.000Z" },
    { name: "web-server", tenant_id: "tenant-1", created_at: "2025-01-01T00:00:00.000Z", updated_at: "2025-01-01T00:00:00.000Z" },
  ],
  created_at: "2025-01-01T00:00:00.000Z",
  position: { latitude: 0, longitude: 0 },
  remote_addr: "192.168.1.1",
};

/**
 * Mock device with offline status and no tags.
 * Provides an Alpine Linux device in offline state.
 */
export const mockDeviceOfflineNoTags: IDevice = {
  uid: "device-2",
  name: "device-two",
  namespace: "user-ns",
  identity: { mac: "00:00:00:00:00:02" },
  info: {
    id: "alpine",
    pretty_name: "Alpine Linux",
    version: "3.18",
    arch: "x86_64",
    platform: "linux",
  },
  public_key: "---- PUBLIC KEY ----",
  tenant_id: "tenant-1",
  last_seen: "2025-01-14T10:00:00.000Z",
  online: false,
  status: "accepted",
  status_updated_at: "2025-01-14T10:00:00.000Z",
  tags: [],
  created_at: "2025-01-01T00:00:00.000Z",
  position: { latitude: 0, longitude: 0 },
  remote_addr: "192.168.1.2",
};

/**
 * Mock device with pending status.
 * Provides a Debian device awaiting acceptance.
 */
export const mockDevicePending: IDevice = {
  uid: "device-3",
  name: "device-three",
  namespace: "user-ns",
  identity: { mac: "00:00:00:00:00:03" },
  info: {
    id: "debian",
    pretty_name: "Debian 12",
    version: "12",
    arch: "x86_64",
    platform: "linux" as const,
  },
  public_key: "---- PUBLIC KEY ----",
  tenant_id: "tenant-1",
  last_seen: "2025-01-16T10:00:00.000Z",
  online: true,
  status: "pending" as const,
  status_updated_at: "2025-01-16T10:00:00.000Z",
  tags: [],
  created_at: "2025-01-16T10:00:00.000Z",
  position: { latitude: 0, longitude: 0 },
  remote_addr: "192.168.1.3",
};
