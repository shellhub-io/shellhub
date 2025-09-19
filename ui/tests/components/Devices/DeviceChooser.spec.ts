import { setActivePinia, createPinia } from "pinia";
import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { nextTick } from "vue";
import DeviceChooser from "@/components/Devices/DeviceChooser.vue";
import { router } from "@/router";
import { billingApi, devicesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";
import useDevicesStore from "@/store/modules/devices";

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

describe("Device Chooser", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceChooser>>;
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const devicesStore = useDevicesStore();
  const vuetify = createVuetify();

  const mockBillingApi = new MockAdapter(billingApi.getAxios());
  const mockDevicesApi = new MockAdapter(devicesApi.getAxios());

  beforeEach(async () => {
    localStorage.setItem("tenant", "fake-tenant-data");

    mockBillingApi.onGet("http://localhost:3000/api/billing/devices-most-used").reply(200, devices);
    mockDevicesApi.onGet("http://localhost:3000/api/devices?page=1&per_page=5&status=accepted").reply(200, devices);
    mockDevicesApi.onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted").reply(200, devices);

    authStore.role = "owner";
    devicesStore.showDeviceChooser = true;

    wrapper = mount(DeviceChooser, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component with FormDialog", async () => {
    await flushPromises();

    expect(wrapper.findComponent({ name: "FormDialog" }).exists()).toBe(true);

    const formDialog = wrapper.findComponent({ name: "FormDialog" });
    expect(formDialog.props("title")).toBe("Update account or select three devices");
    expect(formDialog.props("icon")).toBe("mdi-devices");
    expect(formDialog.props("confirmText")).toBe("Accept");
    expect(formDialog.props("cancelText")).toBe("Close");
    expect(formDialog.props("threshold")).toBe("md");
  });

  it("Render V-Tabs", async () => {
    await flushPromises();
    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="Suggested-tab"]').exists()).toBe(true);
    expect(dialog.find('[data-test="All-tab"]').exists()).toBe(true);
  });

  it("Accepts the devices listed (Suggested Devices)", async () => {
    mockBillingApi.onGet("http://localhost:3000/api/billing/device-most-used").reply(200);
    mockBillingApi.onPost("http://localhost:3000/api/billing/device-choice").reply(200, { devices });

    const storeSpy = vi.spyOn(devicesStore, "sendDeviceChoices");

    await wrapper.findComponent('[data-test="Suggested-tab"]').trigger("click");
    await nextTick();

    const formDialog = wrapper.findComponent({ name: "FormDialog" });
    await formDialog.vm.$emit("confirm");

    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith([
      {
        identity: {
          mac: "00:00:00:00:00:00",
        },
        info: {
          id: "linuxmint",
          pretty_name: "Linux Mint 19.3",
          version: "",
        },
        last_seen: "2020-05-20T18:58:53.276Z",
        name: "39-5e-2a",
        namespace: "user",
        online: false,
        public_key: "----- PUBLIC KEY -----",
        status: "accepted",
        tenant_id: "fake-tenant-data",
        uid: "a582b47a42d",
      },
      {
        identity: {
          mac: "00:00:00:00:00:00",
        },
        info: {
          id: "linuxmint",
          pretty_name: "Linux Mint 19.3",
          version: "",
        },
        last_seen: "2020-05-20T19:58:53.276Z",
        name: "39-5e-2b",
        namespace: "user",
        online: true,
        public_key: "----- PUBLIC KEY -----",
        status: "accepted",
        tenant_id: "fake-tenant-data",
        uid: "a582b47a42e",
      },
    ]);
  });

  it("Accepts the devices listed(All Devices)", async () => {
    mockBillingApi.onGet("http://localhost:3000/api/billing/device-most-used").reply(200);
    mockBillingApi.onPost("http://localhost:3000/api/billing/device-choice").reply(200, { devices: [devices] });
    mockDevicesApi.onGet("http://localhost:3000/api/devices?page=1&per_page=5&status=accepted").reply(200, devices);

    const storeSpy = vi.spyOn(devicesStore, "fetchDeviceList");

    await wrapper.findComponent('[data-test="All-tab"]').trigger("click");
    await nextTick();

    const formDialog = wrapper.findComponent({ name: "FormDialog" });
    await formDialog.vm.$emit("confirm");

    await flushPromises();

    expect(storeSpy).toHaveBeenCalled();
  });
});
