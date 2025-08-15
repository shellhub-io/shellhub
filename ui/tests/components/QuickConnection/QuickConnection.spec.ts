import { createPinia, setActivePinia } from "pinia";
import { DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import QuickConnection from "@/components/QuickConnection/QuickConnection.vue";
import { router } from "@/router";
import { devicesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type QuickConnectionWrapper = VueWrapper<InstanceType<typeof QuickConnection>>;

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
      .onGet("http://localhost:3000/api/devices?filter=W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJvbmxpbmUiLCJvcGVyYXRvciI6ImVxIiwidmFsdWUiOnRydWV9fSx7InR5cGUiOiJwcm9wZXJ0eSIsInBhcmFtcyI6eyJuYW1lIjoibmFtZSIsIm9wZXJhdG9yIjoiY29udGFpbnMiLCJ2YWx1ZSI6IiJ9fSx7InR5cGUiOiJvcGVyYXRvciIsInBhcmFtcyI6eyJuYW1lIjoiYW5kIn19XQ%3D%3D&page=1&per_page=10&status=accepted")
      .reply(200, devices);

    wrapper = mount(QuickConnection, {
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
});
