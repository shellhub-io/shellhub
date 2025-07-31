import { createPinia, setActivePinia } from "pinia";
import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach } from "vitest";
import { store, key } from "@/store";
import PaywallChat from "@/components/User/PaywallChat.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("PaywallChat", () => {
  let wrapper: VueWrapper<InstanceType<typeof PaywallChat>>;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(PaywallChat, {
      global: {
        plugins: [[store, key], vuetify, SnackbarPlugin],
      },
      attachTo: document.body,
    });
  });

  it("is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("renders dialog elements when opened", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();

    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="card-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="icon-chat"]').exists()).toBe(true);
    expect(dialog.find('[data-test="upgrade-heading"]').exists()).toBe(true);
    expect(dialog.find('[data-test="upgrade-description"]').exists()).toBe(true);
    expect(dialog.find('[data-test="upgrade-description2"]').exists()).toBe(true);
    expect(dialog.find('[data-test="link-anchor"]').exists()).toBe(true);
    expect(dialog.find('[data-test="card-actions"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="upgrade-btn"]').exists()).toBe(true);
  });

  it("ensures the upgrade button has correct href", () => {
    wrapper.vm.showDialog = true;
    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="upgrade-btn"]').attributes("href")).toBe("https://www.shellhub.io/pricing");
  });
});
