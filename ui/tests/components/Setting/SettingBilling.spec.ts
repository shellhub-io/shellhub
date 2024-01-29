import { createVuetify } from "vuetify";
import { mount, VueWrapper, flushPromises } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { nextTick } from "vue";
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
      username: "test",
      role: "owner",
    },
  ];

  const billingData = {
    active: false,
    status: "inactive",
    customer_id: "cus_test",
    subscription_id: "sub_test",
    current_period_end: 2068385820,
    created_at: "",
    updated_at: "",
    invoices: [],
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

  beforeEach(() => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;
    // Create a mock adapter for the billingApi and namespacesApi instance
    mockBilling = new MockAdapter(billingApi.getAxios());

    mockNamespace = new MockAdapter(namespacesApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockBilling.onGet("http://localhost:3000/api/billing/customer").reply(200, customerData);
    mockBilling.onGet("http://localhost:3000/api/billing/subscription").reply(200, billingData);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("billing/setSubscription", billingData);

    wrapper = mount(SettingBilling, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
  });

  afterEach(() => {
    wrapper.unmount();
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

  it("Renders the free plan section", () => {
    expect(wrapper.find('[data-test="freePlan-div"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="portal-button"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="subscribe-button"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-portal-text"]').exists()).toBe(true);
  });

  it("Dialog pops up on free mode", async () => {
    await wrapper.findComponent('[data-test="subscribe-button"]').trigger("click");
    expect(document.querySelector('[data-test="dialog-checkout"]')).not.toBeNull();
  });

  it("Renders dialog text", async () => {
    expect(wrapper.find('[data-test="card-first-page"]').exists());
    expect(wrapper.find('[data-test="card-second-page"]').exists());
    expect(wrapper.find('[data-test="card-third-page"]').exists());
    expect(wrapper.find('[data-test="card-fourth-page"]').exists());
  });

  it("Render pagination", () => {
    expect(wrapper.findComponent('[data-test="payment-letter-next-button"]').exists());
    expect(wrapper.findComponent('[data-test="payment-letter-close-button"]').exists());
    expect(wrapper.findComponent('[data-test="payment-details-back-button"]').exists());
    expect(wrapper.findComponent('[data-test="payment-details-next-button"]').exists());
    expect(wrapper.findComponent('[data-test="checkout-back-button"]').exists());
    expect(wrapper.findComponent('[data-test="checkout-button"]').exists());
    expect(wrapper.findComponent('[data-test="successful-close-button"]').exists());
  });

  it("Pagination logic test", async () => {
    await wrapper.findComponent('[data-test="subscribe-button"]').trigger("click");
    expect(wrapper.vm.el).toEqual(1);
    await wrapper.findComponent('[data-test="payment-letter-next-button"]').trigger("click");
    expect(wrapper.vm.el).toEqual(2);
    await wrapper.findComponent('[data-test="payment-details-back-button"]').trigger("click");
    expect(wrapper.vm.el).toEqual(1);
  });

  it("Subscribe to Premium", async () => {
    await wrapper.findComponent('[data-test="subscribe-button"]').trigger("click");
    wrapper.vm.el = 3;

    await nextTick();
    await flushPromises();

    mockBilling.onPost("http://localhost:3000/api/billing/subscription").reply(200);

    const subscribeSpy = vi.spyOn(store, "dispatch");
    await wrapper.findComponent('[data-test="checkout-button"]').trigger("click");

    vi.runOnlyPendingTimers();

    expect(subscribeSpy).toHaveBeenCalledWith("customer/createSubscription");
  });
});

describe("Billing Settings Premium Usage", () => {
  let wrapper: SettingBillingWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockBilling: MockAdapter;

  const members = [
    {
      id: "xxxxxxxx",
      username: "test",
      role: "owner",
    },
  ];

  const billingData = {
    active: true,
    status: "active",
    customer_id: "cus_test",
    subscription_id: "sub_test",
    current_period_end: 2068385820,
    created_at: "",
    updated_at: "",
    invoices: [
      {
        id: "xxxxx",
        status: "open",
        currency: "brl",
        amount: 12,
      },
    ],
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

  beforeEach(() => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;
    // Create a mock adapter for the usersApi instance
    mockBilling = new MockAdapter(billingApi.getAxios());

    mockNamespace = new MockAdapter(namespacesApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockBilling.onGet("http://localhost:3000/api/billing/customer").reply(200, customerData);
    mockBilling.onGet("http://localhost:3000/api/billing/subscription").reply(200, billingData);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("billing/setSubscription", billingData);
    store.commit("customer/setCustomer", customerData);

    wrapper = mount(SettingBilling, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("Does not render free mode component", () => {
    expect(wrapper.find('[data-test="freePlan-div"]').exists()).toBe(false);
    expect(wrapper.findComponent('[data-test="subscribe-button"]').exists()).toBe(false);
  });

  it("Render premium usage component", () => {
    expect(wrapper.find('[data-test="premiumPlan-div"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="subscriptionActive-div"]').exists()).toBe(true);
  });

  it("Render alerts for status", async () => {
    const billingData = {
      active: true,
      status: "to_cancel_at_end_of_period",
      customer_id: "cus_test",
      subscription_id: "sub_test",
      current_period_end: 2068385820,
      created_at: "",
      updated_at: "",
      invoices: [
        {
          id: "xxxxx",
          status: "open",
          currency: "brl",
          amount: 12,
        },
      ],
    };
    mockBilling.onGet("http://localhost:3000/api/billing/subscription").reply(200, billingData);
    store.commit("billing/setSubscription", billingData);
    await nextTick();
    await flushPromises();
    expect(wrapper.find('[data-test="message-alert"]')).toBeTruthy();
  });
});
