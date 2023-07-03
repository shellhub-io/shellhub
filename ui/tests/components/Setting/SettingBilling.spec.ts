import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import SettingBilling from "@/components/Setting/SettingBilling.vue";
import { billingApi, namespacesApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

type SettingBillingWrapper = VueWrapper<InstanceType<typeof SettingBilling>>;

describe("Billing Settings Free Mode", () => {
  let wrapper: SettingBillingWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockBilling: MockAdapter;

  const members = [
    {
      id: "xxxxxxxx",
      type: "owner",
      username: "test",
    },
    {
      id: "xxxxxxxy",
      type: "observer",
      username: "test2",
    },
  ];

  const billingData = {
    active: false,
    status: "inactive",
    customer_id: "cus_test",
    subscription_id: "sub_test",
    current_period_end: 123781839,
    created_at: "",
    updated_at: "",
  };

  const namespaceData = {
    name: "test",
    owner: "test",
    tenant_id: "fake-tenant-data",
    members,
    max_devices: 3,
    devices_count: 3,
    created_at: "",
    billing: billingData,
  };

  beforeEach(() => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    // Create a mock adapter for the usersApi instance
    mockBilling = new MockAdapter(billingApi.getAxios());

    mockNamespace = new MockAdapter(namespacesApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);

    mockBilling.onGet("http://localhost:3000/api/billing/customer").reply(200, billingData);

    store.commit("billing/setSubscription", billingData);

    wrapper = mount(SettingBilling, {
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

  // WIP: Namespace not found
  // it("Renders the free plan section", () => {
  //   expect(wrapper.find('[data-test="freePlan-div"]').exists()).toBe(true);
  //   expect(wrapper.find('[data-test="freePlan-plan"]').text()).toBe("Plan: Free");
  //   expect(wrapper.find('[data-test="freePlan-description"]').text()).toBe(
  //     "Description: You can add up to 3 devices while using the 'Free' plan.",
  //   );
  // });
});
