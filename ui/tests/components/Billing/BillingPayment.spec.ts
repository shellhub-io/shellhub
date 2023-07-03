import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { billingApi, namespacesApi } from "@/api/http";
import { store, key } from "@/store";
import BillingPayment from "@/components/Billing/BillingPayment.vue";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("Billing Payment", () => {
  let wrapper: VueWrapper<InstanceType<typeof BillingPayment>>;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockCustomer: MockAdapter;

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
    customer_id: "cus_test123",
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

  const customerData = {
    id: "cus_test123",
    name: "test",
    email: "test@test.com",
    payment_methods: [
      {
        id: "pm_test123",
        number: "**** **** **** 1234",
        brand: "visa",
        exp_month: 12,
        exp_year: 2024,
        cvc: "***",
        default: false,
      },
      {
        id: "pm_test456",
        number: "**** **** **** 5678",
        brand: "mastercard",
        exp_month: 9,
        exp_year: 2026,
        cvc: "***",
        default: true,
      },
    ],
  };

  beforeEach(() => {
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockCustomer = new MockAdapter(billingApi.getAxios());
    mockNamespace = new MockAdapter(namespacesApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockCustomer.onGet("http://localhost:3000/api/billing/customer").reply(200, customerData);

    store.commit("namespace/setNamespace", namespaceData);
    store.commit("customer/setCustomer", customerData);

    wrapper = mount(BillingPayment, {
      global: {
        plugins: [[store, key], vuetify, SnackbarPlugin],
      },
      mocks: {
        $stripe: {
          elements: () => ({
            create: () => ({
              mount: () => null,
            }),
          }),
        },
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
  it("Renders the correct html", async () => {
    await flushPromises();
    expect(wrapper.findComponent('[data-test="customer-name"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="customer-email"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="payment-methods-list"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="payment-methods-item"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="payment-methods-delete-btn"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="add-card-btn"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="alert-message"]').exists()).toBe(false);
  });

  it("Detach payment method", async () => {
    await flushPromises();
    mockCustomer.onPost("http://localhost:3000/api/billing/paymentmethod/detach").reply(200);

    const detachPaymentMethodSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="payment-methods-delete-btn"]').trigger("click");

    expect(detachPaymentMethodSpy).toHaveBeenCalledWith("customer/detachPaymentMethod", "pm_test123");
  });

  it("Set default payment method", async () => {
    await flushPromises();
    mockCustomer.onPost("http://localhost:3000/api/billing/paymentmethod/default").reply(200);

    const defaultPaymentMethodSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="payment-methods-item"]').trigger("click");

    expect(defaultPaymentMethodSpy).toHaveBeenCalledWith("customer/setDefaultPaymentMethod", "pm_test123");
  });

  // TODO STRIPE TEST SAVE PAYMENT METHOD
  // it("Saves the payment method", async () => {
  //   mockCustomer.onPost("http://localhost:3000/api/billing/paymentmethod/attach", { stripeTestCard }).reply(200);
  //   const addPaymentMethodSpy = vi.spyOn(store, "dispatch");
  //   wrapper.vm.addNewCard = true;
  //   await wrapper.find('[data-test="add-card-btn"]').trigger("click");
  //   vi.runOnlyPendingTimers();
  //   expect(addPaymentMethodSpy).toHaveBeenCalledWith("customer/attachPaymentMethod", stripeTestCard);
  // });

  // it("Fails to save the payment method", async () => {
  //   await flushPromises();
  //   mockCustomer.onPost("http://localhost:3000/api/billing/paymentmethod/attach").reply(424);
  //   wrapper.vm.addNewCard = true;
  //   const addPaymentMethodSpy = vi.spyOn(store, "dispatch");
  //   await wrapper.findComponent('[data-test="payment-methods-item"]').trigger("click");
  // });
});
