import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import BillingWarning from "@/components/Billing/BillingWarning.vue";
import { router } from "@/router";
import useAuthStore from "@/store/modules/auth";

describe("BillingWarning", () => {
  let wrapper: VueWrapper<InstanceType<typeof BillingWarning>>;
  let dialog: DOMWrapper<HTMLElement>;
  const vuetify = createVuetify();

  setActivePinia(createPinia());
  const authStore = useAuthStore();
  authStore.role = "owner";

  beforeEach(() => {
    wrapper = mount(BillingWarning, {
      global: { plugins: [router, vuetify] },
      props: { modelValue: true },
    });
    dialog = new DOMWrapper(document.body);
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("should render dialog when user has authorization", () => {
    expect(dialog.find('[data-test="billing-warning-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="card-title"]').text()).toBe("Maximum Device Limit Reached");
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="go-to-billing-btn"]').exists()).toBe(true);
  });

  it("should not render dialog when user lacks authorization", () => {
    wrapper.unmount();
    authStore.role = "observer";
    wrapper = mount(BillingWarning, {
      global: { plugins: [router, vuetify] },
      props: { modelValue: true },
    });
    dialog = new DOMWrapper(document.body);

    expect(dialog.find('[data-test="billing-warning-dialog"]').exists()).toBe(false);
  });
});
