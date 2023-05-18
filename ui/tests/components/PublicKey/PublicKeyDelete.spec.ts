import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import PublicKeyDelete from "../../../src/components/PublicKeys/PublicKeyDelete.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const tests = [
  {
    description: "Dialog closed",
    props: {
      fingerprint: "b7:25:f8",
      notHasAuthorization: false,
    },
    template: {
      "remove-icon": true,
      "remove-title": true,
      "publicKeyDelete-card": false,
    },
    templateText: {
      "remove-title": "Remove",
    },
  },
  {
    description: "Dialog opened",
    props: {
      fingerprint: "b7:25:f8",
      notHasAuthorization: false,
    },
    template: {
      "remove-icon": true,
      "remove-title": true,
      "publicKeyDelete-card": true,
      "text-title": true,
      "text-text": true,
      "close-btn": true,
      "remove-btn": true,
    },
    templateText: {
      "remove-title": "Remove",
      "text-title": "Are you sure?",
      "text-text": "You are about to remove this public key.",
      "close-btn": "Close",
      "remove-btn": "Remove",
    },
  },
];

const store = createStore({
  state: {},
  getters: {},
  actions: {
    "publicKeys/remove": vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});
describe("PublicKeyDelete", () => {
  let wrapper: VueWrapper<InstanceType<typeof PublicKeyDelete>>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        const wrapper = mount(PublicKeyDelete, {
          global: {
            plugins: [[store, key], routes, vuetify],
          },
          props: {
            ...test.props,
          },
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

      ///////
      // HTML validation
      //////

      it("Renders the correct HTML", () => {
        expect(wrapper.find('[data-test="remove-icon"]').exists()).toBeTruthy();
        expect(
          wrapper.find('[data-test="remove-title"]').exists(),
        ).toBeTruthy();
      });

      // todo
    });
  });
});
