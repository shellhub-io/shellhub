import { setActivePinia, createPinia } from "pinia";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach, vi } from "vitest";
import WelcomeSecondScreen from "@/components/Welcome/WelcomeSecondScreen.vue";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

type WelcomeSecondScreenWrapper = VueWrapper<InstanceType<typeof WelcomeSecondScreen>>;

const mockSnackbar = {
  showInfo: vi.fn(),
};

describe("Welcome Second Screen", () => {
  let wrapper: WelcomeSecondScreenWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(WelcomeSecondScreen, {
      global: {
        plugins: [vuetify],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: {
        command: "",
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the components", async () => {
    expect(wrapper.find('[data-test="welcome-second-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="welcome-second-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="welcome-second-run-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="command-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="welcome-second-link-docs"]').exists()).toBe(true);
  });
});
