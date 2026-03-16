import apiClient from "./client";
import { Namespace, NamespaceMember } from "../types/namespace";

/* ─── Members ─── */

export async function getMembers(tenantId: string): Promise<NamespaceMember[]> {
  const { data } = await apiClient.get<Namespace>(`/api/namespaces/${tenantId}`);
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
