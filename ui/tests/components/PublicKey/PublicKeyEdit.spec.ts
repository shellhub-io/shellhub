import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import PublicKeyEdit from "../../../src/components/PublicKeys/PublicKeyEdit.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const supportedKeys =
  "Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.";

const keyObject = {
  name: "ShellHub",
  username: "ShellHub",
  data: "",
  filter: {
    hostname: ".*",
  },
};

const keyObject2 = {
  name: "ShellHub",
  username: "ShellHub",
  data: "",
  filter: {
    tags: ["tag1", "tag2"],
  },
};

const tests = [
  {
    description: "Dialog closed",
    props: {
      keyObject,
      show: false,
    },
    data: {
      keyObject,
      supportedKeys,
      showDialog: false,
      choiceFilter: "hostname",
      validateLength: true,
      errMsg: "",
      choiceUsername: "username",
    },
    template: {
      "edit-icon": true,
      "edit-title": true,
      "namespaceNewMember-dialog": false,
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
      keyObject,
      supportedKeys,
      showDialog: false,
      choiceFilter: "hostname",
      validateLength: true,
      errMsg: "",
      choiceUsername: "username",
    },
    template: {
      "edit-icon": true,
      "edit-title": true,
      "publicKeyFormDialog-card": true,
      "text-title": true,
      "name-field": true,
      "hostname-field": true,
      "tags-field": false,
      "username-field": true,
      "data-field": true,
      "cancel-btn": true,
      "edit-btn": true,
      "access-restriction-field": true,
    },
    templateText: {
      "edit-title": "Edit",
      "text-title": "Edit Public Key",
      "cancel-btn": "Cancel",
      "edit-btn": "Edit",
    },
  },
  {
    description: "Dialog with tags",
    props: {
      keyObject: keyObject2,
      show: true,
    },
    data: {
      keyObject: keyObject2,
      supportedKeys,
      showDialog: false,
      choiceFilter: "hostname",
      validateLength: true,
      errMsg: "",
      choiceUsername: "username",
    },
    template: {
      "edit-icon": true,
      "edit-title": true,
      "publicKeyFormDialog-card": true,
      "text-title": true,
      "name-field": true,
      "username-field": true,
      "data-field": true,
      "cancel-btn": true,
      "edit-btn": true,
      "access-restriction-field": true,
    },
    templateText: {
      "edit-title": "Edit",
      "text-title": "Edit Public Key",
      "cancel-btn": "Cancel",
      "edit-btn": "Edit",
    },
  },
];

const store = createStore({
  state: {},
  getters: {},
  actions: {
    "publicKeys/put": vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

describe("PublicKeyFormDialogEdit", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(PublicKeyEdit, {
          global: {
            plugins: [[store, key], routes, vuetify],
          },
          props: {
            keyObject: test.props.keyObject,
          },
          shallow: false,
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
        expect(wrapper.vm.keyObject).toEqual(test.data.keyObject);
        expect(wrapper.vm.showDialog).toEqual(test.data.showDialog);
        expect(wrapper.vm.choiceFilter).toEqual(test.data.choiceFilter);
        expect(wrapper.vm.validateLength).toEqual(test.data.validateLength);
        expect(wrapper.vm.errMsg).toEqual(test.data.errMsg);
        expect(wrapper.vm.choiceUsername).toEqual(test.data.choiceUsername);
      });

      //////
      // HTML validation
      //////

      it("Renders the correct HTML", () => {
        expect(wrapper.find('[data-test="publicKey-edit-icon"]').exists()).toBeTruthy();
        expect(wrapper.find('[data-test="publicKey-edit-title"]').exists()).toBeTruthy();
      });

      // Todo dialog opened
    });
  });
});
