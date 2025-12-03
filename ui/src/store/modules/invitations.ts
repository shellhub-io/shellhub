import { ref } from "vue";
import { defineStore } from "pinia";
import { IInvitation, IInviteMemberPayload } from "@/interfaces/IInvitation";
import * as invitationsApi from "../api/invitations";

const useInvitationsStore = defineStore("invitations", () => {
  const pendingInvitations = ref<IInvitation[]>([]);
  const namespaceInvitations = ref<IInvitation[]>([]);

  const pendingInvitesFilter = JSON.stringify([{ type: "property", params: { name: "status", operator: "eq", value: "pending" } }]);
  const encodedFilter = Buffer.from(pendingInvitesFilter).toString("base64");

  const fetchUserPendingInvitationList = async () => {
    const res = await invitationsApi.fetchUserPendingInvitations({
      filter: encodedFilter,
      page: 1,
      perPage: 100,
    });
    pendingInvitations.value = res.data as IInvitation[];
  };

  const fetchNamespaceInvitationList = async (tenantId: string, page: number, perPage: number) => {
    const res = await invitationsApi.fetchNamespaceInvitations(tenantId, {
      filter: undefined,
      page,
      perPage,
    });
    namespaceInvitations.value = res.data as IInvitation[];
  };

  const acceptInvitation = async (tenant: string) => {
    await invitationsApi.acceptNamespaceInvitation(tenant);
  };

  const declineInvitation = async (tenant: string) => {
    await invitationsApi.declineNamespaceInvitation(tenant);
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
    fetchUserPendingInvitationList,
    fetchNamespaceInvitationList,
    acceptInvitation,
    declineInvitation,
    sendInvitationEmail,
    generateInvitationLink,
  };
});

export default useInvitationsStore;
