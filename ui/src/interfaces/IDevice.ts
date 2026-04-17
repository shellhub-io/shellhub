import { UpdateDeviceRequest } from "@/api/client";
import { ITag } from "./ITags";

type Identity = {
  mac: string;
};

type InfoDetails = {
  id: string;
  pretty_name: string;
  version: string;
  arch: string;
  platform: string;
};

type Position = {
  latitude: number;
  longitude: number;
};

type SSHSettings = {
  allow_password: boolean;
  allow_public_key: boolean;
  allow_root: boolean;
  allow_empty_passwords: boolean;
  allow_tty: boolean;
  allow_tcp_forwarding: boolean;
  allow_web_endpoints: boolean;
  allow_sftp: boolean;
  allow_agent_forwarding: boolean;
};

export type DeviceStatus = "accepted" | "pending" | "rejected";

export interface IDevice {
  uid: string;
  name: string;
  identity: Identity;
  info: InfoDetails;
  public_key: string;
  tenant_id: string;
  last_seen: string;
  online: boolean;
  namespace: string;
  status: DeviceStatus;
  created_at: string;
  status_updated_at: string;
  remote_addr: string;
  position: Position;
  tags: Array<ITag>;
  settings?: SSHSettings;
}

export interface IDeviceRename {
  uid: string;
  name: UpdateDeviceRequest;
}

export interface FetchDevicesParams {
  perPage?: number;
  page?: number;
  filter?: string;
  status?: DeviceStatus;
  sortField?: string;
  sortOrder?: "asc" | "desc";
}

export interface IDeviceMethods {
  fetchDevices: (params: FetchDevicesParams) => Promise<void>;
  getList: () => IDevice[];
  getCount: () => number;
  getFilter: () => string | undefined;
}
