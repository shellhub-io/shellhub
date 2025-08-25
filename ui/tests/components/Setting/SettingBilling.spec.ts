import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper, flushPromises } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { nextTick } from "vue";
import SettingBilling from "@/components/Setting/SettingBilling.vue";
import { billingApi, namespacesApi } from "@/api/http";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";
import useBillingStore from "@/store/modules/billing";

type SettingBillingWrapper = VueWrapper<InstanceType<typeof SettingBilling>>;

setActivePinia(createPinia());
const authStore = useAuthStore();
const billingStore = useBillingStore();
const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());
const mockBillingApi = new MockAdapter(billingApi.getAxios());
const vuetify = createVuetify();

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

describe("Billing Settings Free Mode", () => {
  let wrapper: SettingBillingWrapper;

  beforeEach(() => {
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespacesApi.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockBillingApi.onGet("http://localhost:3000/api/billing/customer").reply(200, customerData);
    mockBillingApi.onGet("http://localhost:3000/api/billing/subscription").reply(200, billingData);

    authStore.role = "owner";
    wrapper = mount(SettingBilling, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
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

  it("Renders the free plan section", () => {
    expect(wrapper.find('[data-test="billing-card"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-header"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-subtitle"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="subscribe-button"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-details-list"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-portal-section"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-portal-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-portal-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-portal-description"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-portal-button"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-divider"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-plan-section"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-plan-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-plan-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-plan-description-free"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-plan-description-premium"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="billing-plan-free"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-plan-premium"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="billing-active-section"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="billing-status-section"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="billing-status-icon"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="billing-status-title"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="billing-status-message"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="billing-total-section"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="billing-total-icon"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="billing-total-title"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="billing-total-amount"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="billing-end-date-section"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="billing-end-date-icon"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="billing-end-date-title"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="billing-end-date"]').exists()).toBe(false);
  });
});

describe("Billing Settings Premium Usage", () => {
  let wrapper: SettingBillingWrapper;

  const members = [
    {
      id: "xxxxxxxx",
      username: "test",
      role: "owner",
    },
  ];

  const billingData = {
    id: "sub_test",
    active: true,
    status: "active",
    customer_id: "cus_test",
    subscription_id: "sub_test",
    current_period_end: 2068385820,
    end_at: 2068385820,
    created_at: "",
    updated_at: "",
    invoices: [
      {
        id: "xxxxx",
        status: "open" as const,
        currency: "brl" as const,
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

  beforeEach(() => {
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespacesApi.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockBillingApi.onGet("http://localhost:3000/api/billing/customer").reply(200, customerData);
    mockBillingApi.onGet("http://localhost:3000/api/billing/subscription").reply(200, billingData);

    billingStore.billing = billingData;
    authStore.role = "owner";
    wrapper = mount(SettingBilling, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
    });
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("Does not render free mode component", () => {
    expect(wrapper.find('[data-test="billing-plan-description-free"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="billing-plan-free"]').exists()).toBe(false);
  });

  it("Render premium usage component", () => {
    expect(wrapper.find('[data-test="billing-plan-description-premium"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-plan-premium"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-active-section"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-status-section"]').exists());
    expect(wrapper.find('[data-test="billing-status-icon"]').exists());
    expect(wrapper.find('[data-test="billing-status-title"]').exists());
    expect(wrapper.find('[data-test="billing-status-message"]').exists());
    expect(wrapper.find('[data-test="billing-total-section"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-total-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-total-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-total-amount"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-end-date-section"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-end-date-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-end-date-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="billing-end-date"]').exists()).toBe(true);
  });

  it("Render alerts for status", async () => {
    const billingData = {
      id: "sub_test",
      active: true,
      status: "to_cancel_at_end_of_period",
      customer_id: "cus_test",
      subscription_id: "sub_test",
      current_period_end: 2068385820,
      end_at: 2068385820,
      created_at: "",
      updated_at: "",
      invoices: [
        {
          id: "xxxxx",
          status: "open" as const,
          currency: "brl" as const,
          amount: 12,
        },
      ],
    };
    mockBillingApi.onGet("http://localhost:3000/api/billing/subscription").reply(200, billingData);
    billingStore.billing = billingData;
    await nextTick();
    await flushPromises();
    expect(wrapper.find('[data-test="billing-status-message"]')).toBeTruthy();
  });
});
