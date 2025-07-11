import {
  IApiKey, ApiKeyEdit,
  ApiKeyRemove } from "@/interfaces/IApiKey";
import { apiKeysApi } from "@/api/http";

export const generateApiKey = async (data: IApiKey) => apiKeysApi.apiKeyCreate({
  name: data.name,
  role: data.role,
  expires_at: data.expires_at,
});

export const getApiKey = async (
  page: number,
  perPage: number,
  sortStatusString?: "asc" | "desc",
  sortStatusField?: string,
) => apiKeysApi.apiKeyList(page, perPage, sortStatusString, sortStatusField);

export const removeApiKey = async (data: ApiKeyRemove) => apiKeysApi.apiKeyDelete(data.key);

export const editApiKey = async (data: ApiKeyEdit) => apiKeysApi.apiKeyUpdate(data.key, { name: data.name, role: data.role });
