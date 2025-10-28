import { adminApi } from "@/api/http";

export const getDevices = (
  page: number,
  perPage: number,
  filter?: string,
  sortField?: string,
  sortOrder?: "asc" | "desc",
) => adminApi.getDevicesAdmin(
  filter,
  page,
  perPage,
  undefined, // status
  sortField,
  sortOrder,
);

export const getDevice = (uid: string) => adminApi.getDeviceAdmin(uid);
