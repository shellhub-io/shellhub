import { devicesApi, tagsApi } from "../../api/http";


export const postTag = async (data: any) => tagsApi.createDeviceTag(data.uid, data.name);

export const fetchDevices = async (
  page : any,
  perPage: any,
  filter : any,
  status : any,
  sortStatusField : any,
  sortStatusString : any,
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

export const getDevice = async (uid : any) =>  devicesApi.getDevice(uid);


export const renameDevice = async (data : any) =>  devicesApi.updateDeviceName(data.uid, data.name);

export const acceptDevice = async (uid : any) => devicesApi.updateDeviceStatus(uid ,"accept");

export const rejectDevice = async (uid : any) => devicesApi.updateDeviceStatus(uid ,"reject");

export const updateDeviceTag = async (data : any) => devicesApi.updateTagsDevice(data.uid, data.tags);

export const removeDevice = async (uid : any) => devicesApi.deleteDevice(uid);
