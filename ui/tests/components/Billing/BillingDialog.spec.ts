import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import BillingDialog from "@/components/Billing/BillingDialog.vue";
import { billingApi, namespacesApi } from "@/api/http";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useCustomerStore from "@/store/modules/customer";

describe("Billing Dialog", async () => {
  setActivePinia(createPinia());
  const customerStore = useCustomerStore();
  const vuetify = createVuetify();
  const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());
  const mockBillingApi = new MockAdapter(billingApi.getAxios());

  const members = [
    {
      id: "xxxxxxxx",
      username: "test",
      role: "owner",
    },
  ];

  const namespaceData = {
    name: "test",
    owner: "test",
    tenant_id: "fake-tenant-data",
    members,
    settings: {
      session_record: true,
    },
    max_devices: 3,
    devices_count: 3,
    created_at: "",
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

  localStorage.setItem("tenant", "fake-tenant-data");
  envVariables.isCloud = true;

  mockNamespacesApi.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
  mockBillingApi.onPost("http://localhost:3000/api/billing/customer").reply(200);
  mockBillingApi.onGet("http://localhost:3000/api/billing/customer").reply(200, customerData);

  const wrapper = mount(BillingDialog, {
    global: {
      plugins: [vuetify, router, SnackbarPlugin],
    },
  });
  wrapper.vm.showCheckoutDialog = true;
  await flushPromises();

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component (Welcome window)", () => {
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Renders the component (Payment Details window)", async () => {
    wrapper.vm.el = 2;
    await flushPromises();
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Renders the component (Checkout window)", async () => {
    wrapper.vm.el = 3;
    await flushPromises();
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Renders the component (Success window)", async () => {
    wrapper.vm.el = 4;
    await flushPromises();
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Goes to next and previous step when clicking on buttons", async () => {
    wrapper.vm.resetDialog();
    wrapper.vm.showCheckoutDialog = true;
    await flushPromises();
    expect(wrapper.vm.el).toEqual(1);
    await wrapper.findComponent('[data-test="payment-letter-next-button"]').trigger("click");
    expect(wrapper.vm.el).toEqual(2);
    await wrapper.findComponent('[data-test="payment-details-back-button"]').trigger("click");
    expect(wrapper.vm.el).toEqual(1);
  });

  it("Subscribes to Premium", async () => {
    wrapper.vm.el = 3;
    await flushPromises();

    mockBillingApi.onPost("http://localhost:3000/api/billing/subscription").reply(200);

    const subscribeSpy = vi.spyOn(customerStore, "createSubscription");
    await wrapper.findComponent('[data-test="checkout-button"]').trigger("click");

    expect(subscribeSpy).toHaveBeenCalled();
  });
});
