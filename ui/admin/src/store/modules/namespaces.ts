import { defineStore } from "pinia";
import { ref } from "vue";
import { IAdminNamespace } from "@admin/interfaces/INamespace";
import * as namespacesApi from "../api/namespaces";

const useNamespacesStore = defineStore("adminNamespaces", () => {
  const namespaces = ref<IAdminNamespace[]>([]);
  const namespaceCount = ref(0);
  const namespace = ref<IAdminNamespace>({} as IAdminNamespace);

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
    namespaceCount.value = parseInt(res.headers["x-total-count"] as string, 10);
  };

  const fetchNamespaceById = async (id: string) => {
    const { data } = await namespacesApi.getNamespace(id);
    namespace.value = data as IAdminNamespace;
  };

  const exportNamespacesToCsv = async (filter: string) => {
    const { data } = await namespacesApi.exportNamespaces(filter);
    return data;
  };

  const deleteNamespace = async (tenant: string) => {
    await namespacesApi.deleteNamespace(tenant);
  };

  const updateNamespace = async (data: IAdminNamespace) => {
    await namespacesApi.updateNamespace(data);
  };

  return {
    namespaces,
    namespaceCount,
    namespace,
    currentFilter,
    setFilter,
    fetchNamespaceList,
    fetchNamespaceById,
    exportNamespacesToCsv,
    deleteNamespace,
    updateNamespace,
  };
});

export default useNamespacesStore;
