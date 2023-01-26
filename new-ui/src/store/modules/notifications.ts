import { Module } from "vuex";
import { State } from "./../index";
import * as apiDevice from "../api/devices";
import { IDevice } from "@/interfaces/IDevice";

export interface NotificationsState {
  notifications: Array<IDevice>;
  numberNotifications: number;
}

export const notifications: Module<NotificationsState, State> = {
  namespaced: true,
  state: {
    notifications: [],
    numberNotifications: 0,
  },

  getters: {
    list: (state) => state.notifications,
    getNumberNotifications: (state) => state.numberNotifications,
  },

  mutations: {
    setNotifications: (state, res) => {
      state.notifications = res.data;
      state.numberNotifications = parseInt(res.headers["x-total-count"], 10);
    },

    clearListNotifications: (state) => {
      state.notifications = [];
      state.numberNotifications = 0;
    },
  },

  actions: {
    fetch: async (context) => {
      try {
        const res = await apiDevice.fetchDevices(10, 1, "", "pending", "", "");
        context.commit("setNotifications", res);
      } catch (error) {
        context.commit("clearListNotifications");
        throw error;
      }
    },
  },
};
