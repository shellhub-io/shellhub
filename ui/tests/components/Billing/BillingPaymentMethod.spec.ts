import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import BillingPaymentMethod from "../../../src/components/Billing/BillingPaymentMethod.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const typeOperation = "subscription";
const hasSpinner = false;
const stats = { registered_devices: 36 };

const store = createStore({
  state: {
    hasSpinner,
    stats,
  },
  getters: {
    "spinner/getStatus": (state) => state.hasSpinner,
    "stats/stats": (state) => state.stats,
  },
  actions: {
    "stats/get": vi.fn(),
    "billing/subscritionPaymentMethod": vi.fn(),
    "billing/updatePaymentMethod": vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

///////
// In this case, it's testing the button rendering.
///////
describe("BillingDialogPaymentMethod", () => {
  let wrapper: VueWrapper<InstanceType<typeof BillingPaymentMethod>>;
  const vuetify = createVuetify();
  ///////
  // In this case, it's testing the button rendering.
  ///////
  describe("Button", () => {
    beforeEach(() => {
      wrapper = mount(BillingPaymentMethod, {
        global: {
          plugins: [[store, key], routes, vuetify],
        },
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
      expect(wrapper.vm.dialog).toEqual(false);
      expect(wrapper.vm.card).toEqual(undefined);
      expect(wrapper.vm.elementError).toEqual("");
      expect(wrapper.vm.elms).toEqual(undefined);
      expect(wrapper.vm.lockButton).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it("Renders the template with data", async () => {
      expect(wrapper.find('[data-test="show-btn"]').exists()).toBe(true);
    });
  });

  ///////
  // In this case, it's testing the subscription.
  ///////

  describe("Dialog Subscription", () => {
    beforeEach(() => {
      wrapper = mount(BillingPaymentMethod, {
        global: {
          plugins: [[store, key], routes, vuetify],
        },
        props: { typeOperation },
        shallow: true,
      });

      wrapper.vm.dialog = true;
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

    it("Receive data in props", () => {
      expect(wrapper.vm.typeOperation).toBe(typeOperation);
    });

    it("Compare data with default value", () => {
      expect(wrapper.vm.dialog).toEqual(true);
      expect(wrapper.vm.card).toEqual(undefined);
      expect(wrapper.vm.elementError).toEqual("");
      expect(wrapper.vm.elms).toEqual(undefined);
      expect(wrapper.vm.lockButton).toEqual(false);
    });

    it("Process data in methods", () => {
      const priceTable = {
        22: 55.92,
        123: 319.25,
        171: 426.7,
      };
      Reflect.ownKeys(priceTable).forEach((k) => {
        expect(wrapper.vm.priceEstimator(parseInt(k as string, 10))).toContain(
          priceTable[k],
        );
      });
    });

    //////
    // HTML validation
    //////

    it("Renders the template with data", async () => {
      expect(wrapper.find('[data-test="show-btn"]').exists()).toBe(true);
      // TODO
    });
  });

  ///////
  // In this case, it's testing the update subscription.
  ///////

  describe("Dialog Update", () => {
    const typeOperationUpdate = "update";

    beforeEach(() => {
      wrapper = mount(BillingPaymentMethod, {
        global: {
          plugins: [[store, key], routes, vuetify],
        },
        props: { typeOperation: typeOperationUpdate },
        shallow: true,
      });

      wrapper.vm.dialog = true;
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

    /// //
    // Data and Props checking
    /// /

    it("Receive data in props", () => {
      expect(wrapper.vm.typeOperation).toBe(typeOperationUpdate);
    });
    it("Compare data with default value", () => {
      expect(wrapper.vm.dialog).toEqual(true);
      expect(wrapper.vm.card).toEqual(undefined);
      expect(wrapper.vm.elementError).toEqual("");
      expect(wrapper.vm.elms).toEqual(undefined);
    });

    //////
    // HTML validation
    //////

    it("Renders the template with data", async () => {
      expect(wrapper.find('[data-test="show-btn"]').exists()).toBe(true);
      // TODO
    });
  });
});
