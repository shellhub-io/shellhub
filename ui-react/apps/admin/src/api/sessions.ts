import apiClient, { getTotalCount } from "./client";
import { Session } from "../types/session";
import { PaginatedResponse } from "../types/api";

export async function getSessions(
  page = 1,
  perPage = 10,
): Promise<PaginatedResponse<Session>> {
  const response = await apiClient.get<Session[]>("/api/sessions", {
    params: { page, per_page: perPage },
  });
  return { data: response.data, totalCount: getTotalCount(response) };
}

export async function getSession(uid: string): Promise<Session> {
  const response = await apiClient.get<Session>(`/api/sessions/${uid}`);
  return response.data;
}

export async function closeSession(
  uid: string,
  deviceUid: string,
): Promise<void> {
  await apiClient.post(`/api/sessions/${uid}/close`, { device: deviceUid });
}
