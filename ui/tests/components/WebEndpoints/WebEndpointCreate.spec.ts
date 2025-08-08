import { setActivePinia, createPinia } from "pinia";
import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { store, key } from "@/store";
import WebEndpointCreate from "@/components/WebEndpoints/WebEndpointCreate.vue";
import { webEndpointsApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type WebEndpointCreateWrapper = VueWrapper<InstanceType<typeof WebEndpointCreate>>;

describe("WebEndpointCreate.vue", () => {
  let wrapper: WebEndpointCreateWrapper;
  const mockWebEndpointsApi = new MockAdapter(webEndpointsApi.getAxios());
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(() => {
    wrapper = mount(WebEndpointCreate, {
      attachTo: document.body,
      global: {
        plugins: [[store, key], vuetify, SnackbarPlugin],
      },
      props: {
        uid: "fake-uid",
        useDevicesList: false,
        modelValue: true,
      },
    });
  });

  it("is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("renders the dialog and basic fields", () => {
    const dialog = new DOMWrapper(document.body);

    expect(dialog.find('[data-test="tunnel-create-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="create-dialog-title"]').text()).toContain("Create Device Web Endpoint");
    expect(dialog.find('[data-test="host-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="port-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="timeout-combobox"]').exists()).toBe(true);
    expect(dialog.find('[data-test="custom-timeout"]').exists()).toBe(false);
    expect(dialog.find('[data-test="create-tunnel-btn"]').exists()).toBe(true);
  });

  it("successfully creates a Web Endpoint", async () => {
    mockWebEndpointsApi.onPost("http://localhost:3000/api/web-endpoints").reply(200);

    const spy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="host-text"]').setValue("127.0.0.1");
    await wrapper.findComponent('[data-test="port-text"]').setValue("8080");
    await wrapper.findComponent('[data-test="create-tunnel-btn"]').trigger("click");

    await flushPromises();

    expect(spy).toHaveBeenCalledWith("webEndpoints/create", {
      uid: "fake-uid",
      host: "127.0.0.1",
      port: 8080,
      ttl: -1,
    });
  });

  it("successfully creates a Web Endpoint with custom timeout", async () => {
    mockWebEndpointsApi.onPost("http://localhost:3000/api/web-endpoints").reply(200);

    const spy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="host-text"]').setValue("127.0.0.1");
    await wrapper.findComponent('[data-test="port-text"]').setValue("8080");

    await wrapper.findComponent('[data-test="timeout-combobox"]').setValue(600);

    await wrapper.findComponent('[data-test="create-tunnel-btn"]').trigger("click");
    await flushPromises();

    expect(spy).toHaveBeenCalledWith("webEndpoints/create", {
      uid: "fake-uid",
      host: "127.0.0.1",
      port: 8080,
      ttl: 600,
    });
  });

  it("shows alert on 403 error", async () => {
    mockWebEndpointsApi.onPost("http://localhost:3000/api/web-endpoints").reply(403);

    await wrapper.findComponent('[data-test="host-text"]').setValue("127.0.0.1");
    await wrapper.findComponent('[data-test="port-text"]').setValue("8080");
    await wrapper.findComponent('[data-test="create-tunnel-btn"]').trigger("click");
    await flushPromises();

    const alert = wrapper.findComponent('[data-test="tunnel-create-alert"]');
    expect(alert.exists()).toBe(true);
    expect(alert.text()).toBe("This device has reached the maximum allowed number of Web Endpoints");
  });

  it("successfully creates a Web Endpoint using device selector", async () => {
    mockWebEndpointsApi.onPost("http://localhost:3000/api/web-endpoints").reply(200);

    const spy = vi.spyOn(store, "dispatch");

    wrapper.unmount();
    wrapper = mount(WebEndpointCreate, {
      attachTo: document.body,
      global: {
        plugins: [[store, key], vuetify, SnackbarPlugin],
      },
      props: {
        useDevicesList: true,
        modelValue: true,
      },
    });

    await flushPromises();

    await wrapper.findComponent('[data-test="web-endpoint-autocomplete"]').setValue({
      uid: "device-abc",
      name: "device-abc-name",
      info: {
        id: "linuxmint",
        pretty_name: "Linux Mint",
      },
    });

    await wrapper.findComponent('[data-test="host-text"]').setValue("127.0.0.1");
    await wrapper.findComponent('[data-test="port-text"]').setValue(8080);

    await wrapper.findComponent('[data-test="create-tunnel-btn"]').trigger("click");
    await flushPromises();

    expect(spy).toHaveBeenCalledWith("webEndpoints/create", {
      uid: "device-abc",
      host: "127.0.0.1",
      port: 8080,
      ttl: -1,
    });
  });
});
