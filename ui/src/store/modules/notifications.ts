import { Module } from "vuex";
import * as apiDevice from "../api/devices";
import * as apiContainer from "../api/container";
import { IDevice } from "@/interfaces/IDevice";
import { State } from "..";
import { IContainer } from "@/interfaces/IContainer";

export interface NotificationsState {
  notifications: Array<IDevice | IContainer>;
  total: number;
}

export const notifications: Module<NotificationsState, State> = {
  namespaced: true,
  state: {
    notifications: [],
    total: 0,
  },

  getters: {
    notifications: (state) => state.notifications,
    total: (state) => state.total,
  },

  mutations: {
    setNotifications: (state, res) => {
      state.notifications = res.data;
      state.total = parseInt(res.headers["x-total-count"], 10);
    },

    clearNotifications: (state) => {
      state.notifications = [];
      state.total = 0;
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

        const combinedCount = parseInt(deviceRes.headers["x-total-count"], 10) + parseInt(containerRes.headers["x-total-count"], 10);

        context.commit("setNotifications", { data: combinedData, headers: { "x-total-count": combinedCount } });
      } catch (error) {
        context.commit("clearNotifications");
        throw error;
      }
    },
  },
};
