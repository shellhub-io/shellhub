import { Device } from "./device";

export interface WebEndpointTLS {
  enabled: boolean;
  verify: boolean;
  domain: string;
}

export interface WebEndpoint {
  address: string;
  full_address: string;
  device: Device;
  device_uid: string;
  host: string;
  port: number;
  tls?: WebEndpointTLS;
  expires_in: string;
  created_at: string;
}

export interface WebEndpointCreate {
  uid: string;
  host: string;
  port: number;
  ttl: number;
  tls?: WebEndpointTLS;
}
