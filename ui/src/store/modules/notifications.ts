import { Module } from "vuex";
import { fetchDevices } from "../api/devices";
import { fetchContainers } from "../api/containers";
import { State } from "..";
import { deviceToNotification, containerToNotification } from "@/utils/notificationAdapters";
import { IDevice } from "@/interfaces/IDevice";
import { IContainer } from "@/interfaces/IContainer";
import { INotification } from "@/interfaces/INotification";

export interface NotificationsState {
  notifications: INotification[];
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
    notificationsByType: (state) => (type: string) => state.notifications.filter((notification) => notification.type === type),
  },

  mutations: {
    setNotifications: (state, notifications) => {
      state.notifications = notifications;
      state.total = notifications.length;
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
          fetchDevices(1, 10, "", "pending", undefined, ""),
          fetchContainers(1, 10, "pending"),
        ]);

        const devices = deviceRes.data as IDevice[];
        const containers = containerRes.data as IContainer[];

        const deviceNotifications = devices.map(deviceToNotification);
        const containerNotifications = containers.map(containerToNotification);

        const combinedNotifications = [
          ...deviceNotifications,
          ...containerNotifications,
        ];

        context.commit("setNotifications", combinedNotifications);
      } catch (error) {
        context.commit("clearNotifications");
        throw error;
      }
    },
  },
};
