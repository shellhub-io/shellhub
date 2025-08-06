export interface IWebEndpoints {
  address: string,
  full_address: string,
  expires_in: string,
  device: string,
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
