export interface IStats {
  registered_devices: number;
  online_devices: number;
  active_sessions: number;
  pending_devices: number;
  rejected_devices: number;
}

export interface StatCardItem {
  title: string;
  content: string;
  icon: string;
  buttonLabel: string;
  path: string;
  stat: number;
}
