import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import NamespaceMemberList from "../../../src/components/Namespace/NamespaceMemberList.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const tenantGlobal = "xxxxxxxx";

const namespace = {
  name: "nsxxx",
  members: [
    { username: "user1", role: "owner" },
    { username: "user2", role: "administrator" },
    { username: "user3", role: "observer" },
  ],
};

const headers = [
  {
    text: "Username",
    value: "username",
    align: "start",
    sortable: false,
  },
  {
    text: "Role",
    value: "role",
    align: "center",
    sortable: false,
  },
  {
    text: "Actions",
    value: "actions",
    align: "end",
    sortable: false,
  },
];

const tests = [
  {
    description: "List data when user has owner role",
    variables: {
      tenant: tenantGlobal,
    },
    role: {
      type: "owner",
      permission: true,
    },
    props: {
      namespace,
    },
    data: {
      menu: false,
      headers,
    },
    computed: {
      tenant: tenantGlobal,
      members: namespace.members,
      hasAuthorizationEditMember: true,
      hasAuthorizationRemoveMember: true,
    },
  },
  {
    description: "List data when user has observer role",
    variables: {
      tenant: tenantGlobal,
    },
    role: {
      type: "observer",
      permission: false,
    },
    props: {
      namespace,
    },
    data: {
      menu: false,
      headers,
    },
    computed: {
      tenant: tenantGlobal,
      members: namespace.members,
      hasAuthorizationEditMember: false,
      hasAuthorizationRemoveMember: false,
    },
  },
];

const store = (tenant: string, currentRole: string) => createStore({
  state: {
    tenant,
    currentRole,
  },
  getters: {
    "auth/tenant": (state) => state.tenant,
    "auth/role": (state) => state.currentRole,
  },
  actions: {
    "snackbar/showSnackbarErrorAssociation": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
    "namespaces/get": vi.fn(),
  },
});

describe("NamespaceMemberList", () => {
  let wrapper: VueWrapper<InstanceType<typeof NamespaceMemberList>>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(NamespaceMemberList, {
          global: {
            plugins: [
              [store(test.variables.tenant, test.role.type), key],
              routes,
              vuetify,
            ],
          },
          props: {
            namespace: test.props.namespace,
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
      it("Receive data in props", () => {
        expect(wrapper.vm.namespace).toStrictEqual(test.props.namespace);
      });
      it("Compare data with default value", () => {
        expect(wrapper.vm.headers).toStrictEqual(test.data.headers);
      });
      it("Process data in the computed", () => {
        expect(wrapper.vm.tenant).toStrictEqual(test.computed.tenant);
        expect(wrapper.vm.members).toStrictEqual(test.computed.members);
        expect(wrapper.vm.hasAuthorizationEditMember()).toStrictEqual(test.computed.hasAuthorizationEditMember);
        expect(wrapper.vm.hasAuthorizationRemoveMember()).toStrictEqual(test.computed.hasAuthorizationRemoveMember);
      });
    });
  });
});
