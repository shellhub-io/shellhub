import { ConnectorTLS } from "@/api/client";

export interface IConnectorPayload {
  enable: boolean,
  secure: boolean,
  address: string,
  port: number,
  tls?: ConnectorTLS,
  uid: string,
}

export interface IConnector {
  uid: string,
  tenant_id: string,
  address: string,
  port: number,
  status: {
    state: string,
    message: string,
  },
  enable: boolean,
  secure: boolean,
}
