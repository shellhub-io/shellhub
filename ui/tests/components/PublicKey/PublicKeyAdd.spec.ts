import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import PublicKeyAdd from "../../../src/components/PublicKeys/PublicKeyAdd.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const tests = [
  {
    description: "Button create publicKey has authorization",
    role: {
      type: "owner",
      permission: true,
    },
    data: {
      dialog: true,
      keyLocal: {},
      action: "create",
      hostname: "",
      tagChoices: [],
      validateLength: true,
      errMsg: "",
      choiceFilter: "all",
      choiceUsername: "all",
      supportedKeys:
        "Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.",
    },
    computed: {
      hasAuthorization: true,
    },
    template: {
      "createKey-btn": true,
      "publicKeyFormDialog-card": false,
    },
    templateText: {
      "createKey-btn": "Add Public Key",
    },
  },
  {
    description: "Button create publicKey has no authorization",
    role: {
      type: "operator",
      permission: false,
    },
    data: {
      dialog: true,
      keyLocal: {},
      action: "create",
      hostname: "",
      tagChoices: [],
      validateLength: true,
      errMsg: "",
      choiceFilter: "all",
      choiceUsername: "all",
      supportedKeys:
        "Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.",
    },
    computed: {
      hasAuthorization: false,
    },
    template: {
      "createKey-btn": true,
      "publicKeyFormDialog-card": false,
    },
    templateText: {
      "createKey-btn": "Add Public Key",
    },
  },
  {
    description: "Dialog create publicKey has authorization",
    role: {
      type: "owner",
      permission: true,
    },
    data: {
      dialog: true,
      keyLocal: {},
      action: "create",
      hostname: "",
      tagChoices: [],
      validateLength: true,
      errMsg: "",
      choiceFilter: "all",
      choiceUsername: "all",
      supportedKeys:
        "Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.",
    },
    computed: {
      hasAuthorization: true,
    },
    template: {
      "createKey-btn": true,
      "publicKeyFormDialog-card": true,
      "text-title": true,
      "name-field": true,
      "hostname-field": false,
      "username-field": false,
      "data-field": true,
      "cancel-btn": true,
      "create-btn": true,
    },
    templateText: {
      "createKey-btn": "Add Public Key",
      "text-title": "New Public Key",
      "name-field": "",
      "data-field": "",
      "cancel-btn": "Cancel",
      "create-btn": "Create",
    },
  },
];

const store = (currentrole: string) => createStore({
  state: {
    currentrole,
  },
  getters: {
    "auth/role": (state) => state.currentrole,
  },
  actions: {
    "publicKeys/post": vi.fn(),
    "publicKeys/put": vi.fn(),
    "privateKey/set": vi.fn(),
    "privateKey/edit": vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
    "snackbar/showSnackbarSuccessNotRequest": vi.fn(),
    "snackbar/showSnackbarErrorNotRequest": vi.fn(),
  },
});

describe("PublicKeyFormDialogAdd", () => {
  let wrapper: VueWrapper<InstanceType<typeof PublicKeyAdd>>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    describe(`${test.description} - ${test.role.type}`, () => {
      beforeEach(() => {
        wrapper = mount(PublicKeyAdd, {
          global: {
            plugins: [[store(test.role.type), key], routes, vuetify],
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
        expect(wrapper.vm.supportedKeys).toEqual(test.data.supportedKeys);
        expect(wrapper.vm.keyLocal).toEqual(test.data.keyLocal);
        expect(wrapper.vm.hostname).toEqual(test.data.hostname);
        expect(wrapper.vm.tagChoices).toEqual(test.data.tagChoices);
        expect(wrapper.vm.validateLength).toEqual(test.data.validateLength);
        expect(wrapper.vm.errMsg).toEqual(test.data.errMsg);
        expect(wrapper.vm.choiceFilter).toEqual(test.data.choiceFilter);
        expect(wrapper.vm.choiceUsername).toEqual(test.data.choiceUsername);
      });

      it("Compare the computed with the default value", () => {
        expect(wrapper.vm.hasAuthorization).toEqual(
          test.computed.hasAuthorization,
        );
      });

      ///////
      // HTML validation
      //////

      it("Renders the correct HTML", () => {
        // TODO
      });
    });
  });
});
