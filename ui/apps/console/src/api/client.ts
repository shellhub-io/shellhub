import axios, { type AxiosResponse } from "axios";
import { setupInterceptors } from "./interceptors";

const apiClient = axios.create({
  baseURL: `${window.location.protocol}//${window.location.host}`,
});

setupInterceptors(apiClient);

export function getTotalCount(response: AxiosResponse): number {
  return Number(response.headers["x-total-count"]) || 0;
}

export default apiClient;
