import { INamespace } from "@/interfaces/INamespace";
import { namespacesApi } from "../../api/http";

interface INamespaceSettings {
  connection_announcement: string;
  session_record: boolean;
}
interface INamespaceResponse {
  settings: INamespaceSettings;
  id: string;
  tenant_id: string;
  name: string;
  user_id: string;
  identifier: string;
  role: "administrator" | "operator" | "observer" | "owner";
}

export const postNamespace = async (data: string) => namespacesApi.createNamespace({ name: data });

export const fetchNamespaces = async (page: number, perPage: number, filter: string) => {
  if (filter) return namespacesApi.getNamespaces(filter, page, perPage);

  return namespacesApi.getNamespaces(filter, page, perPage);
};

export const getNamespace = async (id: string) => namespacesApi.getNamespace(id);

export const removeNamespace = async (id: string) => namespacesApi.deleteNamespace(id);

export const putNamespace = async (data: INamespaceResponse) => namespacesApi.editNamespace(data.id, {
  name: data.name,
  settings: {
    connection_announcement: data.settings.connection_announcement,
    session_record: data.settings.session_record,
  },
});

export const addUserToNamespace = async (data: INamespaceResponse) => namespacesApi.addNamespaceMember(data.tenant_id, {
  identifier: data.identifier,
  role: data.role,
});

export const editUserToNamespace = async (data: INamespaceResponse) => namespacesApi.updateNamespaceMember(data.tenant_id, data.user_id, {
  role: data.role,
});

export const removeUserFromNamespace = async (
  data: INamespaceResponse,
) => namespacesApi.removeNamespaceMember(data.tenant_id, data.user_id);

export const tenantSwitch = async (data: INamespace) => namespacesApi.getNamespaceToken(data.tenant_id);
