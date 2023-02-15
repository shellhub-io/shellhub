import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import BillingCancel from "../../../src/components/Billing/BillingCancel.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const role = ["owner", "operator"];

const hasAuthorization = {
  owner: true,
  operator: false,
};

const tests = [
  {
    description: "Button",
    props: {
      nextPaymentDue: 1234,
      currency: "USD",
    },
    data: {
      dialog: false,
    },
    template: {
      "cancel-btn": true,
      "billingCancel-dialog": false,
      "close-btn": false,
      "cancelDialog-btn": false,
    },
    templateText: {
      "cancel-btn": "Cancel",
    },
  },
  {
    description: "Dialog",
    props: {
      nextPaymentDue: 1234,
      currency: "BRL",
    },
    data: {
      dialog: false,
    },
    template: {
      "cancel-btn": true,
      "billingCancel-dialog": true,
      "close-btn": true,
      "cancelDialog-btn": true,
    },
    templateText: {
      "cancel-btn": "Cancel",
    },
  },
];

const store = (currentrole: string) => createStore({
  state: {
    currentrole,
  },
  getters: {
    "auth/role": (state) => state.currentrole,
  },
  actions: {
    "billing/cancelSubscription": vi.fn(),
    "billing/unsubscribe": vi.fn(),
    "devices/setDeviceChooserStatus": vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

describe("BillingCancel", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(() => {
          wrapper = mount(BillingCancel, {
            global: {
              plugins: [[store(currentrole), key], routes, vuetify],
            },
            props: {
              nextPaymentDue: test.props.nextPaymentDue,
              currency: test.props.currency,
            },
            shallow: true,
          });
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

        ///////
        // Data checking
        //////
        it("Data is defined", () => {
          expect(wrapper.vm.$data).toBeDefined();
        });
        it("Compare data with default value", () => {
          expect(wrapper.vm.dialog).toBe(test.data.dialog);
        });
        it("Compare the computed with the default value", () => {
          expect(wrapper.vm.hasAuthorization).toBe(
            hasAuthorization[currentrole],
          );
        });
      });
    });
  });
});
