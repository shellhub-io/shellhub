import { Module } from "vuex";
import * as apiDevice from "../api/devices";
import * as apiContainer from "../api/container";
import { IDevice } from "@/interfaces/IDevice";
import { State } from "..";

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
        const [deviceRes, containerRes] = await Promise.all([
          apiDevice.fetchDevices(1, 10, "", "pending", undefined, ""),
          apiContainer.fetchContainers(1, 10, "", "pending", undefined, ""),
        ]);

        const combinedData = [
          ...deviceRes.data,
          ...containerRes.data,
        ];

        const combinedCount = parseInt(deviceRes.headers["x-total-count"], 10)
                              + parseInt(containerRes.headers["x-total-count"], 10);

        context.commit("setNotifications", { data: combinedData, headers: { "x-total-count": combinedCount } });
      } catch (error) {
        context.commit("clearListNotifications");
        throw error;
      }
    },
  },
};
