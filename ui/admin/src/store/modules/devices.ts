import { Module } from "vuex";
import { State } from "./../index";
import * as apiDevice from "../api/devices";
import { IDevice } from "./../../interfaces/IDevice";

export interface DevicesState {
  devices: Array<IDevice>;
  device: IDevice;
  numberDevices: number;
  page: number;
  perPage: number;
  filter: string;
  sortStatusField: string;
  sortStatusString: "asc" | "desc" | undefined;
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
    sortStatusField: "",
    sortStatusString: undefined,
  },

  getters: {
    list: (state) => state.devices,
    get: (state) => state.device,
    page: (state) => state.page,
    perPage: (state) => state.perPage,
    filter: (state) => state.filter,
    numberDevices: (state) => state.numberDevices,
    sortStatusField: (state) => state.sortStatusField,
    sortStatusString: (state) => state.sortStatusString,
  },

  mutations: {
    setDevices: (state, res) => {
      state.devices = res.data;
      state.numberDevices = parseInt(res.headers["x-total-count"], 10);
    },

    setDevice: (state, res) => {
      state.device = res.data;
    },

    setPagePerpageFilter: (state, data) => {
      state.page = data.page;
      state.perPage = data.perPage;
      state.filter = data.filter;
      state.sortStatusField = data.sortStatusField;
      state.sortStatusString = data.sortStatusString;
    },

    setFilterDevices: (state, filter) => {
      state.filter = filter;
    },

    setSortStatusDevice: (state, data) => {
      state.sortStatusField = data.sortStatusField;
      state.sortStatusString = data.sortStatusString;
    },

    clearListDevices: (state) => {
      state.devices = [];
      state.numberDevices = 0;
    },
  },

  actions: {
    async fetch({ commit }, data) {
      const res = await apiDevice.fetchDevices(
        data.page,
        data.perPage,
        data.filter,
        data.sortStatusField,
        data.sortStatusString,
      );
      if (res.data.length) {
        commit("setDevices", res);
        commit("setPagePerpageFilter", data);
        return res;
      }

      return false;
    },

    async setFilter({ commit }, filter) {
      commit("setFilterDevices", filter);
    },

    async setSortStatus({ commit }, data) {
      commit("setSortStatusDevice", data);
    },

    async refresh({ commit, state }) {
      try {
        const res = await apiDevice.fetchDevices(
          state.page,
          state.perPage,
          state.filter,
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
          state.sortStatusField,
          state.sortStatusString,
        );
        commit("setDevices", res);
        commit("setFilterDevices", data.filter);
      } catch (error) {
        commit("clearListDevices");
        throw error;
      }
    },

    async get({ commit }, uid) {
      const res = await apiDevice.getDevice(uid);
      commit("setDevice", res);
      return res;
    },
  },
};
