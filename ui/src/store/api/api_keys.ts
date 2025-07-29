import {
  IApiKeyCreate, IApiKeyEdit, IApiKeyRemove,
} from "@/interfaces/IApiKey";
import { apiKeysApi } from "@/api/http";

export const generateApiKey = async (data: IApiKeyCreate) => apiKeysApi.apiKeyCreate({
  name: data.name,
  role: data.role,
  expires_at: data.expires_in,
});

export const getApiKey = async (
  page: number,
  perPage: number,
  sortOrder?: "asc" | "desc",
  sortField?: string,
) => apiKeysApi.apiKeyList(page, perPage, sortOrder, sortField);

export const removeApiKey = async (data: IApiKeyRemove) => apiKeysApi.apiKeyDelete(data.key);

export const editApiKey = async (data: IApiKeyEdit) => apiKeysApi.apiKeyUpdate(data.key, { name: data.name, role: data.role });
