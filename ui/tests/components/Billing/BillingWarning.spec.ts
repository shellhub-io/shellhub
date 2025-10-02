import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import BillingWarning from "@/components/Billing/BillingWarning.vue";
import { router } from "@/router";
import useAuthStore from "@/store/modules/auth";

describe("BillingWarning", () => {
  const vuetify = createVuetify();
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  authStore.role = "owner";

  const wrapper = mount(BillingWarning, {
    global: { plugins: [router, vuetify] },
    props: { modelValue: true },
  });

  it("should render dialog when user has authorization", () => {
    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="billing-warning-dialog"]').exists()).toBe(true);
    expect(dialog.text()).toContain("Maximum Device Limit Reached");
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="go-to-billing-btn"]').exists()).toBe(true);
  });

  it("should not render dialog when user lacks authorization", () => {
    wrapper.unmount();
    authStore.role = "observer";
    mount(BillingWarning, {
      global: { plugins: [router, vuetify] },
      props: { modelValue: true },
    });
    const dialog = new DOMWrapper(document.body);

    expect(dialog.find('[data-test="billing-warning-dialog"]').exists()).toBe(false);
  });
});
