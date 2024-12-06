export interface ITunnel {
  address: string,
  namespace: string,
  device: string,
  host: string,
  port: number
}

export interface ITunnelCreate {
  uid: string,
  host: string,
  port: number
}

export interface ITunnelDelete{
  uid: string,
  address: string,
}
