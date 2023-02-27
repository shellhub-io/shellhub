import { IDevicePostTag, IDeviceRename, IUpdateDeviceTag } from "@/interfaces/IDevice";
import { devicesApi, tagsApi } from "../../api/http";

export const postTag = async (data: IDevicePostTag) => tagsApi.createDeviceTag(data.uid, data.name);

export const fetchDevices = async (
  page : number,
  perPage: number,
  filter : string | undefined,
  status : "accepted" | "rejected" | "pending" | "unused",
  sortStatusField : string | undefined,
  sortStatusString : "asc" | "desc" | "",
) => {
  if (sortStatusField && sortStatusString) {
    return devicesApi.getDevices(
      filter,
      page,
      perPage,
      status,
      sortStatusField,
      sortStatusString,
    );
  }

  return devicesApi.getDevices(filter, page, perPage, status);
};

export const getDevice = async (uid : string) => devicesApi.getDevice(uid);

export const renameDevice = async (data : IDeviceRename) => devicesApi.updateDeviceName(data.uid, data.name);

export const acceptDevice = async (uid : string) => devicesApi.updateDeviceStatus(uid, "accept");

export const rejectDevice = async (uid : string) => devicesApi.updateDeviceStatus(uid, "reject");

export const updateDeviceTag = async (data : IUpdateDeviceTag) => devicesApi.updateTagsDevice(data.uid, data.tags);

export const removeDevice = async (uid : string) => devicesApi.deleteDevice(uid);
