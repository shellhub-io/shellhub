import http from "../helpers/http";
import { namespacesApi } from "../../api/http";

export const postNamespace = async (data: any) => namespacesApi.createNamespace({ name: data });

export const fetchNamespaces = async (page: any, perPage: any, filter: any) => {
  if (filter) return namespacesApi.getNamespaces(filter, page, perPage);

  return namespacesApi.getNamespaces(filter, page, perPage);
};

export const getNamespace = async (id: any) => namespacesApi.getNamespace(id);

export const removeNamespace = async (id: any) =>
  namespacesApi.deleteNamespace(id);

export const putNamespace = async (data: any) =>
  namespacesApi.editNamespace(data.id, { name: data.name });

export const addUserToNamespace = async (data: any) =>
  namespacesApi.addNamespaceMember(data.tenant_id, {
    username: data.username,
    role: data.role,
  });

export const editUserToNamespace = async (data: any) =>
  namespacesApi.updateNamespaceMember(data.tenant_id, data.user_id, {
    role: data.role,
  });

export const removeUserFromNamespace = async (data: any) =>
  namespacesApi.removeNamespaceMember(data.tenant_id, data.user_id);

export const tenantSwitch = async (data: any) =>
  namespacesApi.getNamespaceToken(data.tenant_id);
