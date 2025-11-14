import { defineStore } from "pinia";
import { ref } from "vue";
import * as namespacesApi from "../api/namespaces";
import {
  INamespace,
  INamespaceAcceptInvite,
  INamespaceAddMember,
  INamespaceEdit,
  INamespaceEditMember,
  INamespaceRemoveMember,
} from "@/interfaces/INamespace";

const useNamespacesStore = defineStore("namespaces", () => {
  const currentNamespace = ref<INamespace>({} as INamespace);
  const namespaceList = ref<Array<INamespace>>([]);
  const userStatus = ref<string>();

  const fetchNamespaceList = async (data?: { page?: number; perPage?: number; filter?: string }) => {
    const res = await namespacesApi.fetchNamespaces(data?.page || 1, data?.perPage || 10, data?.filter);
    namespaceList.value = res.data as INamespace[];
  };

  const fetchNamespace = async (id: string) => {
    const res = await namespacesApi.getNamespace(id);
    currentNamespace.value = res.data as INamespace;
  };

  const createNamespace = async (name: string) => {
    const res = await namespacesApi.createNamespace(name);
    return res.data.tenant_id as string;
  };

  const editNamespace = async (data: INamespaceEdit) => {
    const res = await namespacesApi.editNamespace(data);
    currentNamespace.value = res.data as INamespace;
  };

  const deleteNamespace = async (id: string) => {
    await namespacesApi.deleteNamespace(id);
    currentNamespace.value = {} as INamespace;
    namespaceList.value = [];
  };

  const leaveNamespace = async (tenant: string) => {
    const res = await namespacesApi.leaveNamespace(tenant);

    localStorage.setItem("token", res.data.token || "");

    if (res.data.tenant) {
      localStorage.setItem("tenant", res.data.tenant || "");
      localStorage.setItem("role", res.data.role || "");
    }
  };

  const sendEmailInvitation = async (data: INamespaceAddMember) => {
    await namespacesApi.sendNamespaceLink(data);
  };

  const generateInvitationLink = async (data: INamespaceAddMember) => {
    const res = await namespacesApi.generateNamespaceLink(data);
    return res.data.link as string;
  };

  const updateNamespaceMember = async (data: INamespaceEditMember) => {
    await namespacesApi.updateNamespaceMember(data);
  };

  const removeMemberFromNamespace = async (data: INamespaceRemoveMember) => {
    await namespacesApi.removeUserFromNamespace(data);
  };

  const acceptInvite = async (data: INamespaceAcceptInvite) => {
    await namespacesApi.acceptNamespaceInvite(data);
  };

  const lookupUserStatus = async (data: { tenant: string; id: string; sig: string; }) => {
    const res = await namespacesApi.lookupUserStatus(data);
    userStatus.value = res.data.status;
  };

  const switchNamespace = async (tenantId: string) => {
    localStorage.removeItem("role");

    const res = await namespacesApi.switchNamespace(tenantId);
    if (res.status === 200) {
      localStorage.setItem("token", res.data.token || "");
      localStorage.setItem("tenant", tenantId);
      localStorage.setItem("role", res.data.role || "");
    }
  };

  const reset = () => {
    currentNamespace.value = {} as INamespace;
    namespaceList.value = [];
    userStatus.value = undefined;
  };

  return {
    currentNamespace,
    namespaceList,
    userStatus,

    createNamespace,
    fetchNamespaceList,
    fetchNamespace,
    editNamespace,
    deleteNamespace,
    leaveNamespace,
    sendEmailInvitation,
    generateInvitationLink,
    updateNamespaceMember,
    removeMemberFromNamespace,
    acceptInvite,
    lookupUserStatus,
    switchNamespace,
    reset,
  };
});

export default useNamespacesStore;
