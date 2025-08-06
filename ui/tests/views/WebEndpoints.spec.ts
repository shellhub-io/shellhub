import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { describe, it, expect, beforeEach, vi } from "vitest";
import { nextTick } from "vue";
import { store, key } from "@/store";
import WebEndpoints from "@/views/WebEndpoints.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { webEndpointsApi } from "@/api/http";

type WebEndpointsWrapper = VueWrapper<InstanceType<typeof WebEndpoints>>;

const mockEndpoints = {
  data: [
    {
      address: "123abc",
      full_address: "localhost:8080",
      device: "device-uid",
      host: "localhost",
      port: 8080,
      expires_in: "2099-01-01T00:00:00Z",
    },
  ],
  headers: { "x-total-count": "1" },
};

describe("WebEndpoints.vue", () => {
  let wrapper: WebEndpointsWrapper;
  let mockWebEndpoints: MockAdapter;

  const vuetify = createVuetify();

  beforeEach(() => {
    mockWebEndpoints = new MockAdapter(webEndpointsApi.getAxios());

    mockWebEndpoints
      .onGet("http://localhost:3000/api/web-endpoints?filter=&page=1&per_page=10&sort_by=uid&order_by=asc")
      .reply(200, mockEndpoints.data, mockEndpoints.headers);

    store.commit("auth/authSuccess", {
      role: "owner",
      tenant: "fake-tenant",
    });

    store.commit("webEndpoints/setWebEndpoints", {
      data: mockEndpoints.data,
      headers: mockEndpoints.headers,
    });

    store.commit("webEndpoints/setPagePerPage", {
      page: 1,
      perPage: 10,
      filter: "",
      sortBy: "uid",
      orderBy: "asc",
    });

    store.commit("webEndpoints/setShowWebEndpoints");

    wrapper = mount(WebEndpoints, {
      global: {
        plugins: [[store, key], vuetify, SnackbarPlugin],
      },
    });
  });

  it("is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("renders the title, search field, and create button", () => {
    expect(wrapper.find("h1").text()).toContain("Web Endpoints");
    expect(wrapper.find('[data-test="search-text"]').exists()).toBe(true);
    expect(wrapper.findAll('[data-test="tunnel-create-dialog-btn"]')).toHaveLength(1);
  });

  it("renders NoItemsMessage when show = false", async () => {
    store.commit("webEndpoints/clearListEndpoints");
    await nextTick();

    expect(wrapper.find('[data-test="web-endpoints-table-component"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(true);
  });

  it("renders the WebEndpointList component when show = true", () => {
    expect(wrapper.find('[data-test="web-endpoints-table-component"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(false);
  });

  it("filters web endpoints when typing in the search input", async () => {
    const spy = vi.spyOn(store, "dispatch");

    await wrapper.getComponent('[data-test="search-text"]').setValue("localhost");
    await wrapper.getComponent('[data-test="search-text"]').trigger("keyup");

    expect(spy).toHaveBeenCalledWith("webEndpoints/search", expect.objectContaining({
      page: 1,
      perPage: 10,
      // eslint-disable-next-line vue/max-len
      filter: "W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJmdWxsX2FkZHJlc3MiLCJvcGVyYXRvciI6ImNvbnRhaW5zIiwidmFsdWUiOiJsb2NhbGhvc3QifX1d",
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
