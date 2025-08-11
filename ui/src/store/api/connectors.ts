import { IConnectorPayload } from "@/interfaces/IConnector";
import { namespacesApi } from "@/api/http";

export const createConnector = async (data: Omit<IConnectorPayload, "uid">) => namespacesApi.connectorCreate({
  enable: data.enable,
  secure: data.secure,
  address: data.address,
  port: data.port,
  tls: data.tls,
});

export const getConnectorList = async (page: number, perPage: number) => namespacesApi.connectorList(undefined, page, perPage);

export const updateConnector = async (data: IConnectorPayload) => namespacesApi.connectorUpdate(data.uid, {
  enable: data.enable,
  secure: data.secure,
  address: data.address,
  port: data.port,
  tls: data.tls,
});

export const deleteConnector = async (uid: string) => namespacesApi.connectorDelete(uid);

export const getConnector = async (uid: string) => namespacesApi.connectorGet(uid);

export const getConnectorInfo = async (uid: string) => namespacesApi.connectorInfo(uid);
