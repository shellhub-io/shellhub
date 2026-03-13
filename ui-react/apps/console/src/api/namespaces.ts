import apiClient from "./client";
import { Namespace } from "../types/namespace";

export async function getNamespaces(
  page = 1,
  perPage = 30,
): Promise<Namespace[]> {
  const { data } = await apiClient.get<Namespace[]>("/api/namespaces", {
    params: { page, per_page: perPage },
  });
  return data;
}

export async function getNamespace(tenantId: string): Promise<Namespace> {
  const { data } = await apiClient.get<Namespace>(
    `/api/namespaces/${tenantId}`,
  );
  return data;
}

export async function createNamespace(
  name: string,
): Promise<Namespace> {
  const { data } = await apiClient.post<Namespace>("/api/namespaces", { name });
  return data;
}

export async function getNamespaceToken(
  tenantId: string,
): Promise<{ token: string; role: string }> {
  const { data } = await apiClient.get<{ token: string; role: string }>(
    `/api/auth/token/${tenantId}`,
  );
  return data;
}

export async function updateNamespace(
  tenantId: string,
  body: { name?: string; settings?: { session_record?: boolean; connection_announcement?: string } },
): Promise<void> {
  await apiClient.put(`/api/namespaces/${tenantId}`, body);
}

export async function deleteNamespace(tenantId: string): Promise<void> {
  await apiClient.delete(`/api/namespaces/${tenantId}`);
}

export async function leaveNamespace(tenantId: string): Promise<void> {
  await apiClient.delete(`/api/namespaces/${tenantId}/members`);
}
