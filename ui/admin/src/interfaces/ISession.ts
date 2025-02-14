import { IDevice } from "./IDevice";

export interface ISessions {
  active: boolean;
  authenticated: boolean;
  device: IDevice;
  device_uid: string;
  ip_address: string;
  last_seen: string;
  recorded: boolean;
  started_at: string;
  tenant_id: string;
  uid: string;
  username: string;
  term: string;
  type: string;
}
