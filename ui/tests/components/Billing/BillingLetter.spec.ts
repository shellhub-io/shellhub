import { mount } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it } from "vitest";
import BillingLetter from "@/components/Billing/BillingLetter.vue";

describe("Billing Letter", () => {
  const wrapper = mount(BillingLetter, {
    global: { plugins: [createVuetify()] },
  });

  it("renders the correct text", async () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
