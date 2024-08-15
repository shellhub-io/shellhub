import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { store, key } from "@/store";
import QuickConnectionList from "@/components/QuickConnection/QuickConnectionList.vue";
import { envVariables } from "@/envVariables";
import { router } from "@/router";
import { namespacesApi, billingApi, devicesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type QuickConnectionListWrapper = VueWrapper<InstanceType<typeof QuickConnectionList>>;

describe("Quick Connection List", () => {
  let wrapper: QuickConnectionListWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockDevices: MockAdapter;

  let mockBilling: MockAdapter;

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

  const members = [
    {
      id: "xxxxxxxx",
      username: "test",
      role: "owner",
    },
  ];

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
    mockDevices
      // eslint-disable-next-line vue/max-len
      .onGet("http://localhost:3000/api/devices?filter=W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJvbmxpbmUiLCJvcGVyYXRvciI6ImVxIiwidmFsdWUiOnRydWV9fV0%3D&per_page=10&status=accepted")
      .reply(200, devices);
    mockDevices.onGet("http://localhost:3000/api/stats").reply(200, stats);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("billing/setSubscription", billingData);
    store.commit("customer/setCustomer", customerData);

    wrapper = mount(QuickConnectionList, {
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

  it("Data is defined", () => {
    expect(wrapper.vm.$data).toBeDefined();
  });

  it("Renders the devices list", () => {
    expect(wrapper.find('[data-test="devices-list"]').exists()).toBe(true);
  });

  it("Renders each device card", () => {
    expect(wrapper.find('[data-test="device-list-item"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-name"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-info"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-ssh-id"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-tags"]').exists()).toBe(true);
  });

  it("Renders the copy ID button", () => {
    expect(wrapper.find('[data-test="copy-id-button"]').exists()).toBe(true);
  });

  it("Renders the tag chips", () => {
    expect(wrapper.find('[data-test="tag-chip"]').exists()).toBe(true);
  });

  it("Renders the no tags chip", async () => {
    // Change the value of tags[0] to an empty string for the first device
    devices[0].tags[0] = "";
    await flushPromises();
    expect(wrapper.find('[data-test="no-tags-chip"]').exists()).toBe(true);
  });

  it("Renders the no online devices message", async () => {
    mockDevices.reset();
    // Test with an empty online filtered request
    mockDevices
    // eslint-disable-next-line vue/max-len
      .onGet("http://localhost:3000/api/devices?filter=W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJvbmxpbmUiLCJvcGVyYXRvciI6ImVxIiwidmFsdWUiOnRydWV9fV0%3D&per_page=10&status=accepted")
      .reply(200, []);
    await flushPromises();
    expect(wrapper.find('[data-test="no-online-devices"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="no-online-devices-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="no-online-devices-message"]').exists()).toBe(true);
  });
});
