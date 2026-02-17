import apiClient from "./client";
import { PaginatedResponse } from "../types/api";
import { NamespaceMember } from "../types/namespace";
import { ApiKey } from "../types/apiKey";

/* ─── Members ─── */

export async function getMembers(tenantId: string): Promise<NamespaceMember[]> {
  const { data } = await apiClient.get(`/api/namespaces/${tenantId}`);
  return data.members ?? [];
}

export async function addMember(
  tenantId: string,
  payload: { email: string; role: string },
): Promise<void> {
  await apiClient.post(`/api/namespaces/${tenantId}/members`, payload);
}

export async function updateMemberRole(
  tenantId: string,
  userId: string,
  role: string,
): Promise<void> {
  await apiClient.put(`/api/namespaces/${tenantId}/members/${userId}`, {
    role,
  });
}

export async function removeMember(
  tenantId: string,
  userId: string,
): Promise<void> {
  await apiClient.delete(`/api/namespaces/${tenantId}/members/${userId}`);
}

/* ─── API Keys ─── */

export async function getApiKeys(
  page = 1,
  perPage = 10,
  sort = "expires_in",
  order: "asc" | "desc" = "desc",
): Promise<PaginatedResponse<ApiKey>> {
  const res = await apiClient.get<ApiKey[]>("/api/namespaces/api-key", {
    params: { page, per_page: perPage, sort, order },
  });
  const totalCount = parseInt(res.headers["x-total-count"] || "0", 10);
  return { data: res.data, totalCount };
}

export async function generateApiKey(payload: {
  name: string;
  role: string;
  expires_at: number;
}): Promise<ApiKey> {
  const { data } = await apiClient.post<ApiKey>(
    "/api/namespaces/api-key",
    payload,
  );
  return data;
}

export async function updateApiKey(
  name: string,
  payload: { name?: string; role?: string },
): Promise<void> {
  await apiClient.patch(`/api/namespaces/api-key/${name}`, payload);
}

export async function deleteApiKey(name: string): Promise<void> {
  await apiClient.delete(`/api/namespaces/api-key/${name}`);
}
