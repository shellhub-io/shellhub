import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach } from "vitest";
import BillingSuccessful from "@/components/Billing/BillingSuccessful.vue";

describe("Billing Sucessful Card", () => {
  let wrapper: VueWrapper<InstanceType<typeof BillingSuccessful>>;
  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(BillingSuccessful, {
      global: {
        plugins: [vuetify],
      },
    });
  });
  it("renders the correct text", async () => {
    expect(wrapper.findComponent('[data-test="title"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="green-cloud-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="h4-letter"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="p-letter"]').exists()).toBe(true);
  });
});
