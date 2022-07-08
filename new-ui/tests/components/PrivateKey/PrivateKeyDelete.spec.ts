import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import PrivateKeyDelete from "../../../src/components/PrivateKeys/PrivateKeyDelete.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const tests = [
  {
    description: "Dialog closed",
    props: {
      fingerprint: "b7:25:f8",
    },
    data: {
      showDialog: false,
    },
    template: {
      "remove-icon": true,
      "remove-title": true,
      "privateKeyDelete-card": false,
    },
    templateText: {
      "remove-title": "Remove",
    },
  },
  {
    description: "Dialog opened",
    props: {
      fingerprint: "b7:25:f8",
    },
    data: {
      showDialog: true,
    },
    template: {
      "remove-icon": true,
      "remove-title": true,
      "privateKeyDelete-card": true,
      "text-title": true,
      "text-text": true,
      "close-btn": true,
      "remove-btn": true,
    },
    templateText: {
      "remove-title": "Remove",
      "text-title": "Are you sure?",
      "text-text": "You are about to remove this private key.",
      "close-btn": "Close",
      "remove-btn": "Remove",
    },
  },
];

const store = createStore({
  state: {},
  getters: {},
  actions: {
    "privatekeys/remove": vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

describe("PrivateKeyDelete", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(PrivateKeyDelete, {
          global: {
            plugins: [[store, key], routes, vuetify],
          },
          props: {
            fingerprint: test.props.fingerprint,
          },
          setup() {
            return {
              showDialog: test.data.showDialog,
            };
          },
          shallow: true,
        });
      });

      ///////
      // Component Rendering
      //////

      it("Is a Vue instance", () => {
        expect(wrapper).toBeTruthy();
      });
      it("Renders the component", () => {
        expect(wrapper.html()).toMatchSnapshot();
      });

      ///////
      // Data checking
      //////
      it("Data is defined", () => {
        expect(wrapper.vm.$data).toBeDefined();
      });

      it("Compare data with default value", () => {
        expect(wrapper.vm.showDialog).toBe(test.data.showDialog);
      });

      it("Compare props with default value", () => {
        expect(wrapper.vm.fingerprint).toBe(test.props.fingerprint);
      });

      //////
      // HTML validation
      //////

      // Todo
    });
  });
});
