import { ConnectorTLS } from "@/api/client";

export interface IConnectorCreate {
  enable: boolean,
  secure: boolean,
  address: string,
  port: number,
  tls: ConnectorTLS,
}

export interface IConnectorUpdate {
  uid: string,
  enable: boolean,
  secure: boolean,
  address: string,
  port: number,
  tls: ConnectorTLS,
}

export interface IConnectorPayload {
  enable: boolean;
  secure: boolean;
  address: string;
  port: number;
  uid?: string;
  tls?: ConnectorTLS | unknown;
}

export interface IConnector {
    uid: string,
    tenant_id: string,
    enable: boolean,
    secure: boolean,
    hostname: string,
    tls: null,
}
