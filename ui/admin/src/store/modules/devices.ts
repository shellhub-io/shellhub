import { defineStore } from "pinia";
import { IAdminDevice } from "@admin/interfaces/IDevice";
import * as apiDevice from "../api/devices";

export interface DevicesState {
  devices: Array<IAdminDevice>;
  device: IAdminDevice;
  numberDevices: number;
  page: number;
  perPage: number;
  filter: string;
  sortStatusField: string;
  sortStatusString: "asc" | "desc" | undefined;
}

export const useDevicesStore = defineStore("devices", {
  state: (): DevicesState => ({
    devices: [],
    device: {} as IAdminDevice,
    numberDevices: 0,
    page: 1,
    perPage: 10,
    filter: "",
    sortStatusField: "",
    sortStatusString: undefined,
  }),

  getters: {
    list: (state) => state.devices,
    getDevice: (state) => state.device,
    getPage: (state) => state.page,
    getPerPage: (state) => state.perPage,
    getFilter: (state) => state.filter,
    getNumberDevices: (state) => state.numberDevices,
    getSortStatusField: (state) => state.sortStatusField,
    getSortStatusString: (state) => state.sortStatusString,
  },

  actions: {
    async fetch(data: {
      page: number;
      perPage: number;
      filter: string;
      sortStatusField: string;
      sortStatusString: "asc" | "desc" | undefined;
    }) {
      const res = await apiDevice.fetchDevices(
        data.page,
        data.perPage,
        data.filter,
        data.sortStatusField,
        data.sortStatusString,
      );
      if (res.data.length) {
        this.devices = res.data as never;
        this.numberDevices = parseInt(res.headers["x-total-count"], 10);
        this.page = data.page;
        this.perPage = data.perPage;
        this.filter = data.filter;
        this.sortStatusField = data.sortStatusField;
        this.sortStatusString = data.sortStatusString;
        return res;
      }
      return false;
    },

    setFilter(filter: string) {
      this.filter = filter;
    },

    setSortStatus(data: { sortStatusField: string; sortStatusString: "asc" | "desc" | undefined }) {
      this.sortStatusField = data.sortStatusField;
      this.sortStatusString = data.sortStatusString;
    },

    async refresh() {
      try {
        const res = await apiDevice.fetchDevices(
          this.page,
          this.perPage,
          this.filter,
          this.sortStatusField,
          this.sortStatusString,
        );
        this.devices = res.data as never;
        this.numberDevices = parseInt(res.headers["x-total-count"], 10);
      } catch (error) {
        this.devices = [];
        this.numberDevices = 0;
        throw error;
      }
    },

    async search(data: { page: number; perPage: number; filter: string }) {
      try {
        const res = await apiDevice.fetchDevices(
          data.page,
          data.perPage,
          data.filter,
          this.sortStatusField,
          this.sortStatusString,
        );
        this.devices = res.data as never;
        this.numberDevices = parseInt(res.headers["x-total-count"], 10);
        this.filter = data.filter;
      } catch (error) {
        this.devices = [];
        this.numberDevices = 0;
        throw error;
      }
    },

    async get(uid: string) {
      const res = await apiDevice.getDevice(uid);
      this.device = res.data as never;
      return res;
    },
  },
});

export default useDevicesStore;
