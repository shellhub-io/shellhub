import { defineStore } from "pinia";
import { ref } from "vue";
import * as apiKeysApi from "../api/api_keys";
import { parseTotalCount } from "@/utils/headers";
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
    const res = await apiKeysApi.getApiKeys(
      page || 1,
      perPage || 10,
      sortOrder,
      sortField,
    );
    apiKeys.value = res.data as IApiKey[] || [];
    apiKeysCount.value = parseTotalCount(res.headers);
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
