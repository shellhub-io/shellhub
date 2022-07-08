import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import NamespaceMemberDelete from "../../../src/components/Namespace/NamespaceMemberDelete.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

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
    description: "Dialog closed",
    variables: {
      namespace: namespaceGlobal,
    },
    props: {
      member: members[0],
      hasAuthorization: true,
      show: false,
    },
    template: {
      "remove-icon": true,
      "remove-title": true,
      "namespaceMemberDelete-dialog": false,
      "close-btn": false,
      "remove-btn": false,
    },
  },
  {
    description: "Dialog opened",
    variables: {
      namespace: namespaceGlobal,
    },
    props: {
      member: members[0],
      show: true,
      hasAuthorization: true,
    },
    template: {
      "remove-icon": true,
      "remove-title": true,
      "namespaceMemberDelete-dialog": true,
      "close-btn": true,
      "remove-btn": true,
    },
  },
];

const store = (namespace: any, tenant: any) => {
  return createStore({
    state: {
      namespace,
      tenant,
    },
    getters: {
      "namespaces/get": (state) => state.namespace,
      "auth/tenant": (state) => state.tenant,
    },
    actions: {
      "namespaces/removeUser": vi.fn(),
      "snackbar/showSnackbarSuccessAction": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
    },
  });
};

describe("NamespaceMemberDelete", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(NamespaceMemberDelete, {
          global: {
            plugins: [
              [
                store(
                  test.variables.namespace,
                  test.variables.namespace.tenant_id
                ),
                key,
              ],
              routes,
              vuetify,
            ],
          },
          props: {
            member: test.props.member,
            hasAuthorization: test.props.hasAuthorization,
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
      it("Receive data in props", () => {
        expect(wrapper.vm.member).toEqual(test.props.member);
        expect(wrapper.vm.hasAuthorization).toEqual(
          test.props.hasAuthorization
        );
      });

      //////
      // HTML validation
      //////

      it("Renders the correct HTML", () => {
        expect(
          wrapper.find('[data-test="namespace-delete-icon"]').exists()
        ).toBeTruthy();
        expect(
          wrapper.find('[data-test="namespace-delete-title"]').exists()
        ).toBeTruthy();
      });
    });
  });
});
