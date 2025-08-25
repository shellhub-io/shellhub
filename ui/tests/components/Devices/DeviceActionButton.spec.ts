import { setActivePinia, createPinia } from "pinia";
import { DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import DeviceActionButton from "@/components/Devices/DeviceActionButton.vue";
import { router } from "@/router";
import { devicesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

const devices = [
  {
    uid: "a582b47a42d",
    name: "39-5e-2a",
    identity: {
      mac: "00:00:00:00:00:00",
    },
    info: {
      id: "linuxmint",
      pretty_name: "Linux Mint 19.3",
      version: "",
    },
    public_key: "----- PUBLIC KEY -----",
    tenant_id: "fake-tenant-data",
    last_seen: "2020-05-20T18:58:53.276Z",
    online: false,
    namespace: "user",
    status: "accepted",
  },
  {
    uid: "a582b47a42e",
    name: "39-5e-2b",
    identity: {
      mac: "00:00:00:00:00:00",
    },
    info: {
      id: "linuxmint",
      pretty_name: "Linux Mint 19.3",
      version: "",
    },
    public_key: "----- PUBLIC KEY -----",
    tenant_id: "fake-tenant-data",
    last_seen: "2020-05-20T19:58:53.276Z",
    online: true,
    namespace: "user",
    status: "accepted",
  },
];

describe("Device Action Button", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceActionButton>>;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  let mockDevices: MockAdapter;

  beforeEach(async () => {
    mockDevices = new MockAdapter(devicesApi.getAxios());

    mockDevices.onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted").reply(200, devices);

    wrapper = mount(DeviceActionButton, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
      props: {
        uid: devices[0].uid,
        variant: "device",
        isInNotification: false,
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component data table", async () => {
    await wrapper.setProps({ name: "test-device", uid: "test-uid", isInNotification: true });
    const notificationButton = wrapper.find('[data-test="notification-action-button"]');
    expect(notificationButton.exists()).toBe(true);
    await notificationButton.trigger("click");
    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="device-action-dialog"]').exists()).toBe(true);
  });

  it("Clicking on notification button opens dialog", async () => {
    await wrapper.setProps({ isInNotification: true });
    const notificationButton = wrapper.find('[data-test="notification-action-button"]');
    await notificationButton.trigger("click");
    expect(wrapper.vm.showDialog).toBe(true);
  });

  it("Closing dialog sets dialog value to false", async () => {
    vi.spyOn(console, "warn").mockImplementation((message) => {
      if (message.includes("click:outside")) return;
      console.log(message);
    });
    wrapper.vm.showDialog = true;
    const dialogComponent = wrapper.findComponent({ name: "VDialog" });
    await dialogComponent.vm.$emit("click:outside");
    expect(wrapper.vm.showDialog).toBe(false);
  });

  it("Close button in dialog emits 'update' event with false", async () => {
    wrapper.vm.showDialog = true;
    await wrapper.setProps({ isInNotification: true });
    const closeButton = wrapper.findComponent('[data-test="close-btn"]');
    await closeButton.trigger("click");
    expect(wrapper.emitted("update")).toBeTruthy();
  });
});
