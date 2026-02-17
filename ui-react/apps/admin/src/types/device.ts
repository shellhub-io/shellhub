export interface DeviceIdentity {
  mac: string;
}

export interface DeviceInfo {
  id: string;
  pretty_name: string;
  version: string;
  arch: string;
  platform: string;
}

export interface DevicePosition {
  latitude: number;
  longitude: number;
}

export interface Device {
  uid: string;
  name: string;
  identity: DeviceIdentity;
  info: DeviceInfo;
  public_key: string;
  tenant_id: string;
  last_seen: string;
  online: boolean;
  namespace: string;
  status: string;
  created_at: string;
  status_updated_at: string;
  remote_addr: string;
  position: DevicePosition;
  tags: string[];
  public_url: boolean;
  acceptable: boolean;
}
