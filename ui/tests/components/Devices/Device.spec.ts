import { setActivePinia, createPinia } from "pinia";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import Device from "@/components/Devices/Device.vue";
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

describe("Device", () => {
  let wrapper: VueWrapper<InstanceType<typeof Device>>;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  let mockDevices: MockAdapter;

  beforeEach(async () => {
    mockDevices = new MockAdapter(devicesApi.getAxios());

    mockDevices.onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted").reply(200, devices);

    wrapper = mount(Device, {
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

  it("Renders correctly", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Contains the correct tabs", () => {
    const tabs = wrapper.findAllComponents({ name: "VBtn" });
    expect(tabs).toHaveLength(3); // Three tabs expected
    expect(tabs[0].text()).toBe("Accepted");
    expect(tabs[1].text()).toBe("Pending");
    expect(tabs[2].text()).toBe("Rejected");
  });
});
