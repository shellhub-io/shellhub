import { IDevice } from "./IDevice";

export interface IWebEndpoint {
  address: string,
  full_address: string,
  device: IDevice
  expires_in: string,
  device_uid: string,
  host: string,
  port: number
}

export interface IWebEndpointsCreate {
  uid: string,
  host: string,
  port: number,
  ttl: number
}

export interface FetchWebEndpointsParams {
  perPage?: number;
  page?: number;
  filter?: string;
  sortField?: "created_at" | "updated_at" | "address" | "uid";
  sortOrder?: "asc" | "desc";
}
