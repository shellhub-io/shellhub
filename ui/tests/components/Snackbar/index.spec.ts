import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import Snackbar from "../../../src/components/Snackbar/Snackbar.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

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
  let wrapper: VueWrapper<InstanceType<typeof Snackbar>>;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(Snackbar, {
      global: {
        plugins: [[store, key], vuetify, routes],
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
