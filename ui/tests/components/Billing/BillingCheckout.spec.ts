import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import { store } from "@/store";
import BillingCheckout from "@/components/Billing/BillingCheckout.vue";
import { billingApi } from "@/api/http";

describe("Billing Checkout", () => {
  let wrapper: VueWrapper<InstanceType<typeof BillingCheckout>>;
  const vuetify = createVuetify();
  let mock: MockAdapter;

  const customerData = {
    id: "cus_test123",
    name: "testuser",
    email: "test@test.com",
    payment_methods: [
      {
        id: "pm_test123",
        number: "**** **** **** 1234",
        brand: "visa",
        exp_month: 12,
        exp_year: 2024,
        cvc: "***",
        default: true,
      },
      {
        id: "pm_test456",
        number: "**** **** **** 5678",
        brand: "mastercard",
        exp_month: 9,
        exp_year: 2023,
        cvc: "***",
        default: false,
      },
    ],
  };
  beforeEach(async () => {
    mock = new MockAdapter(billingApi.getAxios());
    mock.onGet("http://localhost:3000/api/billing/customer").reply(200, customerData);
    store.commit("customer/setCustomer", customerData);
    wrapper = mount(BillingCheckout, {
      global: {
        plugins: [[store], vuetify],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("renders the correct html", async () => {
    expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="sub-title"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="card"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="additional-information"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="additional-information-list"]').exists()).toBe(true);
  });
});
