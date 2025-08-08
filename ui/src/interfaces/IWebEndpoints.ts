import { IDevice } from "./IDevice";

export interface IWebEndpoints {
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

export interface IWebEndpointsDelete {
  address: string,
}
