import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import BillingPaymentList from "../../../src/components/Billing/BillingPaymentList.vue";
import { createStore } from "vuex";
import { store, key } from "../../../src/store";
import routes from "../../../src/router";

const pms = [
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
];

const headers = [
  {
    text: "Brand",
    value: "brand",
    align: "center",
    sortable: false,
  },
  {
    text: "Exp. Date",
    value: "expdate",
    align: "center",
    sortable: false,
  },
  {
    text: "Ends with",
    value: "last4",
    align: "center",
    sortable: false,
  },
  {
    text: "Actions",
    value: "actions",
    align: "center",
    sortable: false,
  },
];

describe("BillingPaymentList", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  beforeEach(() => {
    wrapper = mount(BillingPaymentList, {
      global: {
        plugins: [[store, key],routes, vuetify],
      },
      props: { cards: pms },
      shallow: false,
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
  it('Compares data with default value', () => {
    expect(wrapper.vm.headers).toStrictEqual(headers);
  });
  it('Process data in the computed', () => {
    expect(wrapper.vm.paymentList).toStrictEqual(pms);
  });
  it('Process data in props', () => {
    expect(wrapper.props("cards")).toStrictEqual(pms);
  });

  ///////
  // HTML validation
  //////

  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="dataTable-field"]').exists()).toBe(true);
  });
});
