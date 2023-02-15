import { IBilling } from "./IBilling";

export interface INamespaceMember {
  id: string;
  username: string;
  role: string;
}

export interface settings {
  session_record: boolean;
}

export interface INamespace {
  name: string;
  owner: string;
  tenant_id: string;
  members: INamespaceMember[];
  settings: settings;
  max_devices: number;
  devices_count: number;
  created_at: string;
  billing: IBilling;
}
