import { setActivePinia, createPinia } from "pinia";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import { router } from "@/router";
import WebEndpointList from "@/components/WebEndpoints/WebEndpointList.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { webEndpointsApi } from "@/api/http";

type WebEndpointListWrapper = VueWrapper<InstanceType<typeof WebEndpointList>>;

const mockEndpoints = [
  {
    address: "abc123",
    namespace: "namespace",
    device: {
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
    host: "192.168.0.1",
    port: 8080,
    full_address: "192.168.0.1:8080",
    expires_in: "2099-12-31T23:59:59Z",
  },
];

describe("WebEndpointList.vue", () => {
  let wrapper: WebEndpointListWrapper;
  const mockWebEndpointsApi = new MockAdapter(webEndpointsApi.getAxios());

  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(() => {
    mockWebEndpointsApi.onGet("http://localhost:3000/api/web-endpoints?page=1&per_page=10")
      .reply(200, mockEndpoints, { "x-total-count": "1" });

    wrapper = mount(WebEndpointList, { global: { plugins: [vuetify, router, SnackbarPlugin] } });
  });

  it("is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("renders the DataTable", () => {
    expect(wrapper.find('[data-test="web-endpoints-table"]').exists()).toBe(true);
  });

  it("renders table headers correctly", () => {
    const headers = wrapper.findAll('[data-test="web-endpoints-table"] thead th');
    expect(headers.length).toBe(7);
    expect(headers[0].text()).toContain("Device");
    expect(headers[1].text()).toContain("Address");
    expect(headers[2].text()).toContain("Host");
    expect(headers[3].text()).toContain("Port");
    expect(headers[4].text()).toContain("Domain");
    expect(headers[5].text()).toContain("Expiration Date");
    expect(headers[6].text()).toContain("Actions");
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
    wrapper.unmount();

    mockWebEndpointsApi.onGet("http://localhost:3000/api/web-endpoints?page=1&per_page=10")
      .reply(200, [], { "x-total-count": "0" });

    wrapper = mount(WebEndpointList, { global: { plugins: [vuetify, router, SnackbarPlugin] } });

    await flushPromises();

    expect(wrapper.text()).toContain("No data available");
  });
});
