import apiClient from "./client";
import { WebEndpoint, WebEndpointCreate } from "../types/webEndpoint";

export async function getWebEndpoints(
  page = 1,
  perPage = 10,
  filter?: string,
): Promise<{ data: WebEndpoint[]; totalCount: number }> {
  const params: Record<string, string | number> = { page, per_page: perPage };
  if (filter) params.filter = filter;

  const res = await apiClient.get<WebEndpoint[]>("/api/web-endpoints", {
    params,
  });
  const totalCount = parseInt(res.headers["x-total-count"] || "0", 10);
  return { data: res.data, totalCount };
}

export async function createWebEndpoint(
  payload: WebEndpointCreate,
): Promise<WebEndpoint> {
  const { data } = await apiClient.post<WebEndpoint>(
    "/api/web-endpoints",
    payload,
  );
  return data;
}

export async function deleteWebEndpoint(address: string): Promise<void> {
  await apiClient.delete(`/api/web-endpoints/${address}`);
}
