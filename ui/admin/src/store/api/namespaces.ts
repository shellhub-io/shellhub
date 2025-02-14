import { INamespace } from "../../interfaces/INamespace";
import { adminApi } from "./../../api/http";

export const fetchNamespaces = async (
  page: number,
  perPage: number,
  search: string,
) => adminApi.getNamespacesAdmin(search, page, perPage);

export const exportNamespaces = async (filter: string) => adminApi.exportNamespaces(filter);

export const getNamespace = async (id: string) => adminApi.getNamespaceAdmin(id);

export const updateNamespace = async (
  data: INamespace,
) => adminApi.editNamespaceAdmin(data.tenant_id, {
  ...data,
});
