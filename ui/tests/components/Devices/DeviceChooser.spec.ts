import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { nextTick } from "vue";
import { store, key } from "@/store";
import DeviceChooser from "@/components/Devices/DeviceChooser.vue";
import { router } from "@/router";
import { namespacesApi, billingApi, devicesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

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
  name: "test",
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
  registered_devices: 3,
  online_devices: 1,
  active_sessions: 0,
  pending_devices: 0,
  rejected_devices: 0,
};

describe("Device Chooser", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceChooser>>;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;
  let mockBilling: MockAdapter;
  let mockDevices: MockAdapter;

  beforeEach(async () => {
    localStorage.setItem("tenant", "fake-tenant-data");

    mockBilling = new MockAdapter(billingApi.getAxios());
    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockDevices = new MockAdapter(devicesApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockBilling.onGet("http://localhost:3000/api/billing/customer").reply(200, customerData);
    mockBilling.onGet("http://localhost:3000/api/billing/subscription").reply(200, billingData);
    mockBilling.onGet("http://localhost:3000/api/billing/devices-most-used").reply(200, devices);
    mockDevices.onGet("http://localhost:3000/api/devices?filter=&page=1&per_page=10&status=accepted").reply(200, devices);
    mockDevices.onGet("http://localhost:3000/api/stats").reply(200, stats);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("billing/setSubscription", billingData);
    store.commit("customer/setCustomer", customerData);
    store.commit("devices/setDeviceChooserStatus", true);

    wrapper = mount(DeviceChooser, {
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

  it("Renders the component data table", async () => {
    const wrapper = new DOMWrapper(document.body);
    expect(wrapper.find('[data-test="device-chooser-dialog"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="subtext"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-list-chooser-component"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="accept-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-chooser-card"]').exists()).toBe(true);
  });

  it("Render V-Tabs", async () => {
    await flushPromises();
    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="Suggested-tab"]').exists()).toBe(true);
    expect(dialog.find('[data-test="All-tab"]').exists()).toBe(true);
  });

  it("Accepts the devices listed (Suggested Devices)", async () => {
    mockBilling.onGet("http://localhost:3000/api/billing/device-most-used").reply(200);
    mockBilling.onPost("http://localhost:3000/api/billing/device-choice").reply(200, { devices });

    const StoreSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="Suggested-tab"]').trigger("click");
    await nextTick();
    await wrapper.findComponent('[data-test="accept-btn"]').trigger("click");

    await flushPromises();

    expect(StoreSpy).toHaveBeenCalledWith(
      "devices/postDevicesChooser",
      {
        devices: [
          {
            identity: {
              mac: "00:00:00:00:00:00",
            },
            info: {
              id: "linuxmint",
              pretty_name: "Linux Mint 19.3",
              version: "",
            },
            last_seen: "2020-05-20T18:58:53.276Z",
            name: "39-5e-2a",
            namespace: "user",
            online: false,
            public_key: "----- PUBLIC KEY -----",
            status: "accepted",
            tenant_id: "fake-tenant-data",
            uid: "a582b47a42d",
          },
          {
            identity: {
              mac: "00:00:00:00:00:00",
            },
            info: {
              id: "linuxmint",
              pretty_name: "Linux Mint 19.3",
              version: "",
            },
            last_seen: "2020-05-20T19:58:53.276Z",
            name: "39-5e-2b",
            namespace: "user",
            online: true,
            public_key: "----- PUBLIC KEY -----",
            status: "accepted",
            tenant_id: "fake-tenant-data",
            uid: "a582b47a42e",
          },
        ],

      },
    );
  });

  it("Accepts the devices listed(All Devices)", async () => {
    mockBilling.onGet("http://localhost:3000/api/billing/device-most-used").reply(200);
    mockBilling.onPost("http://localhost:3000/api/billing/device-choice").reply(200, { devices: [devices] });
    mockDevices.onGet("http://localhost:3000/api/devices?filter=&page=1&per_page=5&status=accepted").reply(200, devices);

    const StoreSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="All-tab"]').trigger("click");
    await nextTick();
    await wrapper.findComponent('[data-test="accept-btn"]').trigger("click");

    await flushPromises();

    expect(StoreSpy).toHaveBeenCalledWith(
      "devices/setDevicesForUserToChoose",
      {
        filter: "",
        page: 1,
        perPage: 5,
        sortStatusField: null,
        sortStatusString: "asc",
        status: "accepted",
      },
    );
  });
});
