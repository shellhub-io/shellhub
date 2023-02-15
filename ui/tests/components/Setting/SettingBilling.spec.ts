import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import SettingBilling from "../../../src/components/Setting/SettingBilling.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";
import { envVariables } from "../../../src/envVariables";

const role = ["owner", "operator"];

const hasAuthorization = {
  owner: true,
  operator: false,
};

const stripeData = {
  latest_invoice: { amount_due: 0, amount_paid: 0 },
  upcoming_invoice: { amount_due: 0, amount_paid: 0 },
  product_description: "Premium usage",
};
const warningTitle = "Payment failed";

const warningMessage = `Please update your payment method
by adding a new card, or attempt to pay failed latest
invoices through url`;

const infoData = {
  info: {
    periodEnd: "2021-12-24T18:16:21Z",
    description: "Shellhub",
    latestPaymentDue: 0,
    latestPaymentPaid: 0,
    nextPaymentDue: 0,
    nextPaymenPaid: 0,
  },

  card: {
    brand: "visa",
    expYear: 2024,
    default: true,
    expMonth: 4,
    last4: "4042",
    id: "pm_1JzQ80KJsksFHO6pREJA5TrK",
  },
  cards: [
    {
      brand: "visa",
      expYear: 2024,
      default: true,
      expMonth: 4,
      last4: "4042",
      id: "pm_1JzQ80KJsksFHO6pREJA5TrK",
    },
    {
      brand: "visa",
      expYear: 2028,
      default: false,
      expMonth: 4,
      last4: "4042",
      id: "pm_1JzQ80KJsksFHO6pREJA5TrG",
    },
    {
      brand: "visa",
      expYear: 2029,
      default: false,
      expMonth: 4,
      last4: "4042",
      id: "pm_1JzQ80KJsksFHO6pREJA5TrF",
    },
  ],
  invoices: [],
  warning: false,
};

const info2 = {
  periodEnd: "2021-12-24T18:16:21Z",
  description: "Shellhub",
  latestPaymentDue: 0,
  latestPaymentPaid: 0,
  nextPaymentDue: 0,
  nextPaymenPaid: 0,
};

const card2 = {
  brand: "visa",
  expYear: 2024,
  default: true,
  expMonth: 4,
  last4: "4042",
  id: "pm_123",
};

const tests = [
  {
    description: "Create subscription",
    computed: {
      active: false,
      state: "inactive",
    },
    data: {
      renderData: false,
      action: "subscribe",
      warningTitle,
      warningMessage,
    },
    instance: {
      active: false,
      state: "inactive",
      current_period_end: 0,
      customer_id: "",
      subscription_id: "",
      payment_method_id: "",
    },
    infoData,
    template: {
      "subscriptionPaymentMethod-component": true,
      "freePlan-div": true,
      "premiumPlan-div": false,
      "subscriptionActive-div": false,
      "updatePaymentMethod-component": false,
      "paymentMethods-component": false,
      "invoiceList-component": false,
      "cancel-div": false,
      "warning-div": false,
    },
  },
  {
    description: "Pending request",
    owner: true,
    computed: {
      active: true,
      state: "pending",
    },
    data: {
      renderData: true,
      action: "subscribe",
      warningTitle,
      warningMessage,
    },
    instance: {
      active: true,
      state: "pending",
      current_period_end: 0,
      customer_id: "cus_123",
      subscription_id: "sub_123",
      payment_method_id: "pm_123",
    },
    infoData,
    template: {
      "subscriptionPaymentMethod-component": false,
      "pendingRetrial-div": true,
      "freePlan-div": false,
      "premiumPlan-div": false,
      "subscriptionActive-div": false,
      "updatePaymentMethod-component": false,
      "paymentMethods-component": false,
      "invoiceList-component": false,
      "cancel-div": false,
      "activeLoading-div": false,
      "warning-div": false,
    },
  },
  {
    description: "Premium usage",
    computed: {
      active: true,
      state: "processed",
    },
    data: {
      renderData: true,
      action: "subscribe",
      warningTitle,
      warningMessage,
    },
    instance: {
      active: true,
      state: "processed",
      current_period_end: 0,
      customer_id: "cus_123",
      subscription_id: "sub_123",
      payment_method_id: "pm_123",
      info: info2,
      card: card2,
    },
    infoData,
    template: {
      "subscriptionPaymentMethod-component": false,
      "freePlan-div": false,
      "premiumPlan-div": true,
      "subscriptionActive-div": true,
      "updatePaymentMethod-component": true,
      "invoiceList-component": true,
      "paymentMethods-component": true,
      "cancel-div": true,
      "activeLoading-div": false,
      "warning-div": false,
    },
  },
  {
    description: "Premium usage - warning",
    computed: {
      active: true,
      state: "processed",
    },
    data: {
      renderData: true,
      action: "subscribe",
      warningTitle,
      warningMessage,
    },
    infoData: { ...infoData, warning: true },
    instance: {
      active: true,
      state: "processed",
      current_period_end: 0,
      customer_id: "cus_123",
      subscription_id: "sub_123",
      payment_method_id: "pm_123",
      info: info2,
      card: card2,
    },
    template: {
      "subscriptionPaymentMethod-component": false,
      "freePlan-div": false,
      "premiumPlan-div": true,
      "subscriptionActive-div": true,
      "updatePaymentMethod-component": true,
      "invoiceList-component": true,
      "paymentMethods-component": true,
      "cancel-div": true,
      "activeLoading-div": false,
      "warning-div": true,
    },
  },
];

const store = (billing: any, currentrole: any, info: any) => {
  return createStore({
    state: {
      billing,
      currentrole,
      info,
    },
    getters: {
      "billing/active": (state) => state.billing.active || false,
      "billing/status": (state) => state.billing.state || "inactive",
      "billing/get": (state) => state.billing,
      "auth/role": (state) => state.currentrole,
      "billing/getBillInfoData": (state) => state.info,
      "billing/getInvoices": (state) => state.info.invoices,
    },
    actions: {
      "billing/getSubscription": () => stripeData,
      "namespaces/get": () => {},
      "snackbar/showSnackbarSuccessAction": () => {},
      "snackbar/showSnackbarErrorAction": () => {},
      "snackbar/showSnackbarErrorDefault": () => {},
    },
  });
};

describe("SettingBilling", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(() => {
          wrapper = mount(SettingBilling, {
            global: {
              plugins: [
                [store(test.instance, currentrole, test.infoData), key],
                routes,
                vuetify,
              ],
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
            shallow: true,
          });

          if (test.data.renderData) {
            wrapper.vm.renderData = true;
          }

        });

        ///////
        // Component Rendering
        //////

        it("Is a Vue instance", () => {
          expect(wrapper).toBeTruthy();
        });
        it("Renders the component", () => {
          expect(wrapper.html()).toMatchSnapshot();
        });

        ///////s
        // Data checking
        //////
        it("Data is defined", () => {
          expect(wrapper.vm.$data).toBeDefined();
        });
        it("Compare data with default value", () => {
          expect(wrapper.vm.renderData).toBe(test.data.renderData);
          expect(wrapper.vm.warningTitle).toBe(test.data.warningTitle);
        });
        it('Process data in the computed', () => {
          expect(wrapper.vm.active).toBe(test.computed.active);
          expect(wrapper.vm.state).toBe(test.computed.state);
        });
      });
    });
  });
});
