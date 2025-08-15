import { createPinia, setActivePinia } from "pinia";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi, afterEach } from "vitest";
import { store, key } from "@/store";
import QuickConnectionList from "@/components/QuickConnection/QuickConnectionList.vue";
import { router } from "@/router";
import { devicesApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

type QuickConnectionListWrapper = VueWrapper<InstanceType<typeof QuickConnectionList>>;

const mockSnackbar = {
  showError: vi.fn(),
};

const devices = [
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
    tags: ["test-tag"],
  },
];

// eslint-disable-next-line vue/max-len
const mockDeviceApiUrl = "http://localhost:3000/api/devices?filter=W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJvbmxpbmUiLCJvcGVyYXRvciI6ImVxIiwidmFsdWUiOnRydWV9fSx7InR5cGUiOiJwcm9wZXJ0eSIsInBhcmFtcyI6eyJuYW1lIjoibmFtZSIsIm9wZXJhdG9yIjoiY29udGFpbnMifX0seyJ0eXBlIjoib3BlcmF0b3IiLCJwYXJhbXMiOnsibmFtZSI6ImFuZCJ9fV0%3D&page=1&per_page=10&status=accepted";

describe("Quick Connection List", () => {
  let wrapper: QuickConnectionListWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockDevicesApi = new MockAdapter(devicesApi.getAxios());

  beforeEach(async () => {
    mockDevicesApi.onGet(mockDeviceApiUrl).reply(200, devices);

    wrapper = mount(QuickConnectionList, {
      global: {
        plugins: [[store, key], vuetify, router],
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

  it("Renders the devices list", () => {
    expect(wrapper.find('[data-test="devices-list"]').exists()).toBe(true);
  });

  it("Renders each device card", () => {
    expect(wrapper.find('[data-test="device-list-item"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-name"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-info"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-ssh-id"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-tags"]').exists()).toBe(true);
  });

  it("Renders the copy ID button", () => {
    expect(wrapper.find('[data-test="copy-id-button"]').exists()).toBe(true);
  });

  it("Renders the tag chips", () => {
    expect(wrapper.find('[data-test="tag-chip"]').exists()).toBe(true);
  });

  it("Renders the no tags chip", async () => {
    // Change the value of tags[0] to an empty string for the first device
    devices[0].tags[0] = "";
    await flushPromises();
    expect(wrapper.find('[data-test="no-tags-chip"]').exists()).toBe(true);
  });

  it("Renders the no online devices message", async () => {
    mockDevicesApi.onGet(mockDeviceApiUrl).reply(200, []);
    await flushPromises();
    expect(wrapper.find('[data-test="no-online-devices"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="no-online-devices-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="no-online-devices-message"]').exists()).toBe(true);
  });

  it("Checks if the fetch function handles error on failure", async () => {
    mockDevicesApi.onGet(mockDeviceApiUrl).reply(403);
    await flushPromises();
    expect(mockSnackbar.showError).toHaveBeenCalledWith("An error occurred while loading devices.");
  });
});
