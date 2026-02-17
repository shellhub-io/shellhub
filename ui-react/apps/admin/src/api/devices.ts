import apiClient from "./client";
import { Device } from "../types/device";
import { PaginatedResponse } from "../types/api";

// The API returns tags as Tag objects ({name, tenant_id, ...}) but the UI
// only needs tag names. Normalize at the boundary so the rest of the app
// can treat device.tags as string[].
function normalizeTags(device: Device): Device {
  return {
    ...device,
    tags: Array.isArray(device.tags)
      ? device.tags.map((t: unknown) =>
          typeof t === "object" && t !== null && "name" in t
            ? (t as { name: string }).name
            : String(t),
        )
      : device.tags,
  };
}

function buildFilter(tags: string[]): string {
  const filters: unknown[] = tags.map((tag) => ({
    type: "property",
    params: { name: "tags.name", operator: "eq", value: tag },
  }));
  if (filters.length > 1) {
    filters.push({ type: "operator", params: { name: "and" } });
  }
  return btoa(JSON.stringify(filters));
}

export async function getDevices(
  page = 1,
  perPage = 10,
  status = "",
  filterTags: string[] = [],
): Promise<PaginatedResponse<Device>> {
  const params: Record<string, string | number> = { page, per_page: perPage };
  if (status) params.status = status;
  if (filterTags.length > 0) params.filter = buildFilter(filterTags);

  const response = await apiClient.get<Device[]>("/api/devices", { params });
  const totalCount = parseInt(response.headers["x-total-count"] ?? "0", 10);
  return { data: response.data.map(normalizeTags), totalCount };
}

export async function getDevice(uid: string): Promise<Device> {
  const response = await apiClient.get<Device>(`/api/devices/${uid}`);
  return normalizeTags(response.data);
}

export async function renameDevice(uid: string, name: string): Promise<void> {
  await apiClient.put(`/api/devices/${uid}`, { name });
}

export async function acceptDevice(uid: string): Promise<void> {
  await apiClient.patch(`/api/devices/${uid}/accept`);
}

export async function rejectDevice(uid: string): Promise<void> {
  await apiClient.patch(`/api/devices/${uid}/reject`);
}

export async function removeDevice(uid: string): Promise<void> {
  await apiClient.delete(`/api/devices/${uid}`);
}

export async function addDeviceTag(uid: string, tag: string): Promise<void> {
  await apiClient.post(`/api/devices/${uid}/tags/${tag}`);
}

export async function removeDeviceTag(uid: string, tag: string): Promise<void> {
  await apiClient.delete(`/api/devices/${uid}/tags/${tag}`);
}
