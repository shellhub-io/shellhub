import { DeviceStatus, IDeviceRename } from "@/interfaces/IDevice";
import { devicesApi } from "@/api/http";

export const fetchDevices = async (
  page: number,
  perPage: number,
  status?: DeviceStatus,
  filter?: string,
  sortField?: string,
  sortOrder?: "asc" | "desc",
) => devicesApi.getDevices(
  filter,
  page,
  perPage,
  status,
  sortField,
  sortOrder,
);

export const resolveDevice = async (hostname?: string, uid?: string) => devicesApi.resolveDevice(hostname, uid);

export const renameDevice = async (data: IDeviceRename) => devicesApi.updateDevice(data.uid, data.name);

export const acceptDevice = async (uid: string) => devicesApi.updateDeviceStatus(uid, "accept");

export const rejectDevice = async (uid: string) => devicesApi.updateDeviceStatus(uid, "reject");

export const removeDevice = async (uid: string) => devicesApi.deleteDevice(uid);
