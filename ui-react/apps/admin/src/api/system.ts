import apiClient from "./client";
import type { SystemInfo, SetupRequest } from "../types/system";

export async function getInfo(): Promise<SystemInfo> {
  const { data } = await apiClient.get<SystemInfo>("/info");
  return data;
}

export async function setup(
  sign: string,
  payload: SetupRequest,
): Promise<void> {
  await apiClient.post("/api/setup", payload, {
    params: { sign },
  });
}
