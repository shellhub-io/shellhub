import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import { describe, expect, it } from "vitest";
import { NotificationType } from "@/interfaces/INotification";
import { containersApi, devicesApi } from "@/api/http";
import useNotificationsStore from "@/store/modules/notifications";

const mockDevice = {
  uid: "device1",
  name: "Device 1",
};

const mockContainer = {
  uid: "container1",
  name: "Container 1",
};

describe("Notifications Store", () => {
  const mockDevicesApi = new MockAdapter(devicesApi.getAxios());
  const mockContainersApi = new MockAdapter(containersApi.getAxios());
  setActivePinia(createPinia());
  const notificationsStore = useNotificationsStore();

  it("should have initial state values", () => {
    expect(notificationsStore.notifications).toEqual([]);
    expect(notificationsStore.notificationCount).toEqual(0);
  });

  it("Successfully fetches and combines notifications from API", async () => {
    mockDevicesApi.onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=pending").reply(200, [mockDevice]);
    mockContainersApi.onGet("http://localhost:3000/api/containers?page=1&per_page=10&status=pending").reply(200, [mockContainer]);

    await notificationsStore.fetchNotifications();

    const { notifications } = notificationsStore;
    expect(notifications).toHaveLength(2);
    expect(notifications).toContainEqual({
      id: mockDevice.uid,
      type: NotificationType.DEVICE,
      data: mockDevice,
    });
    expect(notifications).toContainEqual({
      id: mockContainer.uid,
      type: NotificationType.CONTAINER,
      data: mockContainer,
    });
  });

  it("Handles API error when fetching notifications", async () => {
    mockDevicesApi.onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=pending").reply(500);

    await expect(notificationsStore.fetchNotifications()).rejects.toThrow("Request failed with status code 500");

    expect(notificationsStore.notifications).toEqual([]);
    expect(notificationsStore.notificationCount).toEqual(0);
  });
});
