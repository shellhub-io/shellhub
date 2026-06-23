export interface Share {
  token: string;
  url: string;
  name: string;
  writable: boolean;
  command: string;
  device_uid: string;
  device_name: string;
  device_online: boolean;
  device_os: string;
  viewers: number;
  created_at: string;
  expires_at: string;
}
