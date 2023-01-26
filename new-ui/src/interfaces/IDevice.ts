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
