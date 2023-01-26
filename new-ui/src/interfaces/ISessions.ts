import { IDevice } from "./IDevice";

interface position {
  longitude: number;
  latitude: number;
}
export interface ISessions {
  active: boolean;
  authenticated: boolean;
  device: IDevice;
  device_uid: string;
  ip_address: string;
  last_seen: string;
  position: position;
  recorded: boolean;
  started_at: string;
  tenant_id: string;
  uid: string;
  username: string;
  term: string;
  type: string;
}
