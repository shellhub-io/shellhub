import MockAdapter from "axios-mock-adapter";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach, afterEach, vi } from "vitest";
import { key, store } from "@/store";
import NotificationsMenu from "@/components/AppBar/Notifications/NotificationsMenu.vue";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import { containersApi, devicesApi } from "@/api/http";
import { router } from "@/router";

const deviceData = [{
  uid: "a582b47a42d",
  name: "39-5e-2a",
}];

const containerData = [{
  uid: "a582b47a42e",
  name: "39-5e-2b",
}];

const mockSnackbar = {
  showError: vi.fn(),
};

const authData = {
  status: "success",
  token: "",
  user: "test",
  name: "test",
  tenant: "fake-tenant-data",
  email: "test@test.com",
  id: "xxxxxxxx",
  role: "owner",
  mfa: {
    enable: false,
    validate: false,
  },
};

let mockDevices: MockAdapter;
let mockContainers: MockAdapter;

describe("Notifications Menu", async () => {
  let wrapper: VueWrapper<InstanceType<typeof NotificationsMenu>>;
  const vuetify = createVuetify();

  mockDevices = new MockAdapter(devicesApi.getAxios());
  mockContainers = new MockAdapter(containersApi.getAxios());

  const mockPendingNotifications = (deviceData, containerData, status = 200) => {
    mockDevices.onGet("http://localhost:3000/api/devices?filter=&page=1&per_page=10&status=pending").reply(status, deviceData);
    mockContainers.onGet("http://localhost:3000/api/containers?filter=&page=1&per_page=10&status=pending").reply(status, containerData);
  };

  beforeEach(async () => {
    store.commit("auth/authSuccess", authData);
    mockPendingNotifications(deviceData, containerData);

    wrapper = mount(NotificationsMenu, {
      global: {
        plugins: [[store, key], router, vuetify],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
    });
  });

  afterEach(() => { wrapper.unmount(); });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Calls fetch on before mount", () => {
    const fetchSpy = vi.spyOn(store, "dispatch");
    wrapper = mount(NotificationsMenu, {
      global: {
        plugins: [[store, key], router, vuetify],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
    });
    expect(fetchSpy).toHaveBeenCalledWith("notifications/fetch");
  });

  it("Shows correct number of notifications in badge", () => {
    const badge = wrapper.find("[data-test='notifications-badge']");
    expect(badge.exists()).toBe(true);
    expect(badge.text()).toBe("2");
  });

  it("Shows only icon when showNotifications = false", async () => {
    mockPendingNotifications([], [], 200);

    await wrapper.vm.fetchNotifications();

    const badge = wrapper.find("[data-test='notifications-badge']");
    expect(badge.find("span").element.style.display).toBe("none");
  });

  it("Shows snackbar error when fetch fails", async () => {
    mockPendingNotifications([], [], 500);

    await wrapper.vm.fetchNotifications();
    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load notifications.");
  });

  it("Shows empty message when user has permission but no notifications", async () => {
    mockPendingNotifications([], [], 200);

    await wrapper.vm.fetchNotifications();

    const icon = wrapper.find("[data-test='notifications-badge'] i");
    await icon.trigger("click");

    const card = wrapper.findComponent({ name: "VCard" });

    expect(card.exists()).toBe(true);
    expect(card.text()).toContain("You don't have notifications");
  });

  it("Shows permission error message when user lacks permission", async () => {
    store.commit("auth/authSuccess", { ...authData, role: "observer" });

    const icon = wrapper.find("[data-test='notifications-badge'] i");
    await icon.trigger("click");

    const card = wrapper.findComponent({ name: "VCard" });

    expect(card.exists()).toBe(true);
    expect(card.text()).toContain("You don't have permission to view notifications");
  });

  it("Shows notification card and pending devices button", async () => {
    const icon = wrapper.find("[data-test='notifications-badge'] i");
    await icon.trigger("click");

    const card = wrapper.findComponent({ name: "VCard" });
    const showPendingDevicesBtn = card.find("[data-test='pending-devices-btn']");

    expect(card.exists()).toBe(true);
    expect(showPendingDevicesBtn.exists()).toBe(true);
  });
});
