import { shallowMount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import { store, key } from "@/store";
import MemberList from "@/components/Team/Member/MemberList.vue";
import { envVariables } from "@/envVariables";
import { router } from "@/router";
import { namespacesApi, devicesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type MemberListWrapper = VueWrapper<InstanceType<typeof MemberList>>;

describe("Member List", () => {
  let wrapper: MemberListWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockDevices: MockAdapter;

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
    members: [
      {
        id: "xxxxxxxx",
        username: "test",
        email: "test@test.com",
        role: "owner",
        status: "active", // Ensure 'status' is present
        added_at: "2024-01-01T12:00:00Z", // Example valid date
      },
    ],
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
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockDevices = new MockAdapter(devicesApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockDevices.onGet("http://localhost:3000/api/stats").reply(200, stats);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("customer/setCustomer", customerData);

    wrapper = shallowMount(MemberList, {
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
    expect(wrapper.findComponent('[data-test="member-table"]').exists()).toBe(true);
  });
});
