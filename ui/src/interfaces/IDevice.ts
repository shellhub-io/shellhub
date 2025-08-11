import { CreateDeviceTagRequest, UpdateDeviceRequest, UpdateTagsDeviceRequest } from "@/api/client";

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
  remote_addr: string;
  position: Position;
  tags: Array<string>;
}

export interface IUpdateDeviceTags {
  uid: string;
  tags: UpdateTagsDeviceRequest;
}

export interface IDeviceRename {
  uid: string;
  name: UpdateDeviceRequest;
}

export interface IDevicePostTag {
  uid: string;
  name: CreateDeviceTagRequest;
}

export interface SortDevicesParams {
  sortField?: string;
  sortOrder?: "asc" | "desc";
}

export interface FetchDevicesParams extends SortDevicesParams {
  perPage?: number;
  page?: number;
  filter?: string;
  status?: "accepted" | "pending" | "rejected";
}

export interface IDeviceMethods {
  fetchDevices: (params: FetchDevicesParams) => Promise<void>;
  setSort: (params: SortDevicesParams) => void;
  getFilter: () => string;
  getList: () => IDevice[];
  getSortStatusField: () => string;
  getSortStatusString: () => SortDevicesParams["sortOrder"];
  getNumber: () => number;
}
