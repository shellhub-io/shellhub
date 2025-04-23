import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { store, key } from "@/store";
import ConnectorList from "@/components/Connector/ConnectorList.vue";
import { envVariables } from "@/envVariables";
import { router } from "@/router";
import { namespacesApi, billingApi, devicesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type ConnectorListWrapper = VueWrapper<InstanceType<typeof ConnectorList>>;

describe("Connector List", () => {
  let wrapper: ConnectorListWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockDevices: MockAdapter;

  let mockBilling: MockAdapter;

  const devices = [
    {
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
      tags: ["test"],
    },
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
      tags: ["test"],
    },
  ];

  const members = [
    {
      id: "xxxxxxxx",
      username: "test",
      role: "owner",
    },
  ];

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
          State: "connected",
          Message: "",
        },
        tls: null,

      },
    ],
    headers: {
      "x-total-count": 1,
    },
  };

  const billingData = {
    active: false,
    status: "canceled",
    customer_id: "cus_test",
    subscription_id: "sub_test",
    current_period_end: 2068385820,
    created_at: "",
    updated_at: "",
    invoices: [],
  };

  const namespaceData = {
    name: "user",
    owner: "xxxxxxxx",
    tenant_id: "fake-tenant-data",
    members,
    max_devices: 3,
    devices_count: 3,
    devices: 2,
    created_at: "",
    billing: billingData,
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

  const customerData = {
    id: "cus_test",
    name: "test",
    email: "test@test.com",
    payment_methods: [
      {
        id: "test_id",
        number: "xxxxxxxxxxxx4242",
        brand: "visa",
        exp_month: 3,
        exp_year: 2029,
        cvc: "",
        default: true,
      },
    ],
  };

  const stats = {
    registered_devices: 2,
    online_devices: 1,
    active_sessions: 0,
    pending_devices: 0,
    rejected_devices: 0,
  };

  beforeEach(async () => {
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

    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockBilling = new MockAdapter(billingApi.getAxios());
    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockDevices = new MockAdapter(devicesApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockBilling.onGet("http://localhost:3000/api/billing/customer").reply(200, customerData);
    mockBilling.onGet("http://localhost:3000/api/billing/subscription").reply(200, billingData);
    mockBilling.onGet("http://localhost:3000/api/billing/devices-most-used").reply(200, devices);
    mockNamespace.onGet("http://localhost:3000/api/connector?page=1&per_page=10").reply(200, connectors);
    mockDevices.onGet("http://localhost:3000/api/stats").reply(200, stats);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("billing/setSubscription", billingData);
    store.commit("customer/setCustomer", customerData);
    store.commit("connectors/setConnectors", connectors);

    wrapper = mount(ConnectorList, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component HTML", async () => {
    expect(wrapper.findComponent('[data-test="connector-list"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="status-connector"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="switch-enable"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="ip-chip"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="secure-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="menu-key-component"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="connector-list-actions"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="mdi-information-list-item"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="no-connector-validate"]').exists()).toBe(false);
  });
});
