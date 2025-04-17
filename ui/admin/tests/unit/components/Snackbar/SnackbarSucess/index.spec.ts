import { createVuetify } from "vuetify";
import { shallowMount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useSnackbarStore from "@admin/store/modules/snackbar";
import SnackbarSucess from "../../../../../src/components/Snackbar/SnackbarSucess.vue";
import routes from "../../../../../src/router";

type SnackbarSucessWrapper = VueWrapper<InstanceType<typeof SnackbarSucess>>;

describe("SnackbarSucess", () => {
  let wrapper: SnackbarSucessWrapper;
  const vuetify = createVuetify();

  const mountComponent = (typeMessage: string, mainContent: string) => {
    const store = useSnackbarStore();
    store.snackbarSuccess = true;

    return shallowMount(SnackbarSucess, {
      global: {
        plugins: [vuetify, routes],
      },
      props: {
        typeMessage,
        mainContent,
      },
    });
  };

  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("Is a Vue instance", () => {
    wrapper = mountComponent("action", "renaming device");
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    wrapper = mountComponent("action", "renaming device");
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Processes message for 'action'", () => {
    wrapper = mountComponent("action", "renaming device");
    expect(wrapper.vm.snackbar).toBe(true);
    expect(wrapper.vm.message).toBe("The renaming device has succeeded.");
  });

  it("Processes message for 'default'", () => {
    wrapper = mountComponent("default", "renaming device");
    expect(wrapper.vm.message).toBe("The request has succeeded.");
  });

  it("Processes message for 'no-content'", () => {
    wrapper = mountComponent("no-content", "export");
    expect(wrapper.vm.message).toBe("There is no content to export");
  });
});
