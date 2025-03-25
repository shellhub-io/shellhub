import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { shallowMount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import Snackbar from "../../../../src/components/Snackbar/Snackbar.vue";
import { key } from "../../../../src/store";
import router from "../../../../src/router";

const snackbarMessageAndContentType = {
  typeMessage: "",
  typeContent: "",
};
const store = createStore({
  state: {
    snackbarMessageAndContentType,
  },
  getters: {
    "snackbar/snackbarMessageAndContentType": (state) => state.snackbarMessageAndContentType,
  },
  actions: {},
});

describe("Device Icon", () => {
  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = shallowMount(Snackbar, {
      global: {
        plugins: [[store, key], vuetify, router],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });
  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it("Process data in the computed", () => {
    expect(wrapper.vm.message).toEqual(snackbarMessageAndContentType);
  });
});
