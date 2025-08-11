import { createPinia, setActivePinia } from "pinia";
import MockAdapter from "axios-mock-adapter";
import { afterEach, describe, expect, it } from "vitest";
import { store } from "@/store";
import { NotificationType } from "@/interfaces/INotification";
import { containersApi, devicesApi } from "@/api/http";

describe("Notifications Vuex Module", () => {
  const mockDevicesApi = new MockAdapter(devicesApi.getAxios());
  const mockContainersApi = new MockAdapter(containersApi.getAxios());
  setActivePinia(createPinia());
  const mockDevice = {
    uid: "device1",
    name: "Device 1",
  };

  const mockContainer = {
    uid: "container1",
    name: "Container 1",
  };

  const notifications = [
    { id: mockDevice.uid, type: NotificationType.DEVICE, data: mockDevice },
    { id: mockContainer.uid, type: NotificationType.CONTAINER, data: mockContainer },
  ];

  afterEach(() => {
    store.commit("notifications/clearNotifications");
  });

  it("Returns notifications with default variables", () => {
    expect(store.getters["notifications/notifications"]).toEqual([]);
    expect(store.getters["notifications/total"]).toEqual(0);
  });

  it("Commits setNotifications mutation", () => {
    store.commit("notifications/setNotifications", notifications);

    expect(store.getters["notifications/notifications"]).toEqual(notifications);
    expect(store.getters["notifications/total"]).toEqual(2);
  });

  it("Commits clearNotifications mutation", () => {
    store.commit("notifications/setNotifications", notifications);
    store.commit("notifications/clearNotifications");

    expect(store.getters["notifications/notifications"]).toEqual([]);
    expect(store.getters["notifications/total"]).toEqual(0);
  });

  it("Gets notifications by type", () => {
    const notifications = [
      { id: mockDevice.uid, type: NotificationType.DEVICE, data: mockDevice },
      { id: mockContainer.uid, type: NotificationType.CONTAINER, data: mockContainer },
      { id: "device2", type: NotificationType.DEVICE, data: { ...mockDevice, uid: "device2" } },
    ];

    store.commit("notifications/setNotifications", notifications);

    const deviceNotifications = store.getters["notifications/notificationsByType"](NotificationType.DEVICE);
    expect(deviceNotifications).toHaveLength(2);
    expect(deviceNotifications.every((n) => n.type === NotificationType.DEVICE)).toBe(true);

    const containerNotifications = store.getters["notifications/notificationsByType"](NotificationType.CONTAINER);
    expect(containerNotifications).toHaveLength(1);
    expect(containerNotifications.every((n) => n.type === NotificationType.CONTAINER)).toBe(true);
  });

  it("Successfully fetches and combines notifications from API", async () => {
    mockDevicesApi.onGet("http://localhost:3000/api/devices?filter=&page=1&per_page=10&status=pending").reply(200, [mockDevice]);
    mockContainersApi.onGet("http://localhost:3000/api/containers?page=1&per_page=10&status=pending").reply(200, [mockContainer]);

    await store.dispatch("notifications/fetch");

    const notifications = store.getters["notifications/notifications"];
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
    mockDevicesApi.onGet("http://localhost:3000/api/devices?filter=&page=1&per_page=10&status=pending").reply(500);

    await expect(store.dispatch("notifications/fetch")).rejects.toThrow("Request failed with status code 500");

    expect(store.getters["notifications/notifications"]).toEqual([]);
    expect(store.getters["notifications/total"]).toEqual(0);
  });
});
