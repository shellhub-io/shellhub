import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import PrivateKeyAdd from "../../../src/components/PrivateKeys/PrivateKeyAdd.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const privateKey = {
  name: "",
  data: "",
};

const tests = [
  {
    description: "Button create private Key",
    props: {
      size: "default",
    },
    data: {
      privateKey,
      supportedKeys:
        "Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.",
    },
    template: {
      "createKey-btn": true,
      "privateKeyFormDialog-card": false,
    },
    templateText: {
      "createKey-btn": "Add Private Key",
    },
  },
  {
    description: "Dialog opened",
    props: {
      size: "default",
    },
    data: {
      privateKey,
      supportedKeys:
        "Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.",
    },
    template: {
      "createKey-btn": true,
      "privateKeyFormDialog-card": true,
      "text-title": true,
      "name-field": true,
      "data-field": true,
      "cancel-btn": true,
      "create-btn": true,
    },
    templateText: {
      "createKey-btn": "Add Private Key",
      "text-title": "New Private Key",
      "name-field": "",
      "data-field": "",
      "cancel-btn": "Cancel",
      "create-btn": "Create",
    },
  },
];

const store = createStore({
  state: {},
  getters: {
    "auth/role": () => "admin",
  },
  actions: {
    "privateKey/set": vi.fn(),
    "snackbar/showSnackbarSuccessNotRequest": vi.fn(),
    "snackbar/showSnackbarErrorNotRequest": vi.fn(),
  },
});

describe("PrivateKeyFormDialogAdd", () => {
  let wrapper: VueWrapper<InstanceType<typeof PrivateKeyAdd>>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        const wrapper = mount(PrivateKeyAdd, {
          global: {
            plugins: [[store, key], routes, vuetify],
          },
          props: {
            size: test.props.size,
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

      it("Compare data with default value", () => {
        expect(wrapper.vm.name).toBe(test.data.privateKey.name);
        expect(wrapper.vm.privateKeyData).toBe(test.data.privateKey.data);
        expect(wrapper.vm.supportedKeys).toBe(test.data.supportedKeys);
      });

      it("Compare data with props", () => {
        expect(wrapper.vm.size).toBe(test.props.size);
      });

      //////
      // HTML validation
      //////

      // TODO
    });
  });
});
