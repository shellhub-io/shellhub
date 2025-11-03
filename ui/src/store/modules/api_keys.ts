import { defineStore } from "pinia";
import { ref } from "vue";
import * as apiKeysApi from "../api/api_keys";
import { IApiKey, IApiKeyCreate, IApiKeyEdit, IApiKeyRemove } from "@/interfaces/IApiKey";

const useApiKeysStore = defineStore("apiKeys", () => {
  const apiKeys = ref<IApiKey[]>([]);
  const apiKeysCount = ref(0);

  const fetchApiKeys = async ({ page, perPage, sortOrder, sortField }: {
    page: number;
    perPage: number;
    sortOrder?: "asc" | "desc";
    sortField?: string;
  }) => {
    const res = await apiKeysApi.getApiKey(
      page || 1,
      perPage || 10,
      sortOrder,
      sortField,
    );
    apiKeys.value = res.data as IApiKey[] || [];
    apiKeysCount.value = parseInt(res.headers["x-total-count"] as string, 10) || 0;
  };

  const generateApiKey = async (data: IApiKeyCreate) => {
    const res = await apiKeysApi.generateApiKey(data);
    return res.data.id as string;
  };

  const editApiKey = async (data: IApiKeyEdit) => {
    await apiKeysApi.editApiKey(data);
  };

  const removeApiKey = async (data: IApiKeyRemove) => {
    await apiKeysApi.removeApiKey(data);
  };

  return {
    apiKeys,
    apiKeysCount,
    fetchApiKeys,
    generateApiKey,
    editApiKey,
    removeApiKey,
  };
});

export default useApiKeysStore;
