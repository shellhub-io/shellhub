import { createPinia, setActivePinia } from "pinia";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach } from "vitest";
import WelcomeFirstScreen from "@/components/Welcome/WelcomeFirstScreen.vue";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";

type WelcomeFirstScreenWrapper = VueWrapper<InstanceType<typeof WelcomeFirstScreen>>;

describe("Welcome First Screen", () => {
  let wrapper: WelcomeFirstScreenWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const authStore = useAuthStore();

  const authData = {
    token: "",
    username: "test",
    name: "test",
    tenantId: "fake-tenant-data",
    email: "test@test.com",
    id: "xxxxxxxx",
    role: "owner",
  };

  beforeEach(async () => {
    authStore.$patch(authData);

    wrapper = mount(WelcomeFirstScreen, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
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
    expect(wrapper.find('[data-test="welcome-first-screen-name"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="welcome-first-screen-text"]').exists()).toBe(true);
  });

  it("Renders the right namespace name", () => {
    expect(wrapper.vm.name).toEqual("test");
  });
});
