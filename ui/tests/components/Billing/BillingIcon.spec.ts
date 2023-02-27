import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import BillingIcon from "../../../src/components/Billing/BillingIcon.vue";
import routes from "../../../src/router";

const iconName = "cc-amex";
const defaultIcon = "credit-card";

const cardIcon = {
  amex: "fa-cc-amex",
  dinersClub: "fa-cc-diners-club",
  discover: "fa-cc-discover",
  jcb: "fa-cc-jcb",
  mastercard: "fa-cc-mastercard",
  visa: "fa-cc-visa",
};

describe("BillingIcon", () => {
  let wrapper: VueWrapper<InstanceType<typeof BillingIcon>>;
  const vuetify = createVuetify();

  beforeEach(() => {
    wrapper = mount(BillingIcon, {
      global: {
        plugins: [routes, vuetify],
      },
      props: {
        iconName,
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

  it("Receive data in props", () => {
    expect(wrapper.vm.iconName).toBe(iconName);
  });
  it("Compare data with default value", () => {
    expect(wrapper.vm.cardIcon).toEqual(cardIcon);
  });

  //////
  // HTML validation
  //////

  it("Renders the template with data", () => {
    //////
    // In this case, the default icon is tested.
    //////

    wrapper = mount(BillingIcon, {
      global: {
        plugins: [routes, vuetify],
      },
      props: {
        iconName: defaultIcon,
      },
    });

    expect(wrapper.find('[data-test="default-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="type-icon"]').exists()).toBe(false);
  });

  //////
  // In this case, the other icons are tested.
  //////

  Object.keys(cardIcon).forEach((iconKey) => {
    wrapper = mount(BillingIcon, {
      global: {
        plugins: [routes, vuetify],
      },
      props: { iconName: iconKey },
    });

    expect(wrapper.find('[data-test="default-icon"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="type-icon"]').exists()).toBe(true);
  });
});
