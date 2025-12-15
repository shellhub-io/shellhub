import { createPinia, setActivePinia } from "pinia";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { describe, it, expect, beforeEach, vi } from "vitest";
import { nextTick } from "vue";
import WebEndpoints from "@/views/WebEndpoints.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { webEndpointsApi } from "@/api/http";
import { router } from "@/router";
import useWebEndpointsStore from "@/store/modules/web_endpoints";
import { IWebEndpoint } from "@/interfaces/IWebEndpoints";
import useAuthStore from "@/store/modules/auth";
import useDevicesStore from "@/store/modules/devices";

type WebEndpointsWrapper = VueWrapper<InstanceType<typeof WebEndpoints>>;

const mockWebEndpoints = [
  {
    address: "123abc",
    full_address: "localhost:8080",
    device_uid: "device-abc",
    device: {
      uid: "device-abc",
      name: "device-abc-name",
      info: {
        id: "linuxmint",
        pretty_name: "Linux Mint",
      },
    },
    host: "localhost",
    port: 8080,
    expires_in: "2099-01-01T00:00:00Z",
  },
] as IWebEndpoint[];

describe("WebEndpoints.vue", () => {
  let wrapper: WebEndpointsWrapper;
  const mockWebEndpointsApi = new MockAdapter(webEndpointsApi.getAxios());
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const devicesStore = useDevicesStore();
  const webEndpointsStore = useWebEndpointsStore();
  const vuetify = createVuetify();

  devicesStore.fetchDeviceList = vi.fn().mockResolvedValue([]);

  beforeEach(async () => {
    mockWebEndpointsApi
      .onGet("http://localhost:3000/api/web-endpoints?page=1&per_page=10")
      .reply(200, mockWebEndpoints, { "x-total-count": "1" });

    authStore.role = "owner";
    await webEndpointsStore.fetchWebEndpointsList();

    wrapper = mount(WebEndpoints, {
      global: {
        plugins: [router, vuetify, SnackbarPlugin],
      },
    });
  });

  it("is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("renders the title, search field, and create button", () => {
    expect(wrapper.text()).toContain("Web Endpoints");
    expect(wrapper.find('[data-test="search-text"]').exists()).toBe(true);
    expect(wrapper.findAll('[data-test="tunnel-create-dialog-btn"]')).toHaveLength(1);
  });

  it("renders NoItemsMessage when show = false", async () => {
    webEndpointsStore.showWebEndpoints = false;
    await nextTick();
    expect(wrapper.find('[data-test="web-endpoints-table-component"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(true);
  });

  it("renders the WebEndpointList component when show = true", () => {
    expect(wrapper.find('[data-test="web-endpoints-table-component"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(false);
  });

  it("filters web endpoints when typing in the search input", async () => {
    // eslint-disable-next-line vue/max-len
    mockWebEndpointsApi.onGet("http://localhost:3000/api/web-endpoints?filter=W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJhZGRyZXNzIiwib3BlcmF0b3IiOiJjb250YWlucyIsInZhbHVlIjoibG9jYWxob3N0In19XQ%3D%3D&page=1&per_page=10").reply(200, mockWebEndpoints, { "x-total-count": "1" });
    const storeSpy = vi.spyOn(webEndpointsStore, "fetchWebEndpointsList");

    await wrapper.getComponent('[data-test="search-text"]').setValue("localhost");
    await wrapper.getComponent('[data-test="search-text"]').trigger("keyup");

    expect(storeSpy).toHaveBeenCalledWith(expect.objectContaining({
      // eslint-disable-next-line vue/max-len
      filter: "W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJhZGRyZXNzIiwib3BlcmF0b3IiOiJjb250YWlucyIsInZhbHVlIjoibG9jYWxob3N0In19XQ==",
    }));
  });

  it("opens WebEndpointCreate dialog when clicking the button", async () => {
    expect(wrapper.findComponent({ name: "WebEndpointCreate" }).props("modelValue")).toBe(false);

    await wrapper.findAll('[data-test="tunnel-create-dialog-btn"]')[0].trigger("click");
    await wrapper.vm.$nextTick();

    expect(wrapper.findComponent({ name: "WebEndpointCreate" }).props("modelValue")).toBe(true);
  });

  it("calls searchWebEndpoints when WebEndpointCreate emits update", async () => {
    const searchSpy = vi.spyOn(wrapper.vm, "searchWebEndpoints");

    await wrapper.findAll('[data-test="tunnel-create-dialog-btn"]')[0].trigger("click");

    wrapper.findComponent({ name: "WebEndpointCreate" }).vm.$emit("update");
    await wrapper.vm.$nextTick();

    expect(searchSpy).toHaveBeenCalled();
  });
});
