import { IBilling } from "./IBilling";

export type Role = "administrator" | "operator" | "observer" | "owner";

export interface INamespaceMember {
  id: string;
  username: string;
  role: Role;
}

export interface INamespaceSettings {
  connection_announcement?: string;
  session_record: boolean;
}

export interface INamespace {
  name: string;
  owner: string;
  tenant_id: string;
  members: INamespaceMember[];
  settings: INamespaceSettings;
  max_devices: number;
  devices_accepted_count: number;
  devices_pending_count: number;
  devices_rejected_count: number;
  created_at: string;
  billing: IBilling;
}

export interface INamespaceAcceptInvite {
  tenant: string;
  sig: string;
}

export interface INamespaceAddMember {
  email: string;
  role: Role;
  tenant_id: string;
}
