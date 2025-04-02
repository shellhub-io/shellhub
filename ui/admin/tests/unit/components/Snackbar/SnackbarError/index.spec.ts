import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { shallowMount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import SnackbarError from "../../../../../src/components/Snackbar/SnackbarError.vue";
import { key } from "../../../../../src/store";
import routes from "../../../../../src/router";

type SnackbarErrorWrapper = VueWrapper<InstanceType<typeof SnackbarError>>

const snackbarError = true;
let typeMessage = "loading";
let mainContent = "dashboard";
const loadingMessage = `Loading the ${mainContent} has failed, please try again.`;
let actionMessage = `The ${mainContent} request has failed, please try again.`;
const defaultMessage = "The request has failed, please try again.";

const vuetify = createVuetify();

const store = createStore({
  state: {
    snackbarError,
  },
  getters: {
    "snackbar/snackbarError": (state) => state.snackbarError,
  },
  actions: {
    "snackbar/unsetShowStatusSnackbarError": vi.fn(),
  },
});

describe("Device Icon", () => {
  let wrapper: SnackbarErrorWrapper;

  beforeEach(() => {
    wrapper = shallowMount(SnackbarError, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
      propsData: { typeMessage, mainContent },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Process data in the computed", async () => {
    expect(wrapper.vm.snackbar).toEqual(snackbarError);
    expect(wrapper.vm.message).toEqual(loadingMessage);

    typeMessage = "action";
    mainContent = "deviceDelete";
    actionMessage = `The ${mainContent} request has failed, please try again.`;
    wrapper = shallowMount(SnackbarError, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
      propsData: { typeMessage, mainContent },
    });
    expect(wrapper.vm.message).toEqual(actionMessage);

    typeMessage = "default";
    wrapper = shallowMount(SnackbarError, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
      propsData: { typeMessage, mainContent },
    });
    expect(wrapper.vm.message).toEqual(defaultMessage);
  });
});
