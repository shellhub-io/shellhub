import { CreateDeviceTagRequest, UpdateDeviceRequest, UpdateTagsDeviceRequest } from "@/api/client";

type identity = {
    mac: string;
  }

type infoDetails = {
  id: string;
  pretty_name: string;
  version: string;
  arch: string;
  platform: string;
}

type position = {
  latitude: number;
  longitude: number;
}

export interface IDevice {
  uid: string;
  name: string;
  identity: identity;
  info: infoDetails;
  public_key: string;
  tenant_id:string;
  last_seen: string;
  online: boolean;
  namespace: string;
  status: string;
  created_at: string;
  remoteAddr: string;
  position: position;
  tags: Array<string>;
}

export interface IUpdateDeviceTag {
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

export interface FetchDevicesParams {
  perPage?: number;
  page?: number;
  filter?: string;
  status?: "accepted" | "pending" | "rejected";
  sortStatusField: string;
  sortStatusString: string;
}
export interface IDeviceMethods {
    fetchDevices: (params: FetchDevicesParams) => Promise<void>;
    getFilter: () => string;
    getDevicesList: () => IDevice[];
    getSortStatusField: () => string;
    getSortStatusString: () => string;
    getNumberDevices: () => number;
}
