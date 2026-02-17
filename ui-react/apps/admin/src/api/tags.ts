import apiClient from "./client";
import { Tag } from "../types/tag";

export async function getTags(
  page = 1,
  perPage = 100,
): Promise<{ data: Tag[]; totalCount: number }> {
  const res = await apiClient.get<Tag[]>("/api/tags", {
    params: { page, per_page: perPage },
  });
  const totalCount = parseInt(res.headers["x-total-count"] || "0", 10);
  return { data: res.data, totalCount };
}

export async function createTag(name: string): Promise<void> {
  await apiClient.post("/api/tags", { name });
}

export async function updateTag(currentName: string, newName: string): Promise<void> {
  await apiClient.patch(`/api/tags/${currentName}`, { name: newName });
}

export async function deleteTag(name: string): Promise<void> {
  await apiClient.delete(`/api/tags/${name}`);
}
