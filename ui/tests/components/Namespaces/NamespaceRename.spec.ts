import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import NamespaceRename from "../../../src/components/Namespace/NamespaceRename.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const role = ["owner", "operator"];

const hasAuthorizationRenameNamespace = {
  owner: true,
  operator: false,
};

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

const invalidNamespaces = [
  "'",
  '"',
  "!",
  "@",
  "#",
  "$",
  "%",
  "¨",
  "&",
  "*",
  "(",
  ")",
  "-",
  "_",
  "=",
  "+",
  "´",
  "`",
  "[",
  "{",
  "~",
  "^",
  "]",
  ",",
  "<",
  "..",
  ">",
  ";",
  ":",
  "/",
  "?",
];

const invalidMinAndMaxCharacters = [
  "s",
  "sh",
  "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
];

const tests = [
  {
    description: "Open version",
    variables: {
      namespace: openNamespace,
      tenant: "xxxxxxxx",
      hasTenant: true,
      isEnterprise: false,
    },
    data: {
      name: "",
    },
    computed: {
      namespace: openNamespace,
      tenant: "xxxxxxxx",
    },
    template: {
      "name-text": true,
    },
  },
  {
    description: "Hosted version",
    variables: {
      namespace: hostedNamespace,
      tenant: "xxxxxxxx",
      hasTenant: true,
      isEnterprise: true,
    },
    data: {
      name: "",
    },
    computed: {
      namespace: hostedNamespace,
      tenant: "xxxxxxxx",
    },
    template: {
      "name-text": true,
    },
  },
];

const store = (namespace: typeof openNamespace, tenant: string, currentrole: string) => createStore({
  state: {
    namespace,
    tenant,
    currentrole,
  },
  getters: {
    "namespaces/get": (state) => state.namespace,
    "auth/tenant": (state) => state.tenant,
    "auth/role": (state) => state.currentrole,
  },
  actions: {
    "namespaces/put": vi.fn(),
    "namespaces/get": vi.fn(),
    "namespaces/removeUser": vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

describe("NamespaceRename", () => {
  let wrapper: VueWrapper<InstanceType<typeof NamespaceRename>>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    role.forEach((currentrole) => {
      describe(`${test.description} ${currentrole}`, () => {
        beforeEach(async () => {
          wrapper = mount(NamespaceRename, {
            global: {
              plugins: [
                [
                  store(
                    test.variables.namespace,
                    test.variables.tenant,
                    currentrole,
                  ),
                  key,
                ],
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
          expect(wrapper.vm.name).toEqual(test.data.name);
        });
        it("Process data in the computed", () => {
          expect(wrapper.vm.namespace).toEqual(test.computed.namespace);
          expect(wrapper.vm.tenant).toEqual(test.computed.tenant);
        });

        //////
        // HTML validation
        //////
        it("Show validation messages", async () => {
          wrapper.vm.name = "ShelHub";
          wrapper.vm.name = undefined;
          await flushPromises();
          expect(wrapper.vm.nameError).toBe("this is a required field");
        });

        it("Show validation messages", async () => {
          wrapper.vm.name = "ShelHub..";
          await flushPromises();
          expect(wrapper.vm.nameError).toBe("The name must not contain dots");
        });

        // TODO check later
        // it("Show validation messages", () => {
        //   wrapper.vm.name = "";
        //   invalidMinAndMaxCharacters.forEach(async (character) => {
        //     wrapper.vm.name = character;
        //     await flushPromises();
        //     expect(wrapper.vm.nameError).toBe(
        //       "Your namespace should be 3-30 characters long"
        //     );
        //   });
        // });
      });
    });
  });
});
