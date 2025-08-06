import { webEndpointsApi } from "@/api/http";

const getWebEndpoints = (
  filter: string | undefined,
  page: number,
  perPage: number,
  sortBy?: "created_at" | "updated_at" | "address" | "uid",
  orderBy?: "asc" | "desc",
) => webEndpointsApi.listWebEndpoints(
  filter,
  page,
  perPage,
  sortBy,
  orderBy,
);

const createWebEndpoints = (uid: string, host: string, port: number, ttl: number) => webEndpointsApi.createWebEndpoint(
  {
    uid,
    host,
    port,
    ttl,
  },
);

const deleteWebEndpoints = (address: string) => webEndpointsApi.deleteWebEndpoint(address);

export { getWebEndpoints, createWebEndpoints, deleteWebEndpoints };
