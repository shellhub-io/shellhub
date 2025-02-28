import { UpdateDeviceRequest } from "@/api/client";
import { Tags } from "./ITags";

type Identity = {
  mac: string;
}

type InfoDetails = {
  id: string;
  pretty_name: string;
  version: string;
  arch: string;
  platform: string;
}

type Position = {
  latitude: number;
  longitude: number;
}

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
  status: string;
  created_at: string;
  remoteAddr: string;
  position: Position;
  tags: Array<Tags>;
}

export interface IDeviceRename {
  uid: string;
  name: UpdateDeviceRequest;
}

export interface FetchDevicesParams {
  perPage?: number;
  page?: number;
  filter?: string;
  status?: "accepted" | "pending" | "rejected";
  sortField?: string;
  sortOrder?: "asc" | "desc";
}

export interface IDeviceMethods {
  fetchDevices: (params: FetchDevicesParams) => Promise<void>;
  getList: () => IDevice[];
  getCount: () => number;
}
