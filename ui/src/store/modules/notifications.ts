import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { fetchDevices } from "../api/devices";
import { fetchContainers } from "../api/containers";
import { deviceToNotification, containerToNotification } from "@/utils/notificationAdapters";
import { IDevice } from "@/interfaces/IDevice";
import { IContainer } from "@/interfaces/IContainer";
import { INotification } from "@/interfaces/INotification";

const useNotificationsStore = defineStore("notifications", () => {
  const notifications = ref<INotification[]>([]);
  const notificationCount = computed(() => notifications.value.length);

  const fetchNotifications = async () => {
    try {
      const [deviceRes, containerRes] = await Promise.all([
        fetchDevices(1, 10, "pending"),
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

      notifications.value = combinedNotifications;
    } catch (error) {
      notifications.value = [];
      throw error;
    }
  };

  return {
    notifications,
    notificationCount,

    fetchNotifications,
  };
});

export default useNotificationsStore;
