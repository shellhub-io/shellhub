export interface NamespaceMember {
  id: string;
  role: string;
  email: string;
  added_at?: string;
}

export interface NamespaceSettings {
  session_record: boolean;
  connection_announcement: string;
}

export interface Namespace {
  name: string;
  owner: string;
  tenant_id: string;
  members: NamespaceMember[];
  max_devices: number;
  devices_accepted_count: number;
  devices_pending_count: number;
  devices_rejected_count: number;
  created_at: string;
  type: string;
  settings: NamespaceSettings | null;
}
