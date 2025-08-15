import {
  IDevice,
  IUpdateDeviceTags,
  IDeviceRename,
  IDevicePostTag,
  FetchDevicesParams,
  IDeviceMethods,
} from "./IDevice";

// Container is essentially the same as Device
export type IContainer = IDevice;

export type IUpdateContainerTags = IUpdateDeviceTags;
export type IContainerRename = IDeviceRename;
export type IContainerPostTag = IDevicePostTag;

export type FetchContainerParams = FetchDevicesParams;

export interface IContainerMethods extends IDeviceMethods {
  fetchDevices: (params: FetchContainerParams) => Promise<void>; // Keep original method name for compatibility with DeviceTable component
  getList: () => IContainer[];
}
