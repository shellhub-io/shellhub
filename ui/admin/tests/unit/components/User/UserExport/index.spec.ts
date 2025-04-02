import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import UserExport from "../../../../../src/components/User/UserExport.vue";
import { key } from "../../../../../src/store";

type UserExportWrapper = VueWrapper<InstanceType<typeof UserExport>>;

const store = createStore({
  state: {},
  getters: {},
  actions: {
    "users/setFilterUsers": () => vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

describe("User Export", () => {
  let wrapper: UserExportWrapper;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(UserExport, {
      global: {
        plugins: [[store, key], vuetify],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Compare data with default value", () => {
    expect(wrapper.vm.gtNumberOfNamespaces).toEqual(0);
    expect(wrapper.vm.eqNumberOfNamespaces).toEqual(0);
    expect(wrapper.vm.dialog).toEqual(false);
    expect(wrapper.vm.selected).toEqual("moreThan");
  });
});
