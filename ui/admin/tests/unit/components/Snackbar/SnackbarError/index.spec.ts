import { createVuetify } from "vuetify";
import { shallowMount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useSnackbarStore from "@admin/store/modules/snackbar";
import { SnackbarPlugin } from "@/plugins/snackbar";
import SnackbarError from "../../../../../src/components/Snackbar/SnackbarError.vue";
import routes from "../../../../../src/router";

type SnackbarErrorWrapper = VueWrapper<InstanceType<typeof SnackbarError>>

const mainContent = "dashboard";
const snackbarError = true;
const defaultMessage = "The request has failed, please try again.";

describe("SnackbarError", () => {
  let wrapper: SnackbarErrorWrapper;

  const vuetify = createVuetify();

  beforeEach(() => {
    setActivePinia(createPinia());
  });

  const mountComponent = (typeMessage: string, mainContent: string) => {
    const snackbarStore = useSnackbarStore();
    snackbarStore.snackbarError = snackbarError;

    return shallowMount(SnackbarError, {
      global: {
        plugins: [vuetify, routes, SnackbarPlugin],
      },
      props: {
        typeMessage,
        mainContent,
      },
    });
  };

  it("Is a Vue instance", () => {
    wrapper = mountComponent("loading", mainContent);
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    wrapper = mountComponent("loading", mainContent);
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Processes computed message for 'loading'", () => {
    wrapper = mountComponent("loading", mainContent);
    expect(wrapper.vm.snackbar).toEqual(true);
    expect(wrapper.vm.message).toEqual(`Loading the ${mainContent} has failed, please try again.`);
  });

  it("Processes computed message for 'action'", () => {
    wrapper = mountComponent("action", "deviceDelete");
    expect(wrapper.vm.message).toEqual("The deviceDelete request has failed, please try again.");
  });

  it("Processes computed message for 'default'", () => {
    wrapper = mountComponent("default", mainContent);
    expect(wrapper.vm.message).toEqual(defaultMessage);
  });

  it("Processes computed message for 'licenseRequired'", () => {
    wrapper = mountComponent("licenseRequired", "license feature");
    expect(wrapper.vm.message).toEqual("The license feature request has failed, license required.");
  });

  it("Processes computed message for 'custom'", () => {
    wrapper = mountComponent("custom", "Custom error message here");
    expect(wrapper.vm.message).toEqual("Custom error message here");
  });
});
