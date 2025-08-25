import { webEndpointsApi } from "@/api/http";

export const getWebEndpoints = (
  page: number,
  perPage: number,
  filter?: string,
  sortField?: "created_at" | "updated_at" | "address" | "uid",
  sortOrder?: "asc" | "desc",
) => webEndpointsApi.listWebEndpoints(
  filter,
  page,
  perPage,
  sortField,
  sortOrder,
);

export const createWebEndpoint = (uid: string, host: string, port: number, ttl: number) => webEndpointsApi.createWebEndpoint(
  {
    uid,
    host,
    port,
    ttl,
  },
);

export const deleteWebEndpoint = (address: string) => webEndpointsApi.deleteWebEndpoint(address);
