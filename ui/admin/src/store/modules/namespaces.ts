import { defineStore } from "pinia";
import { ref } from "vue";
import { IAdminNamespace } from "@admin/interfaces/INamespace";
import * as namespacesApi from "../api/namespaces";

const useNamespacesStore = defineStore("namespaces", () => {
  const namespaces = ref<IAdminNamespace[]>([]);
  const namespaceCount = ref(0);

  const currentFilter = ref<string>("");

  const setFilter = (filter: string) => {
    currentFilter.value = filter || "";
  };

  const fetchNamespaceList = async (data?: { page?: number; perPage?: number; filter?: string }) => {
    const page = data?.page || 1;
    const perPage = data?.perPage || 10;
    const filter = data?.filter ?? currentFilter.value ?? "";

    const res = await namespacesApi.fetchNamespaces(page, perPage, filter);
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
    currentFilter,
    setFilter,
    fetchNamespaceList,
    fetchNamespaceById,
    exportNamespacesToCsv,
    updateNamespace,
  };
});

export default useNamespacesStore;
