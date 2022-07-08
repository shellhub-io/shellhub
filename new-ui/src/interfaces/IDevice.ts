type identity = {
    mac: string;
  }
  
  type infoDetails = {
    arch: string;
    id: string;
    platform: string;
    pretty_name: string;
    version: string;
  }
  
  type position = {
    latitude: number;
    longitude: number;
  }
  
  export interface IDevice {
    created_at: string;
    identity: identity;
    info: infoDetails;
    last_seen: string;
    name: string;
    namespace: string;
    online: boolean;
    position: position;
    public_key: string;
    remoteAddr: string;
    status: string;
    tags: Array<string>;
    tenant_id:string;
    uid: string;
  }
  