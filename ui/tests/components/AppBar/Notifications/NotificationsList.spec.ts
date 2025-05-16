import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach, afterEach } from "vitest";
import NotificationsList from "@/components/AppBar/Notifications/NotificationsList.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { router } from "@/router";
import { INotification } from "@/interfaces/INotification";
import { key, store } from "@/store";

const mockNotifications = [
  {
    id: "a582b47a42d",
    type: "device",
    data: {
      uid: "a582b47a42d",
      name: "39-5e-2a",
    },
  },
  {
    id: "a582b47a42e",
    type: "container",
    data: {
      uid: "a582b47a42e",
      name: "39-5e-2b",
    },
  },
];

describe("Notifications List", async () => {
  let wrapper: VueWrapper<InstanceType<typeof NotificationsList>>;
  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(NotificationsList, {
      global: {
        plugins: [[store, key], router, vuetify, SnackbarPlugin],
      },
      props: {
        notifications: mockNotifications as INotification[],
      },
    });
  });

  afterEach(() => { wrapper.unmount(); });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the correct number of list items", () => {
    const items = wrapper.findAllComponents({ name: "VListItem" });
    expect(items.length).toBe(mockNotifications.length);
  });

  it("Renders DeviceActionButton with correct props", () => {
    mockNotifications.forEach((notification) => {
      const button = wrapper.find(`[data-test="${notification.data.uid}-btn"]`);
      expect(button.exists()).toBe(true);
    });
  });

  it("Each DeviceActionButton gets correct variant prop", () => {
    const buttons = wrapper.findAllComponents({ name: "DeviceActionButton" });
    expect(buttons).toHaveLength(mockNotifications.length);

    buttons.forEach((button, index) => {
      expect(button.props("variant")).toBe(mockNotifications[index].type);
    });
  });

  it("Renders router-link with correct navigation target", () => {
    mockNotifications.forEach((notification) => {
      const link = wrapper.find(`[data-test="${notification.data.uid}-title"]`);
      expect(link.exists()).toBe(true);
      expect(link.attributes("href")).toContain(`/devices/${notification.data.uid}`);
    });
  });
});
