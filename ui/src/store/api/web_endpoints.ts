import { webEndpointsApi } from "@/api/http";
import { IWebEndpointsCreate } from "@/interfaces/IWebEndpoints";

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

export const createWebEndpoint = (payload: IWebEndpointsCreate) =>
  webEndpointsApi.createWebEndpoint(payload);

export const deleteWebEndpoint = (address: string) => webEndpointsApi.deleteWebEndpoint(address);
