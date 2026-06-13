export interface Connection {
  id: string;
  tenant_id: string;
  label: string;
  username: string;
  kind: string;
  host: string;
  port: number;
  device_uid: string;
  created_at: string;
  updated_at: string;
}
