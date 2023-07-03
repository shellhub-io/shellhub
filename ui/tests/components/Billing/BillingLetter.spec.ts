import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach } from "vitest";
import BillingLetter from "@/components/Billing/BillingLetter.vue";

describe("Billing Letter", () => {
  let wrapper: VueWrapper<InstanceType<typeof BillingLetter>>;
  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(BillingLetter, {
      global: {
        plugins: [vuetify],
      },
    });
  });
  it("renders the correct text", async () => {
    expect(wrapper.findComponent('[data-test="title"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="text"]').exists()).toBe(true);
  });
});
