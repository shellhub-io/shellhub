import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import PrivateKeyEdit from "../../../src/components/PrivateKeys/PrivateKeyEdit.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const keyObject = {
  name: "ShellHub",
  data: "",
};

const privateKey = {
  name: "",
  data: "",
};

const tests = [
  {
    description: "Dialog closed",
    props: {
      keyObject,
      show: false,
    },
    data: {
      privateKey,
      supportedKeys:
        "Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.",
    },
    template: {
      "edit-icon": true,
      "edit-title": true,
      "privateKeyFormDialog-card": false,
    },
    templateText: {
      "edit-title": "Edit",
    },
  },
  {
    description: "Dialog opened",
    props: {
      keyObject,
      show: true,
    },
    data: {
      privateKey,
      supportedKeys:
        "Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.",
    },
    template: {
      "edit-icon": true,
      "edit-title": true,
      "privateKeyFormDialog-card": true,
      "text-title": true,
      "name-field": true,
      "data-field": true,
      "cancel-btn": true,
      "edit-btn": true,
    },
    templateText: {
      "edit-title": "Edit",
      "cancel-btn": "Cancel",
      "edit-btn": "Edit",
    },
  },
];

const store = createStore({
  state: {},
  getters: {},
  actions: {
    "privatekeys/set": vi.fn(),
    "snackbar/showSnackbarSuccessNotRequest": vi.fn(),
    "snackbar/showSnackbarErrorNotRequest": vi.fn(),
  },
});

describe("PrivateKeyFormDialogEdit", () => {
  let wrapper: VueWrapper<InstanceType<typeof PrivateKeyEdit>>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        const wrapper = mount(PrivateKeyEdit, {
          global: {
            plugins: [[store, key], routes, vuetify],
          },
          props: {
            keyObject: test.props.keyObject,
          },
          // shallow: true,
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
        expect(wrapper.vm.name).toEqual(keyObject.name);
        expect(wrapper.vm.supportedKeys).toEqual(test.data.supportedKeys);
      });

      it("Compare data with props", () => {
        expect(wrapper.vm.keyObject).toEqual(test.props.keyObject);
      });

      //////
      // HTML validation
      //////

      it("Renders the correct HTML", () => {
        expect(wrapper.find('[data-test="privatekey-icon"]').exists()).toBeTruthy();
        expect(wrapper.find('[data-test="privatekey-title"]').exists()).toBeTruthy();
      });

      // TODO: make dialog open
    });
  });
});
