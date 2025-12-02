import { defineStore } from "pinia";
import { computed, ref } from "vue";
import * as namespacesApi from "../api/namespaces";
import {
  INamespace,
  INamespaceEdit,
  INamespaceEditMember,
  INamespaceRemoveMember,
} from "@/interfaces/INamespace";

const useNamespacesStore = defineStore("namespaces", () => {
  const currentNamespace = ref<INamespace>({} as INamespace);
  const namespaceList = ref<Array<INamespace>>([]);
  const userStatus = ref<string>();
  const hasNamespaces = computed(() => namespaceList.value.length > 0);

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

  const updateNamespaceMember = async (data: INamespaceEditMember) => {
    await namespacesApi.updateNamespaceMember(data);
  };

  const removeMemberFromNamespace = async (data: INamespaceRemoveMember) => {
    await namespacesApi.removeUserFromNamespace(data);
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
    hasNamespaces,

    createNamespace,
    fetchNamespaceList,
    fetchNamespace,
    editNamespace,
    deleteNamespace,
    leaveNamespace,
    updateNamespaceMember,
    removeMemberFromNamespace,
    lookupUserStatus,
    switchNamespace,
    reset,
  };
});

export default useNamespacesStore;
