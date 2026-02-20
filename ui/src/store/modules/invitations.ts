import { ref } from "vue";
import { defineStore } from "pinia";
import { IInvitation, IInviteMemberPayload } from "@/interfaces/IInvitation";
import * as invitationsApi from "../api/invitations";
import { parseTotalCount } from "@/utils/headers";
import { BasicRole } from "@/interfaces/INamespace";
import { getInvitationStatusFilter } from "@/utils/invitations";

const useInvitationsStore = defineStore("invitations", () => {
  const pendingInvitations = ref<IInvitation[]>([]);
  const namespaceInvitations = ref<IInvitation[]>([]);
  const invitationCount = ref(0);

  const pendingInvitesFilter = getInvitationStatusFilter("pending");

  const fetchUserPendingInvitationList = async () => {
    const res = await invitationsApi.fetchUserPendingInvitations({
      filter: pendingInvitesFilter,
      page: 1,
      perPage: 100,
    });
    pendingInvitations.value = res.data as IInvitation[];
  };

  const fetchNamespaceInvitationList = async (
    tenantId: string,
    page: number,
    perPage: number,
    filter?: string,
  ) => {
    const res = await invitationsApi.fetchNamespaceInvitations(tenantId, {
      filter,
      page,
      perPage,
    });
    namespaceInvitations.value = res.data as IInvitation[];
    invitationCount.value = parseTotalCount(res.headers);
  };

  const acceptInvitation = async (tenant: string) => {
    await invitationsApi.acceptNamespaceInvitation(tenant);
  };

  const declineInvitation = async (tenant: string) => {
    await invitationsApi.declineNamespaceInvitation(tenant);
  };

  const editInvitation = async (data: { tenant: string; user_id: string; role: BasicRole }) => {
    await invitationsApi.editNamespaceInvitation(data);
  };

  const cancelInvitation = async (data: { tenant: string; user_id: string }) => {
    await invitationsApi.cancelNamespaceInvitation(data);
  };

  const sendInvitationEmail = async (data: IInviteMemberPayload) => {
    await invitationsApi.sendNamespaceInvitationEmail(data);
  };

  const generateInvitationLink = async (data: IInviteMemberPayload) => {
    const response = await invitationsApi.generateNamespaceInvitationLink(data);
    return response.data.link as string;
  };

  return {
    pendingInvitations,
    namespaceInvitations,
    invitationCount,
    fetchUserPendingInvitationList,
    fetchNamespaceInvitationList,
    acceptInvitation,
    declineInvitation,
    editInvitation,
    cancelInvitation,
    sendInvitationEmail,
    generateInvitationLink,
  };
});

export default useInvitationsStore;
