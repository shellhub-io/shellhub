import { createPinia, setActivePinia } from "pinia";
import { mount } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it } from "vitest";
import NotificationsList from "@/components/AppBar/Notifications/NotificationsList.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { router } from "@/router";
import { INotification } from "@/interfaces/INotification";

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

describe("Notifications List", () => {
  const vuetify = createVuetify();
  setActivePinia(createPinia());

  const wrapper = mount(NotificationsList, {
    global: {
      plugins: [router, vuetify, SnackbarPlugin],
      stubs: {
        DeviceActionButton: {
          name: "DeviceActionButton",
          template: "<div data-test='device-action-button-stub'></div>",
          props: ["uid", "name", "variant", "isInNotification", "show", "action"]
        },
        RouterLink: {
          template: "<a :href=\"`/devices/${to.params.identifier}`\"><slot /></a>",
          props: ["to"]
        }
      },
    },
    props: {
      notifications: mockNotifications as INotification[],
    },
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
