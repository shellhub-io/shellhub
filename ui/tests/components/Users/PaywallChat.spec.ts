import { createPinia, setActivePinia } from "pinia";
import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach, vi } from "vitest";
import PaywallChat from "@/components/User/PaywallChat.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("PaywallChat", () => {
  let wrapper: VueWrapper<InstanceType<typeof PaywallChat>>;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(() => {
    wrapper = mount(PaywallChat, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
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

    expect(dialog.text()).toContain("Upgrade to have access to chat support!");
    expect(dialog.find('[data-test="upgrade-description-1"]').exists()).toBe(true);
    expect(dialog.find('[data-test="upgrade-description-2"]').exists()).toBe(true);
    expect(dialog.find('[data-test="link-anchor"]').exists()).toBe(true);

    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="upgrade-btn"]').exists()).toBe(true);

    expect(dialog.html()).toContain("mdi-chat-question");
  });

  it("clicking Upgrade triggers redirect to pricing", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();

    const openSpy = vi.spyOn(window, "open").mockImplementation(() => null);
    const dialog = new DOMWrapper(document.body);

    await dialog.find('[data-test="upgrade-btn"]').trigger("click");
    expect(openSpy).toHaveBeenCalledWith(
      "https://www.shellhub.io/pricing",
      "_blank",
      "noopener,noreferrer",
    );

    openSpy.mockRestore();
  });
});
