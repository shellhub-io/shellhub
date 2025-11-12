import { setActivePinia, createPinia } from "pinia";
import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import WebEndpointCreate from "@/components/WebEndpoints/WebEndpointCreate.vue";
import { devicesApi, webEndpointsApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useWebEndpointsStore from "@/store/modules/web_endpoints";

type WebEndpointCreateWrapper = VueWrapper<InstanceType<typeof WebEndpointCreate>>;

describe("WebEndpointCreate.vue", () => {
  let wrapper: WebEndpointCreateWrapper;
  const mockDevicesApi = new MockAdapter(devicesApi.getAxios());
  const mockWebEndpointsApi = new MockAdapter(webEndpointsApi.getAxios());
  setActivePinia(createPinia());
  const webEndpointsStore = useWebEndpointsStore();
  const vuetify = createVuetify();

  beforeEach(async () => {
    mockDevicesApi
      .onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted")
      .reply(200, []);

    wrapper = mount(WebEndpointCreate, {
      attachTo: document.body,
      global: { plugins: [createVuetify(), SnackbarPlugin] },
      props: { uid: "fake-uid", useDevicesList: false, modelValue: true },
    });

    await flushPromises();
  });

  it("is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("renders the dialog and basic fields", async () => {
    const dialog = new DOMWrapper(document.body);
    await flushPromises();

    expect(dialog.text()).toContain("Create Device Web Endpoint");
    expect(dialog.find('[data-test="host-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="port-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="timeout-combobox"]').exists()).toBe(true);
    expect(dialog.find('[data-test="custom-timeout"]').exists()).toBe(false);
    expect(dialog.find('[data-test="tls-enabled-checkbox"]').exists()).toBe(true);
    expect(dialog.find('[data-test="create-tunnel-btn"]').exists()).toBe(true);
  });

  it("successfully creates a Web Endpoint without TLS", async () => {
    mockWebEndpointsApi.onPost("http://localhost:3000/api/web-endpoints").reply(200);

    const storeSpy = vi.spyOn(webEndpointsStore, "createWebEndpoint");

    await wrapper.findComponent('[data-test="host-text"]').setValue("127.0.0.1");
    await wrapper.findComponent('[data-test="port-text"]').setValue("8080");
    await wrapper.findComponent('[data-test="create-tunnel-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith({
      uid: "fake-uid",
      host: "127.0.0.1",
      port: 8080,
      ttl: -1,
    });
  });

  it("successfully creates a Web Endpoint with custom timeout", async () => {
    mockWebEndpointsApi.onPost("http://localhost:3000/api/web-endpoints").reply(200);

    const storeSpy = vi.spyOn(webEndpointsStore, "createWebEndpoint");

    await wrapper.findComponent('[data-test="host-text"]').setValue("127.0.0.1");
    await wrapper.findComponent('[data-test="port-text"]').setValue("8080");
    await wrapper.findComponent('[data-test="timeout-combobox"]').setValue(600);
    await wrapper.findComponent('[data-test="create-tunnel-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith({
      uid: "fake-uid",
      host: "127.0.0.1",
      port: 8080,
      ttl: 600,
    });
  });

  it("successfully creates a Web Endpoint with TLS enabled", async () => {
    mockWebEndpointsApi.onPost("http://localhost:3000/api/web-endpoints").reply(200);

    const storeSpy = vi.spyOn(webEndpointsStore, "createWebEndpoint");

    await wrapper.findComponent('[data-test="host-text"]').setValue("192.168.1.1");
    await wrapper.findComponent('[data-test="port-text"]').setValue("443");

    const tlsCheckbox = wrapper.findComponent('[data-test="tls-enabled-checkbox"]');
    await tlsCheckbox.setValue(true);
    await flushPromises();

    await wrapper.findComponent('[data-test="tls-domain-text"]').setValue("example.com");
    await wrapper.findComponent('[data-test="create-tunnel-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith({
      uid: "fake-uid",
      host: "192.168.1.1",
      port: 443,
      ttl: -1,
      tls: {
        enabled: true,
        verify: false,
        domain: "example.com",
      },
    });
  });

  it("successfully creates a Web Endpoint with TLS and certificate verification", async () => {
    mockWebEndpointsApi.onPost("http://localhost:3000/api/web-endpoints").reply(200);

    const storeSpy = vi.spyOn(webEndpointsStore, "createWebEndpoint");

    await wrapper.findComponent('[data-test="host-text"]').setValue("10.0.0.5");
    await wrapper.findComponent('[data-test="port-text"]').setValue("8443");

    const tlsCheckbox = wrapper.findComponent('[data-test="tls-enabled-checkbox"]');
    await tlsCheckbox.setValue(true);
    await flushPromises();

    const tlsVerifyCheckbox = wrapper.findComponent('[data-test="tls-verify-checkbox"]');
    await tlsVerifyCheckbox.setValue(true);
    await flushPromises();

    await wrapper.findComponent('[data-test="tls-domain-text"]').setValue("secure.local");
    await wrapper.findComponent('[data-test="create-tunnel-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith({
      uid: "fake-uid",
      host: "10.0.0.5",
      port: 8443,
      ttl: -1,
      tls: {
        enabled: true,
        verify: true,
        domain: "secure.local",
      },
    });
  });

  it("shows TLS domain field when TLS is enabled", async () => {
    const tlsCheckbox = wrapper.findComponent('[data-test="tls-enabled-checkbox"]');
    await tlsCheckbox.setValue(true);
    await flushPromises();

    expect(wrapper.findComponent('[data-test="tls-domain-text"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="tls-verify-checkbox"]').exists()).toBe(true);
  });

  it("validates TLS domain is required when TLS is enabled", async () => {
    const dialog = new DOMWrapper(document.body);

    await wrapper.findComponent('[data-test="host-text"]').setValue("127.0.0.1");
    await wrapper.findComponent('[data-test="port-text"]').setValue("8080");

    const tlsCheckbox = wrapper.findComponent('[data-test="tls-enabled-checkbox"]');
    await tlsCheckbox.setValue(true);
    await flushPromises();

    const createButton = dialog.find('[data-test="create-tunnel-btn"]');
    expect(createButton.attributes("disabled")).toBeDefined();
  });

  it("accepts valid domain formats for TLS", async () => {
    mockWebEndpointsApi.onPost("http://localhost:3000/api/web-endpoints").reply(200);

    const storeSpy = vi.spyOn(webEndpointsStore, "createWebEndpoint");

    await wrapper.findComponent('[data-test="host-text"]').setValue("127.0.0.1");
    await wrapper.findComponent('[data-test="port-text"]').setValue("443");

    const tlsCheckbox = wrapper.findComponent('[data-test="tls-enabled-checkbox"]');
    await tlsCheckbox.setValue(true);
    await flushPromises();

    await wrapper.findComponent('[data-test="tls-domain-text"]').setValue("192.168.1.1");
    await wrapper.findComponent('[data-test="create-tunnel-btn"]').trigger("click");
    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        tls: expect.objectContaining({ domain: "192.168.1.1" }),
      }),
    );
  });

  it("shows alert on 403 error", async () => {
    mockWebEndpointsApi.onPost("http://localhost:3000/api/web-endpoints").reply(403);

    await wrapper.findComponent('[data-test="host-text"]').setValue("127.0.0.1");
    await wrapper.findComponent('[data-test="port-text"]').setValue("8080");
    await wrapper.findComponent('[data-test="create-tunnel-btn"]').trigger("click");
    await flushPromises();

    const dialog = new DOMWrapper(document.body);
    const alert = dialog.find('[data-test="form-dialog-alert"]');
    expect(alert.exists()).toBe(true);
    expect(alert.text()).toContain("This device has reached the maximum allowed number of Web Endpoints");
  });

  it("successfully creates a Web Endpoint using device selector", async () => {
    mockWebEndpointsApi.onPost("http://localhost:3000/api/web-endpoints").reply(200);

    const storeSpy = vi.spyOn(webEndpointsStore, "createWebEndpoint");

    wrapper.unmount();
    wrapper = mount(WebEndpointCreate, {
      attachTo: document.body,
      global: {
        plugins: [vuetify, SnackbarPlugin],
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

    expect(storeSpy).toHaveBeenCalledWith({
      uid: "device-abc",
      host: "127.0.0.1",
      port: 8080,
      ttl: -1,
    });
  });

  it("successfully creates a Web Endpoint using device selector with TLS", async () => {
    mockWebEndpointsApi.onPost("http://localhost:3000/api/web-endpoints").reply(200);

    const storeSpy = vi.spyOn(webEndpointsStore, "createWebEndpoint");

    wrapper.unmount();
    wrapper = mount(WebEndpointCreate, {
      attachTo: document.body,
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
      props: {
        useDevicesList: true,
        modelValue: true,
      },
    });

    await flushPromises();

    await wrapper.findComponent('[data-test="web-endpoint-autocomplete"]').setValue({
      uid: "device-xyz",
      name: "device-xyz-name",
      info: {
        id: "ubuntu",
        pretty_name: "Ubuntu",
      },
    });

    await wrapper.findComponent('[data-test="host-text"]').setValue("10.10.10.10");
    await wrapper.findComponent('[data-test="port-text"]').setValue(443);

    const tlsCheckbox = wrapper.findComponent('[data-test="tls-enabled-checkbox"]');
    await tlsCheckbox.setValue(true);
    await flushPromises();

    await wrapper.findComponent('[data-test="tls-domain-text"]').setValue("device.local");
    await wrapper.findComponent('[data-test="create-tunnel-btn"]').trigger("click");

    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith({
      uid: "device-xyz",
      host: "10.10.10.10",
      port: 443,
      ttl: -1,
      tls: {
        enabled: true,
        verify: false,
        domain: "device.local",
      },
    });
  });
});
