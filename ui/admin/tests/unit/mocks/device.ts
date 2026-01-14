import { IAdminDevice } from "@admin/interfaces/IDevice";

export const mockDevice: IAdminDevice = {
  uid: "device-123",
  name: "test-device",
  identity: {
    mac: "00:11:22:33:44:55",
  },
  info: {
    id: "device-123",
    pretty_name: "Ubuntu 22.04",
    version: "v1.0.0",
    arch: "x86_64",
    platform: "linux",
  },
  public_key: "ssh-rsa AAAA...",
  tenant_id: "tenant-123",
  last_seen: "2024-01-10T12:00:00Z",
  online: true,
  namespace: "default",
  status: "accepted",
  status_updated_at: "2024-01-01T00:00:00Z",
  created_at: "2024-01-01T00:00:00Z",
  remote_addr: "192.168.1.100",
  position: {
    latitude: 0,
    longitude: 0,
  },
  tags: [],
};

export const mockDevices: IAdminDevice[] = [
  { ...mockDevice, uid: "device-1", name: "device-one" },
  { ...mockDevice, uid: "device-2", name: "device-two", online: false },
  { ...mockDevice, uid: "device-3", name: "device-three" },
];
