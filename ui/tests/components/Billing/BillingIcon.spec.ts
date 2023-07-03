import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach } from "vitest";
import BillingIcon from "@/components/Billing/BillingIcon.vue";

describe("Billing Icon", () => {
  let wrapper: VueWrapper<InstanceType<typeof BillingIcon>>;
  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(BillingIcon, {
      global: {
        plugins: [vuetify],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("renders the default icon", () => {
    expect(wrapper.findComponent('[data-test="default-icon"]').exists()).toBe(true);
  });

  it("renders the responsive icon", () => {
    expect(wrapper.findComponent('[data-test="type-icon"]').exists()).toBe(false);
  });
});
