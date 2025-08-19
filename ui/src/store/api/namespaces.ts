import {
  INamespaceAcceptInvite,
  INamespaceAddMember,
  INamespaceEdit,
  INamespaceEditMember,
  INamespaceRemoveMember,
} from "@/interfaces/INamespace";
import { namespacesApi } from "@/api/http";

export const createNamespace = async (name: string) => namespacesApi.createNamespace({ name });

export const fetchNamespaces = async (page: number, perPage: number, filter?: string) => namespacesApi.getNamespaces(filter, page, perPage);

export const getNamespace = async (id: string) => namespacesApi.getNamespace(id);

export const deleteNamespace = async (id: string) => namespacesApi.deleteNamespace(id);

export const leaveNamespace = async (tenant: string) => namespacesApi.leaveNamespace(tenant);

export const editNamespace = async (data: INamespaceEdit) => namespacesApi.editNamespace(data.tenant_id, {
  name: data.name,
  settings: {
    connection_announcement: data.settings?.connection_announcement,
    session_record: data.settings?.session_record,
  },
});

export const sendNamespaceLink = async (data: INamespaceAddMember) => namespacesApi.addNamespaceMember(data.tenant_id, {
  email: data.email,
  role: data.role,
});

export const generateNamespaceLink = async (data: INamespaceAddMember) => namespacesApi.generateInvitationLink(data.tenant_id, {
  email: data.email,
  role: data.role,
});

export const updateNamespaceMember = async (data: INamespaceEditMember) => namespacesApi.updateNamespaceMember(
  data.tenant_id,
  data.user_id,
  { role: data.role },
);

export const removeUserFromNamespace = async (data: INamespaceRemoveMember) => namespacesApi.removeNamespaceMember(
  data.tenant_id,
  data.user_id,
);

export const switchNamespace = async (tenantId: string) => namespacesApi.getNamespaceToken(tenantId);

export const acceptNamespaceInvite = async (data: INamespaceAcceptInvite) => namespacesApi.acceptInvite(data.tenant, { sig: data.sig });

export const getSupportID = async (tenant: string) => namespacesApi.getNamespaceSupport(tenant);

export const lookupUserStatus = async (
  data: { tenant: string; id: string; sig: string; },
) => namespacesApi.lookupUserStatus(data.tenant, data.id, data.sig);
