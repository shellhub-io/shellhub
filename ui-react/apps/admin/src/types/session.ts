export interface Session {
  uid: string;
  device_uid: string;
  device: {
    uid: string;
    name: string;
    info?: {
      id: string;
      pretty_name: string;
      version: string;
      arch: string;
      platform: string;
    };
    namespace?: string;
    online?: boolean;
  };
  tenant_id: string;
  username: string;
  ip_address: string;
  started_at: string;
  last_seen: string;
  active: boolean;
  authenticated: boolean;
  recorded: boolean;
  type: string;
  term: string;
  position: {
    latitude: number;
    longitude: number;
  };
  events: {
    types: string[];
    seats: number[];
  };
}
