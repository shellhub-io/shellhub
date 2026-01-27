import { setActivePinia, createPinia } from "pinia";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { billingApi, namespacesApi } from "@/api/http";
import BillingPayment from "@/components/Billing/BillingPayment.vue";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useCustomerStore from "@/store/modules/customer";
import useNamespacesStore from "@/store/modules/namespaces";
import { INamespace, INamespaceMember } from "@/interfaces/INamespace";

describe("Billing Payment", () => {
  let wrapper: VueWrapper<InstanceType<typeof BillingPayment>>;
  setActivePinia(createPinia());
  const customerStore = useCustomerStore();
  const namespacesStore = useNamespacesStore();
  const vuetify = createVuetify();
  const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());
  const mockBillingApi = new MockAdapter(billingApi.getAxios());

  const members = [
    {
      id: "xxxxxxxx",
      role: "owner" as const,
    },
    {
      id: "xxxxxxxy",
      role: "observer" as const,
    },
  ] as INamespaceMember[];

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
    settings: {
      session_record: true,
    },
    devices_accepted_count: 3,
    devices_rejected_count: 0,
    devices_pending_count: 0,
    type: "team" as const,
  };

  const customerData = {
    id: "cus_test123",
    name: "test",
    email: "test@test.com",
    payment_methods: [
      {
        id: "pm_test123",
        number: "1234 1234 1234 1234",
        brand: "visa",
        exp_month: 12,
        exp_year: 2999,
        cvc: "123",
        default: false,
      },
      {
        id: "pm_test456",
        number: "1234 1234 1234 5678",
        brand: "mastercard",
        exp_month: 12,
        exp_year: 2999,
        cvc: "123",
        default: true,
      },
    ],
  };

  beforeEach(() => {
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespacesApi.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockBillingApi.onGet("http://localhost:3000/api/billing/customer").reply(200, customerData);

    namespacesStore.currentNamespace = namespaceData as INamespace;

    wrapper = mount(BillingPayment, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
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
    mockBillingApi.onPost("http://localhost:3000/api/billing/paymentmethod/detach").reply(200);

    const detachPaymentMethodSpy = vi.spyOn(customerStore, "detachPaymentMethod");

    await wrapper.findComponent('[data-test="payment-methods-delete-btn"]').trigger("click");

    expect(detachPaymentMethodSpy).toHaveBeenCalledWith("pm_test123");
  });

  it("Set default payment method", async () => {
    await flushPromises();
    mockBillingApi.onPost("http://localhost:3000/api/billing/paymentmethod/default").reply(200);

    const defaultPaymentMethodSpy = vi.spyOn(customerStore, "setDefaultPaymentMethod");

    await wrapper.findComponent('[data-test="payment-methods-item"]').trigger("click");

    expect(defaultPaymentMethodSpy).toHaveBeenCalledWith("pm_test123");
  });

  // TODO STRIPE TEST SAVE PAYMENT METHOD
  // it("Saves the payment method", async () => {
  //   mockBillingApi.onPost("http://localhost:3000/api/billing/paymentmethod/attach", { stripeTestCard }).reply(200);
  //   const addPaymentMethodSpy = vi.spyOn(customerStore, "attachPaymentMethod");
  //   wrapper.vm.addNewCard = true;
  //   await wrapper.find('[data-test="add-card-btn"]').trigger("click");
  //   vi.runOnlyPendingTimers();
  //   expect(addPaymentMethodSpy).toHaveBeenCalledWith(stripeTestCard);
  // });

  // it("Fails to save the payment method", async () => {
  //   await flushPromises();
  //   mockBillingApi.onPost("http://localhost:3000/api/billing/paymentmethod/attach").reply(424);
  //   wrapper.vm.addNewCard = true;
  //   const addPaymentMethodSpy = vi.spyOn(customerStore, "attachPaymentMethod");
  //   await wrapper.findComponent('[data-test="payment-methods-item"]').trigger("click");
  // });
});
