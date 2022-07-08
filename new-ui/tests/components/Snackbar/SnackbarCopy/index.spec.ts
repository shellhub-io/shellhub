import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import SnackbarCopy from "../../../../src/components/Snackbar/SnackbarCopy.vue";
import { key } from "../../../../src/store";
import routes from "../../../../src/router";

const mainContent = "Command";
const snackbarCopy = true;
const message = `${mainContent} copied to clipboard.`;

const vuetify = createVuetify();

const store = createStore({
  state: {
    snackbarCopy,
  },
  getters: {
    "snackbar/snackbarCopy": (state) => state.snackbarCopy,
  },
  actions: {
    "snackbar/unsetShowStatusSnackbarCopy": vi.fn(),
  },
});

describe("Device Icon", () => {
  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    wrapper = mount(SnackbarCopy, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
      props: { mainContent },
      shallow: true,
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Receive data in props", () => {
    expect(wrapper.vm.mainContent).toEqual(mainContent);
  });
  it("Process data in the computed", async () => {
    expect(wrapper.vm.snackbar).toEqual(snackbarCopy);
    expect(wrapper.vm.message).toEqual(message);
  });
});
