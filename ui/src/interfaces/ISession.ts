import { IDevice } from "./IDevice";

interface Position {
  longitude: number;
  latitude: number;
}
export interface ISession {
  active: boolean;
  authenticated: boolean;
  device: IDevice;
  device_uid: string;
  ip_address: string;
  last_seen: string;
  position: Position;
  recorded: boolean;
  started_at: string;
  tenant_id: string;
  uid: string;
  username: string;
  term: string;
  type: string;
}
