import { IBilling } from "./IBilling";

export interface INamespaceMember {
  id: string;
  username: string;
  role: string;
}

export interface settings {
  session_record: boolean;
  connection_announcement: string;
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

export interface INamespaceInvite {
  tenant: string;
  sig: string;
}

export interface INamespaceSettings {
  connection_announcement?: string;
  session_record?: boolean;
}
export interface INamespaceResponse {
  settings?: INamespaceSettings;
  id: string;
  tenant_id: string;
  name: string;
  user_id: string;
  email: string;
  role: "administrator" | "operator" | "observer" | "owner";
}
