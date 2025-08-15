import { IDevicePostTag, IDeviceRename, IUpdateDeviceTags } from "@/interfaces/IDevice";
import { devicesApi, tagsApi } from "@/api/http";

export const postTag = async (data: IDevicePostTag) => tagsApi.createDeviceTag(data.uid, data.name);

export const fetchDevices = async (
  page: number,
  perPage: number,
  status: "accepted" | "rejected" | "pending",
  filter?: string,
  sortStatusField?: string,
  sortStatusString?: "asc" | "desc",
) => devicesApi.getDevices(
  filter,
  page,
  perPage,
  status,
  sortStatusField,
  sortStatusString,
);

export const resolveDevice = async (hostname?: string, uid?: string) => devicesApi.resolveDevice(hostname, uid);

export const renameDevice = async (data: IDeviceRename) => devicesApi.updateDevice(data.uid, data.name);

export const acceptDevice = async (uid: string) => devicesApi.updateDeviceStatus(uid, "accept");

export const rejectDevice = async (uid: string) => devicesApi.updateDeviceStatus(uid, "reject");

export const updateDeviceTags = async (data: IUpdateDeviceTags) => devicesApi.updateTagsDevice(data.uid, data.tags);

export const removeDevice = async (uid: string) => devicesApi.deleteDevice(uid);
