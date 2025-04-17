import { createVuetify } from "vuetify";
import { shallowMount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import router from "@admin/router";
import useAuthStore from "@admin/store/modules/auth";
import useSnackbarStore from "@admin/store/modules/snackbar";
import useLayoutStore from "@admin/store/modules/layout";
import useLicenseStore from "@admin/store/modules/license";
import SnackbarComponent from "../../../../src/components/Snackbar/Snackbar.vue";
import Login from "../../../../src/views/Login.vue";

type LoginWrapper = VueWrapper<InstanceType<typeof Login>>;

describe("Login", () => {
  let wrapper: LoginWrapper;

  beforeEach(() => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const authStore = useAuthStore();
    authStore.login = vi.fn();

    const layoutStore = useLayoutStore();
    vi.spyOn(layoutStore, "getLayout", "get").mockReturnValue("simpleLayout");

    const licenseStore = useLicenseStore();
    licenseStore.get = vi.fn();

    const snackbarStore = useSnackbarStore();
    snackbarStore.showSnackbarErrorDefault = vi.fn();
    snackbarStore.showSnackbarErrorAction = vi.fn();

    const vuetify = createVuetify();

    wrapper = shallowMount(Login, {
      global: {
        plugins: [pinia, vuetify, router],
        components: { SnackbarComponent },
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBeTruthy();
  });

  it("Renders the component", () => {
    router.push("/login");
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Compare data with default value", () => {
    expect(wrapper.vm.username).toEqual("");
    expect(wrapper.vm.password).toEqual("");
    expect(wrapper.vm.usernameError).toEqual(undefined);
    expect(wrapper.vm.passwordError).toEqual(undefined);
  });
});
