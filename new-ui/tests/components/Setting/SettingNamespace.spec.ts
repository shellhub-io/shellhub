import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import SettingNamespace from "../../../src/components/Setting/SettingNamespace.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";
import { envVariables } from "../../../src/envVariables";

const members = [
  {
    id: "xxxxxxxx",
    type: "owner",
    username: "user1",
  },
  {
    id: "xxxxxxxy",
    type: "observer",
    username: "user2",
  },
];

const openNamespace = {
  name: "namespace",
  members,
  owner: "owner",
  tenant_id: "xxxxxxxx",
  devices_count: 1,
  max_devices: 3,
};

const hostedNamespace = { ...openNamespace, max_devices: -1 };

const tests = [
  {
    description: "Open version",
    variables: {
      namespace: openNamespace,
      authID: "xxxxxxxx",
      tenant: "xxxxxxxx",
      hasTenant: true,
      isEnterprise: false,
    },
    data: {
      namespaceMemberFormShow: false,
    },
    computed: {
      namespace: openNamespace,
      tenant: "xxxxxxxx",
      isEnterprise: false,
    },
    components: {
      "namespaceRename-component": true,
      "namespaceMemberFormDialogAdd-component": true,
      "namespaceDelete-component": true,
    },
    template: {
      "tenant-div": true,
      "editOperation-div": true,
      "userOperation-div": true,
      "securityOperation-div": false,
      "deleteOperation-div": true,
      "speed-select": false,
    },
  },
  {
    description: "Hosted version",
    variables: {
      namespace: hostedNamespace,
      authID: "xxxxxxxx",
      tenant: "xxxxxxxx",
      hasTenant: true,
      isEnterprise: true,
    },
    data: {
      namespaceMemberFormShow: false,
    },
    computed: {
      namespace: hostedNamespace,
      tenant: "xxxxxxxx",
      isEnterprise: true,
    },
    components: {
      "namespaceRename-component": true,
      "namespaceMemberFormDialogAdd-component": true,
      "namespaceDelete-component": true,
    },
    template: {
      "tenant-div": true,
      "editOperation-div": true,
      "userOperation-div": true,
      "securityOperation-div": true,
      "deleteOperation-div": true,
      "speed-select": false,
    },
  },
];

const store = (namespace: any, authID: any, tenant: any) => {
  return createStore({
    state: {
      namespace,
      authID,
      tenant,
    },
    getters: {
      "namespaces/get": (state) => state.namespace,
      "auth/tenant": (state) => state.tenant,
    },
    actions: {
      "namespaces/get": () => {},
      "namespaces/removeUser": () => {},
      "snackbar/showSnackbarSuccessAction": () => {},
      "snackbar/showSnackbarErrorAction": () => {},
      "snackbar/showSnackbarErrorAssociation": () => {},
    },
  });
};

describe("SettingNamespace", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(async () => {
        wrapper = mount(SettingNamespace, {
          global: {
            plugins: [
              [
                store(
                  test.variables.namespace,
                  test.variables.authID,
                  test.variables.tenant
                ),
                key,
              ],
              routes,
              vuetify,
            ],
          },
          mocks: {
            $stripe: {
              elements: () => ({
                create: () => ({
                  mount: () => null,
                }),
              }),
            },
          },
          shallow: true,
        });

        envVariables.isEnterprise = test.variables.isEnterprise;
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

      ///////s
      // Data checking
      //////
      it("Data is defined", () => {
        expect(wrapper.vm.$data).toBeDefined();
      });
      it("Compare computed with default value", () => {
        expect(wrapper.vm.namespace).toStrictEqual(test.computed.namespace);
        expect(wrapper.vm.tenant).toBe(test.computed.tenant);
        expect(wrapper.vm.isEnterprise).toBe(test.computed.isEnterprise);
      });

      it("Process data in methods", () => {
        expect(wrapper.vm.hasTenant()).toEqual(test.variables.hasTenant);
      });
    });
  });
});
