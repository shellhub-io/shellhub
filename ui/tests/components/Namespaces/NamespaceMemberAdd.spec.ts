import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import NamespaceMemberAdd from "../../../src/components/Namespace/NamespaceMemberAdd.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const members = [
  {
    id: "xxxxxxxx",
    role: "owner",
    username: "user1",
  },
  {
    id: "xxxxxxxy",
    role: "observer",
    username: "user2",
  },
];

const namespaceGlobal = {
  name: "namespace",
  owner: "user1",
  members,
  tenant_id: "xxxxxxxx",
  devices_count: 0,
  max_devices: 0,
};

const tests = [
  {
    description: "Button add user has authorization",
    variables: {
      namespaceGlobal,
    },
    role: {
      type: "owner",
      permission: true,
    },
    data: {
      username: "",
      selectedRole: "",
      dialog: false,
      items: ["administrator", "operator", "observer"],
    },
    computed: {
      hasAuthorization: true,
    },
    template: {
      "addMember-btn": true,
      "namespaceNewMember-dialog": false,
      "close-btn": false,
      "add-btn": false,
    },
  },
  {
    description: "Button add user has no authorization",
    variables: {
      namespaceGlobal,
    },
    role: {
      type: "operator",
      permission: false,
    },
    data: {
      username: "",
      selectedRole: "",
      dialog: false,
      items: ["administrator", "operator", "observer"],
    },
    computed: {
      hasAuthorization: false,
    },
    template: {
      "addMember-btn": true,
      "namespaceNewMember-dialog": false,
      "close-btn": false,
      "add-btn": false,
    },
  },
  {
    description: "dialog add user has authorization",
    variables: {
      namespaceGlobal,
    },
    role: {
      type: "owner",
      permission: true,
    },
    data: {
      username: "",
      selectedRole: "",
      dialog: false,
      items: ["administrator", "operator", "observer"],
    },
    computed: {
      hasAuthorization: true,
    },
    template: {
      "addMember-btn": true,
      "namespaceNewMember-dialog": true,
      "close-btn": true,
      "add-btn": true,
    },
  },
];

const store = (namespace: any, currentRole: string) => {
  return createStore({
    state: {
      namespace,
      currentRole,
    },
    getters: {
      "namespaces/get": (state) => state.namespace,
      "auth/role": (state) => state.currentRole,
    },
    actions: {
      "namespaces/adduser": vi.fn(),
      "snackbar/showSnackbarSuccessAction": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
    },
  });
};

describe("NamespaceMemberAdd", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    describe(`${test.description} - ${test.role.type}`, () => {
      beforeEach(() => {
        wrapper = mount(NamespaceMemberAdd, {
          global: {
            plugins: [
              [store(test.variables.namespaceGlobal, test.role.type), key],
              routes,
              vuetify,
            ],
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
        expect(wrapper.vm.username).toEqual(test.data.username);
        expect(wrapper.vm.selectedRole).toEqual(test.data.selectedRole);
        expect(wrapper.vm.dialog).toEqual(test.data.dialog);
        expect(wrapper.vm.items).toEqual(test.data.items);
      });
      it("Compare computed with default value", () => {
        expect(wrapper.vm.hasAuthorization()).toEqual(test.computed.hasAuthorization);
      });
    });
  });
});
