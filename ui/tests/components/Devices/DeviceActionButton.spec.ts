import { setActivePinia, createPinia } from "pinia";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
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

  beforeEach(() => {
    mockDevices = new MockAdapter(devicesApi.getAxios());

    mockDevices.onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted").reply(200, devices);

    wrapper = mount(DeviceActionButton, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
      props: {
        uid: devices[0].uid,
        variant: "device",
      },
    });
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Closing dialog sets showDialog value to false", async () => {
    wrapper.vm.showDialog = true;
    const dialogComponent = wrapper.findComponent({ name: "BaseDialog" });
    await dialogComponent.vm.$emit("close");
    expect(wrapper.vm.showDialog).toBe(false);
  });
});
