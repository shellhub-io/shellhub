import apiClient from "./client";
import { Session } from "../types/session";
import { PaginatedResponse } from "../types/api";

export async function getSessions(
  page = 1,
  perPage = 10,
): Promise<PaginatedResponse<Session>> {
  const response = await apiClient.get<Session[]>("/api/sessions", {
    params: { page, per_page: perPage },
  });
  const totalCount = parseInt(response.headers["x-total-count"] ?? "0", 10);
  return { data: response.data, totalCount };
}

export async function getSession(uid: string): Promise<Session> {
  const response = await apiClient.get<Session>(`/api/sessions/${uid}`);
  return response.data;
}

export async function closeSession(uid: string): Promise<void> {
  await apiClient.post(`/api/sessions/${uid}/finish`);
}
