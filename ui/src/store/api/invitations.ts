import { namespacesApi } from "@/api/http";
import { IInviteMemberPayload } from "@/interfaces/IInvitation";

export const fetchUserPendingInvitations = async (data: { filter: string, page: number, perPage: number }) =>
  namespacesApi.getMembershipInvitationList(data.filter, data.page, data.perPage);

export const fetchNamespaceInvitations = async (tenantId: string, data: { filter?: string, page: number, perPage: number }) =>
  namespacesApi.getNamespaceMembershipInvitationList(tenantId, data.filter, data.page, data.perPage);

export const declineNamespaceInvitation = async (tenant: string) =>
  namespacesApi.declineInvite(tenant);

export const acceptNamespaceInvitation = async (tenant: string) =>
  namespacesApi.acceptInvite(tenant);

export const sendNamespaceInvitationEmail = async (data: IInviteMemberPayload) => namespacesApi.addNamespaceMember(data.tenant_id, {
  email: data.email,
  role: data.role,
});

export const generateNamespaceInvitationLink = async (data: IInviteMemberPayload) => namespacesApi.generateInvitationLink(data.tenant_id, {
  email: data.email,
  role: data.role,
});
