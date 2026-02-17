import apiClient from "./client";
import { Stats } from "../types/stats";

export async function getStats(): Promise<Stats> {
  const response = await apiClient.get<Stats>("/api/stats");
  return response.data;
}
