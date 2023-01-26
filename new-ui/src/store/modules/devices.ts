import { Module } from "vuex";
import { State } from "./../index";

import * as apiDevice from "../api/devices";
import * as apiBilling from "../api/billing";
import { IDevice } from "@/interfaces/IDevice";

export interface DevicesState {
  devices: Array<IDevice>;
  device: IDevice;
  numberDevices: number;
  page: number;
  perPage: number;
  filter: null | string;
  status: string;
  sortStatusField: null | string;
  sortStatusString: string;
  deviceChooserStatus: boolean;
  devicesForUserToChoose: Array<IDevice>;
  numberdevicesForUserToChoose: number;
  devicesSelected: Array<IDevice>;
}

export const devices: Module<DevicesState, State> = {
  namespaced: true,
  state: {
    devices: [],
    device: {} as IDevice,
    numberDevices: 0,
    page: 1,
    perPage: 10,
    filter: "",
    status: "",
    sortStatusField: null,
    sortStatusString: "",
    deviceChooserStatus: false,
    devicesForUserToChoose: [],
    numberdevicesForUserToChoose: 0,
    devicesSelected: [],
  },

  getters: {
    list: (state) => state.devices,
    get: (state) => state.device,
    getNumberDevices: (state) => state.numberDevices,
    getPage: (state) => state.page,
    getPerPage: (state) => state.perPage,
    getFilter: (state) => state.filter,
    getStatus: (state) => state.status,
    getFirstPending: (state) => state.device,
    getDeviceChooserStatus: (state) => state.deviceChooserStatus,
    getDevicesForUserToChoose: (state) => state.devicesForUserToChoose,
    getNumberForUserToChoose: (state) => state.numberdevicesForUserToChoose,
    getDevicesSelected: (state) => state.devicesSelected,
  },

  mutations: {
    setDevices: (state, res) => {
      state.devices = res.data;
      state.numberDevices = parseInt(res.headers["x-total-count"], 10);
    },

    removeDevice: (state, uid) => {
      state.devices.splice(
        state.devices.findIndex((d) => d.uid === uid),
        1,
      );
    },

    renameDevice: (state, data) => {
      const { device } = state;
      device.name = data.name;
      state.device = device;
    },

    setDevice: (state, data) => {
      state.device = data;
    },

    setPagePerpageFilter: (state, data) => {
      state.page = data.page;
      state.perPage = data.perPage;
      state.filter = data.filter;
      state.status = data.status;
      state.sortStatusField = data.sortStatusField;
      state.sortStatusString = data.sortStatusString;
    },

    setFilter: (state, filter) => {
      state.filter = filter;
    },

    setDeviceChooserStatus: (state, status) => {
      state.deviceChooserStatus = status;
    },

    setDevicesForUserToChoose: (state, res) => {
      state.devicesForUserToChoose = res.data;
      state.numberdevicesForUserToChoose = parseInt(
        res.headers["x-total-count"],
        10,
      );
    },

    setDevicesSelected: (state, data) => {
      state.devicesSelected = data;
    },

    setSortStatusDevice: (state, data) => {
      state.sortStatusField = data.sortStatusField;
      state.sortStatusString = data.sortStatusString;
    },

    clearListDevices: (state) => {
      state.devices = [];
      state.numberDevices = 0;
    },

    clearObjectDevice: (state) => {
      state.device = {} as IDevice;
    },

    clearListDevicesForUserToChoose: (state) => {
      state.devicesForUserToChoose = [];
      state.numberdevicesForUserToChoose = 0;
    },
  },

  actions: {
    fetch: async ({ commit }, data) => {
      const res = await apiDevice.fetchDevices(
        data.page,
        data.perPage,
        data.filter,
        data.status,
        data.sortStatusField,
        data.sortStatusString,
      );
      if (res.data.length) {
        commit("setDevices", res);
        commit("setPagePerpageFilter", data);
        return res;
      }

      commit("clearListDevices", res);
      return false;
    },

    remove: async (context, uid) => {
      await apiDevice.removeDevice(uid);
    },

    rename: async (context, data) => {
      await apiDevice.renameDevice(data);
      context.commit("renameDevice", data);
    },

    get: async (context, uid) => {
      try {
        const res = await apiDevice.getDevice(uid);
        context.commit("setDevice", res.data);
      } catch (error) {
        context.commit("clearObjectDevice");
        throw error;
      }
    },

    accept: async (context, uid) => {
      try {
        await apiDevice.acceptDevice(uid);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    reject: async (context, uid) => {
      try {
        await apiDevice.rejectDevice(uid);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    setFilter: async (context, filter) => {
      context.commit("setFilter", filter);
    },

    refresh: async ({ commit, state }) => {
      try {
        const res = await apiDevice.fetchDevices(
          state.page,
          state.perPage,
          state.filter,
          state.status,
          state.sortStatusField,
          state.sortStatusString,
        );
        commit("setDevices", res);
      } catch (error) {
        commit("clearListDevices");
        throw error;
      }
    },

    async search({ commit, state }, data) {
      try {
        const res = await apiDevice.fetchDevices(
          data.page,
          data.perPage,
          data.filter,
          state.status,
          state.sortStatusField,
          state.sortStatusString,
        );
        commit("setDevices", res);
        commit("setFilter", data.filter);
      } catch (error) {
        commit("clearListDevices");
        throw error;
      }
    },

    setFirstPending: async (context) => {
      try {
        const res = await apiDevice.fetchDevices(
          1,
          1,
          null,
          "pending",
          null,
          "",
        );
        context.commit("setDevice", res.data[0]);
      } catch (error) {
        context.commit("clearObjectDevice");
        throw error;
      }
    },

    setDeviceChooserStatus: async (context, status) => {
      context.commit("setDeviceChooserStatus", status);
    },

    setDevicesForUserToChoose: async (context, data) => {
      try {
        const res = await apiDevice.fetchDevices(
          data.page,
          data.perPage,
          data.filter,
          data.status,
          data.sortStatusField,
          data.sortStatusString,
        );
        if (res.data.length) {
          context.commit("setDevicesForUserToChoose", res);
          context.commit("setPagePerpageFilter", data);
          return res;
        }

        return false;
      } catch (error) {
        context.commit("clearListDevicesForUserToChoose");
        throw error;
      }
    },

    setDevicesSelected: (context, data) => {
      context.commit("setDevicesSelected", data);
    },

    async setSortStatus({ commit }, data) {
      commit("setSortStatusDevice", data);
    },

    postDevicesChooser: async (context, data) => {
      try {
        await apiBilling.postDevicesChooser(data);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    getDevicesMostUsed: async (context) => {
      try {
        const res = await apiBilling.getDevicesMostUsed();
        context.commit("setDevicesForUserToChoose", res);
      } catch (error) {
        context.commit("clearListDevicesForUserToChoose");
        throw error;
      }
    },

    resetListDevices: async (context) => {
      context.commit("clearListDevices");
    },

    updateDeviceTag: async (context, data) => {
      try {
        await apiDevice.updateDeviceTag(data);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },
  },
};
