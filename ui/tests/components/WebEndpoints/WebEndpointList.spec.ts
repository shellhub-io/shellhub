import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import { createRouter, createWebHistory } from "vue-router";
import { routes } from "@/router";
import { store, key } from "@/store";
import WebEndpointList from "@/components/WebEndpoints/WebEndpointList.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { webEndpointsApi } from "@/api/http";

type WebEndpointListWrapper = VueWrapper<InstanceType<typeof WebEndpointList>>;

const mockEndpoints = [
  {
    address: "abc123",
    namespace: "namespace",
    device: "device-1",
    host: "192.168.0.1",
    port: 8080,
    full_address: "192.168.0.1:8080",
    expires_in: "2099-12-31T23:59:59Z",
  },
];

describe("WebEndpointList.vue", () => {
  let wrapper: WebEndpointListWrapper;
  let mockWebEndpoints: MockAdapter;
  let router;

  const vuetify = createVuetify();

  beforeEach(async () => {
    router = createRouter({
      history: createWebHistory(),
      routes,
    });

    router.push("/");
    await router.isReady();

    mockWebEndpoints = new MockAdapter(webEndpointsApi.getAxios());
    mockWebEndpoints
      .onGet("http://localhost:3000/api/web-endpoints?filter=&page=1&per_page=10&sort_by=uid&order_by=asc")
      .reply(200, mockEndpoints, { "x-total-count": "1" });

    store.commit("webEndpoints/setWebEndpoints", {
      data: mockEndpoints,
      headers: { "x-total-count": "1" },
    });

    wrapper = mount(WebEndpointList, {
      global: {
        plugins: [[store, key], vuetify, [router], SnackbarPlugin],
      },
    });
  });

  it("is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("renders the DataTable", () => {
    expect(wrapper.find('[data-test="web-endpoints-table"]').exists()).toBe(true);
  });

  it("renders table headers correctly", () => {
    const headers = wrapper.findAll('[data-test="web-endpoints-table"] thead th');
    expect(headers.length).toBe(5);
    expect(headers[0].text()).toContain("Address");
    expect(headers[1].text()).toContain("Host");
    expect(headers[2].text()).toContain("Port");
    expect(headers[3].text()).toContain("Expiration Date");
    expect(headers[4].text()).toContain("Actions");
  });

  it("renders table rows with web endpoints", () => {
    const rows = wrapper.findAll('[data-test^="web-endpoint-url"]');
    expect(rows.length).toBe(mockEndpoints.length);
    expect(rows[0].text()).toContain("192.168.0.1:8080");
  });

  it("renders correct expiration text", () => {
    const text = wrapper.text();
    expect(text).toContain("Expires on");
  });

  it("renders empty state if no web endpoints", async () => {
    store.commit("webEndpoints/setWebEndpoints", {
      data: [],
      headers: { "x-total-count": "0" },
    });

    await wrapper.vm.$nextTick();

    expect(wrapper.text()).toContain("No data available");
  });
});
