import { createPinia, setActivePinia } from "pinia";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import ConnectorList from "@/components/Connector/ConnectorList.vue";
import { router } from "@/router";
import { namespacesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type ConnectorListWrapper = VueWrapper<InstanceType<typeof ConnectorList>>;

describe("Connector List", () => {
  let wrapper: ConnectorListWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());

  const connectors = {
    data: [
      {
        uid: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
        tenant_id: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
        enable: true,
        address: "127.0.0.1",
        port: 2375,
        secure: false,
        status:
        {
          state: "connected",
          message: "",
        },
        tls: null,

      },
    ],
    headers: {
      "x-total-count": 1,
    },
  };

  beforeEach(() => {
    window.matchMedia = vi.fn().mockImplementation((query) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }));

    mockNamespacesApi.onGet("http://localhost:3000/api/connector?page=1&per_page=10").reply(200, connectors.data, connectors.headers);

    wrapper = mount(ConnectorList, {
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

  it("Renders the component HTML", () => {
    expect(wrapper.findComponent('[data-test="connector-list"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="status-connector"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="switch-enable"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="ip-chip"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="secure-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="menu-key-component"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="connector-list-actions"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="mdi-information-list-item"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="no-connector-validate"]').exists()).toBe(false);
  });
});
