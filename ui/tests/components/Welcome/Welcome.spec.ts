import { createPinia, setActivePinia } from "pinia";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach } from "vitest";
import Welcome from "@/components/Welcome/Welcome.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { key, store } from "@/store";

type WelcomeWrapper = VueWrapper<InstanceType<typeof Welcome>>;

describe("Welcome", () => {
  let wrapper: WelcomeWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(Welcome, {
      global: {
        plugins: [[store, key], vuetify, SnackbarPlugin],
      },
      props: {
        modelValue: true,
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the dialog open button and other key elements", async () => {
    const dialog = new DOMWrapper(document.body);
    wrapper.vm.el = 1;
    await flushPromises();
    expect(dialog.find('[data-test="step-counter"]').exists()).toBe(true);
    expect(dialog.find('[data-test="welcome-first-screen"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="first-click-btn"]').exists()).toBe(true);
    wrapper.vm.el = 2;
    await flushPromises();
    expect(dialog.find('[data-test="welcome-second-screen"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close2-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="back-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="waiting-message"]').exists()).toBe(true);
    wrapper.vm.el = 3;
    await flushPromises();
    expect(dialog.find('[data-test="welcome-third-screen"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close3-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="back2-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="accept-btn"]').exists()).toBe(true);
    wrapper.vm.el = 4;
    await flushPromises();
    expect(dialog.find('[data-test="welcome-fourth-screen"]').exists()).toBe(true);
    expect(dialog.find('[data-test="finish-btn"]').exists()).toBe(true);
  });
  it("Renders the next btn when the user setups a device", async () => {
    const dialog = new DOMWrapper(document.body);
    wrapper.vm.el = 2;
    wrapper.vm.enable = true;
    await flushPromises();
    expect(dialog.find('[data-test="next-btn"]').exists()).toBe(true);
  });
  it("Should go to the previous step when goToPreviousStep is called", async () => {
    wrapper.vm.el = 2;

    wrapper.vm.goToPreviousStep();
    await flushPromises();

    expect(wrapper.vm.el).toBe(1);
  });

  it("Should go to the next step when goToNextStep is called", async () => {
    wrapper.vm.el = 1;
    wrapper.vm.goToNextStep();

    await flushPromises();

    expect(wrapper.vm.el).toBe(2);
  });
});
