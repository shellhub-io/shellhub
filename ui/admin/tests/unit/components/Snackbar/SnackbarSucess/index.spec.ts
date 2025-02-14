import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { shallowMount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import SnackbarSucess from "../../../../../src/components/Snackbar/SnackbarSucess.vue";
import { key } from "../../../../../src/store";
import routes from "../../../../../src/router";

const snackbarSuccess = true;
let typeMessage = "action";
const mainContent = "renaming device";
const actionMessage = `The ${mainContent} has succeeded.`;
const defaultMessage = "The request has succeeded.";
const vuetify = createVuetify();

const store = createStore({
  state: {
    snackbarSuccess,
  },
  getters: {
    "snackbar/snackbarSuccess": (state) => state.snackbarSuccess,
  },
  actions: {
    "snackbar/unsetShowStatusSnackbarSuccess": () => vi.fn(),
  },
});

describe("Device Icon", () => {
  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    wrapper = shallowMount(SnackbarSucess, {
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

  it("Renders the correct message", () => {
    expect(wrapper.vm.snackbar).toEqual(snackbarSuccess);
    expect(wrapper.vm.message).toEqual(actionMessage);
    typeMessage = "default";
    wrapper = shallowMount(SnackbarSucess, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
      propsData: { typeMessage, mainContent },
    });
    expect(wrapper.vm.message).toEqual(defaultMessage);
  });
});
