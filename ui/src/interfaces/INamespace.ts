export type BasicRole = "administrator" | "operator" | "observer";
export type Role = BasicRole | "owner";

export interface INamespaceMember {
  id: string;
  role: Role;
  email: string;
  status: string;
  added_at: string;
  expires_at: string;
}

export interface INamespaceSettings {
  connection_announcement?: string;
  session_record: boolean;
}

interface INamespaceBilling {
  customer_id: string;
  subscription_id: string;
  current_period_end: number;
  active: boolean;
  status: string;
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
  billing: INamespaceBilling | null;
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

export interface INamespaceEditMember {
  user_id: string;
  role: Role;
  tenant_id: string;
}

export interface INamespaceRemoveMember {
  tenant_id: string;
  user_id: string;
}

export interface INamespaceEdit {
  tenant_id: string;
  name?: string;
  settings?: Partial<INamespaceSettings>;
}
