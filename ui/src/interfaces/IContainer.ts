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

export interface IContainer {
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

export interface IUpdateContainerTags {
  uid: string;
  tags: UpdateTagsDeviceRequest;
}

export interface IUpdateContainerTag {
  uid: string;
  tags: string;
}

export interface IContainerRename {
  uid: string;
  name: UpdateDeviceRequest;
}

export interface IContainerPostTag {
  uid: string;
  name: CreateDeviceTagRequest;
}

export interface FetchContainerParams {
  perPage?: number;
  page?: number;
  filter?: string;
  status?: "accepted" | "pending" | "rejected";
  sortStatusField: string;
  sortStatusString: string;
}
export interface IContainerMethods {
    fetchDevices: (params: FetchContainerParams) => Promise<void>;
    getFilter: () => string;
    getList: () => IContainer[];
    getSortStatusField: () => string;
    getSortStatusString: () => string;
    getNumber: () => number;
}
