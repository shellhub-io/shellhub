import { createVuetify } from "vuetify";
import { DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { nextTick } from "vue";
import DeviceAcceptWarning from "@/components/Devices/DeviceAcceptWarning.vue";
import { namespacesApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

type DeviceAcceptWarningWrapper = VueWrapper<InstanceType<typeof DeviceAcceptWarning>>;

const members = [
  {
    id: "xxxxxxxx",
    username: "test",
    role: "owner",
  },
];

const namespaceData = {
  name: "test",
  owner: "xxxxxxxx",
  tenant_id: "fake-tenant-data",
  members,
  max_devices: 3,
  devices_count: 3,
  devices: 2,
  created_at: "",
};

const authData = {
  status: "",
  token: "",
  user: "test",
  name: "test",
  tenant: "fake-tenant-data",
  email: "test@test.com",
  id: "xxxxxxxx",
  role: "owner",
};

describe("Device Accept Warning", () => {
  let wrapper: DeviceAcceptWarningWrapper;
  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  beforeEach(() => {
    const el = document.createElement("div");
    document.body.appendChild(el);

    vi.useFakeTimers();
    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    envVariables.isCloud = true;

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.dispatch("users/setDeviceDuplicationOnAcceptance", true);

    wrapper = mount(DeviceAcceptWarning, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
      sync: false,
      attachTo: el,
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", () => {
    const wrapper = new DOMWrapper(document.body);

    expect(wrapper.find('[data-test="device-accept-warning-dialog"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="card-dialog"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="card-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="card-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="close-btn"]').exists()).toBe(true);
  });

  it("Closes the template", async () => {
    const wrapper = new DOMWrapper(document.body);
    const spyStore = vi.spyOn(store, "dispatch");

    await wrapper.find('[data-test="close-btn"]').trigger("click");
    expect(spyStore).toHaveBeenCalledWith("users/setDeviceDuplicationOnAcceptance", false);
  });

  it("Reads the conflicted device name", async () => {
    store.dispatch("devices/setDeviceToBeRenamed", "device");
    const dialogWrapper = new DOMWrapper(document.body);
    wrapper.vm.showMessage = true;

    await nextTick(); // Wait for the next tick to ensure the prop is updated

    const actualText = dialogWrapper.find('[data-test="card-text"]').text();
    const expectedText = "device name is already taken by another accepted device, please choose another name.";
    await nextTick();
    expect(actualText).toEqual(expectedText);
  });
});
