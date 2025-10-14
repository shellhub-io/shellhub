import { createPinia, setActivePinia } from "pinia";
import { DOMWrapper, flushPromises, mount } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it } from "vitest";
import Welcome from "@/components/Welcome/Welcome.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("Welcome", () => {
  setActivePinia(createPinia());

  const wrapper = mount(Welcome, {
    global: { plugins: [createVuetify(), SnackbarPlugin] },
    props: { modelValue: true },
  });

  it("Renders the component", () => {
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Enables 'Next' (confirm) button when the user sets up a device on step 2", async () => {
    const dialog = new DOMWrapper(document.body);
    wrapper.vm.currentStep = 2;
    wrapper.vm.hasDeviceDetected = true;
    await flushPromises();

    const confirmButton = dialog.find('[data-test="confirm-btn"]');
    expect(confirmButton.exists()).toBe(true);
    expect((confirmButton.element as HTMLButtonElement).disabled).toBe(false);
  });
});
