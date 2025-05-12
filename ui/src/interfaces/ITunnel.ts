export interface ITunnel {
  address: string,
  full_address: string,
  expires_in: string,
  device: string,
  host: string,
  port: number
}

export interface ITunnelCreate {
  uid: string,
  host: string,
  port: number,
  ttl: number
}

export interface ITunnelDelete{
  uid: string,
  address: string,
}
