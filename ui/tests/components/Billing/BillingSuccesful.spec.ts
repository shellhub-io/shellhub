import { mount } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it } from "vitest";
import BillingSuccessful from "@/components/Billing/BillingSuccessful.vue";

describe("Billing Successful Card", () => {
  const wrapper = mount(BillingSuccessful, { global: { plugins: [createVuetify()] } });
  it("renders the correct text", () => {
    expect(wrapper.findComponent('[data-test="title"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="green-cloud-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="h4-letter"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="p-letter"]').exists()).toBe(true);
  });
});
