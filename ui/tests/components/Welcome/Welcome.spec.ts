import { createPinia, setActivePinia } from "pinia";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach, afterEach } from "vitest";
import Welcome from "@/components/Welcome/Welcome.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

type WelcomeWrapper = VueWrapper<InstanceType<typeof Welcome>>;

describe("Welcome", () => {
  let wrapper: WelcomeWrapper;

  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(async () => {
    document.body.innerHTML = "";

    wrapper = mount(Welcome, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
      attachTo: document.body,
      props: {
        modelValue: true,
      },
    });

    await flushPromises();
  });

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount();
    }
    document.body.innerHTML = "";
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders key elements across steps with proper buttons", async () => {
    const dialog = new DOMWrapper(document.body);

    wrapper.vm.el = 1;
    await flushPromises();
    expect(dialog.text()).toContain("Step 1 of 4");
    expect(dialog.find('[data-test="welcome-first-screen"]').exists()).toBe(true);
    expect(dialog.find('[data-test="cancel-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="confirm-btn"]').exists()).toBe(true);

    wrapper.vm.el = 2;
    wrapper.vm.enable = false;
    await flushPromises();
    expect(dialog.find('[data-test="welcome-second-screen"]').exists()).toBe(true);
    expect(dialog.text()).toContain("Step 2 of 4");
    const step2Confirm = dialog.find('[data-test="confirm-btn"]');
    expect(step2Confirm.exists()).toBe(true);
    expect((step2Confirm.element as HTMLButtonElement).disabled).toBe(true);

    wrapper.vm.el = 3;
    await flushPromises();
    expect(dialog.find('[data-test="welcome-third-screen"]').exists()).toBe(true);
    expect(dialog.text()).toContain("Step 3 of 4");
    expect(dialog.find('[data-test="cancel-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="confirm-btn"]').exists()).toBe(true);

    wrapper.vm.el = 4;
    await flushPromises();
    expect(dialog.find('[data-test="welcome-fourth-screen"]').exists()).toBe(true);
    expect(dialog.text()).toContain("Step 4 of 4");
    expect(dialog.find('[data-test="confirm-btn"]').exists()).toBe(true);
  });

  it("Shows enabled 'Next' (confirm) button when the user sets up a device on step 2", async () => {
    const dialog = new DOMWrapper(document.body);
    wrapper.vm.el = 2;
    wrapper.vm.enable = true;
    await flushPromises();

    const confirm = dialog.find('[data-test="confirm-btn"]');
    expect(confirm.exists()).toBe(true);
    expect((confirm.element as HTMLButtonElement).disabled).toBe(false);
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
