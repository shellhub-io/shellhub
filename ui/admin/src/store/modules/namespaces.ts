import { defineStore } from "pinia";
import { ref } from "vue";
import { IAdminNamespace } from "@admin/interfaces/INamespace";
import * as namespacesApi from "../api/namespaces";

const useNamespacesStore = defineStore("namespace", () => {
  const namespaces = ref<IAdminNamespace[]>([]);
  const namespaceCount = ref(0);

  const fetchNamespaceList = async (data?: { page?: number; perPage?: number; filter?: string }) => {
    const res = await namespacesApi.fetchNamespaces(data?.page || 1, data?.perPage || 10, data?.filter);
    namespaces.value = res.data as IAdminNamespace[];
    namespaceCount.value = parseInt(res.headers["x-total-count"], 10);
  };

  const fetchNamespaceById = async (id: string) => {
    const { data } = await namespacesApi.getNamespace(id);
    return data as IAdminNamespace;
  };

  const exportNamespacesToCsv = async (filter: string) => {
    const { data } = await namespacesApi.exportNamespaces(filter);
    return data;
  };

  const updateNamespace = async (data: IAdminNamespace) => {
    await namespacesApi.updateNamespace(data);
  };

  return {
    namespaces,
    namespaceCount,
    fetchNamespaceList,
    fetchNamespaceById,
    exportNamespacesToCsv,
    updateNamespace,
  };
});

export default useNamespacesStore;
