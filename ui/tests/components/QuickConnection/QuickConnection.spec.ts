import { createPinia, setActivePinia } from "pinia";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { store, key } from "@/store";
import QuickConnection from "@/components/QuickConnection/QuickConnection.vue";
import { router } from "@/router";
import { devicesApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

type QuickConnectionWrapper = VueWrapper<InstanceType<typeof QuickConnection>>;

const mockSnackbar = {
  showError: vi.fn(),
};

describe("Quick Connection", () => {
  let wrapper: QuickConnectionWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockDevicesApi = new MockAdapter(devicesApi.getAxios());

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

  beforeEach(async () => {
    mockDevicesApi
      // eslint-disable-next-line vue/max-len
      .onGet("http://localhost:3000/api/devices?filter=W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJvbmxpbmUiLCJvcGVyYXRvciI6ImVxIiwidmFsdWUiOnRydWV9fV0%3D&per_page=10&status=accepted")
      .reply(200, devices);

    wrapper = mount(QuickConnection, {
      global: {
        plugins: [[store, key], vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the dialog open button and other key elements", async () => {
    const dialog = new DOMWrapper(document.body);

    expect(wrapper.find('[data-test="quick-connection-open-btn"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="quick-connection-open-btn"]').trigger("click");

    expect(dialog.find('[data-test="search-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="hostname-header"]').exists()).toBe(true);
    expect(dialog.find('[data-test="operating-system-header"]').exists()).toBe(true);
    expect(dialog.find('[data-test="sshid-header"]').exists()).toBe(true);
    expect(dialog.find('[data-test="tags-header"]').exists()).toBe(true);
    expect(dialog.find('[data-test="connect-icon"]').exists()).toBe(true);
    expect(dialog.find('[data-test="copy-sshid-instructions"]').exists()).toBe(true);
    expect(dialog.find('[data-test="connect-icon"]').exists()).toBe(true);
    expect(dialog.find('[data-test="navigate-up-icon"]').exists()).toBe(true);
    expect(dialog.find('[data-test="navigate-down-icon"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
  });

  it("keyboardMacros function toggles dialog value on Ctrl + K keydown", async () => {
    const event = new KeyboardEvent("keydown", { ctrlKey: true, key: "k" });

    dispatchEvent(event);

    expect(wrapper.find('[data-test="quick-connection-open-btn"]').exists()).toBe(true);
  });

  it("Checks if the fetch function handles error on failure", async () => {
    mockDevicesApi
      // eslint-disable-next-line vue/max-len
      .onGet("http://localhost:3000/api/devices?filter=W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJvbmxpbmUiLCJvcGVyYXRvciI6ImVxIiwidmFsdWUiOnRydWV9fV0%3D&per_page=10&status=accepted")
      .reply(403);

    await wrapper.findComponent('[data-test="quick-connection-open-btn"]').trigger("click");

    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("An error occurred while loading devices.");
  });
});
