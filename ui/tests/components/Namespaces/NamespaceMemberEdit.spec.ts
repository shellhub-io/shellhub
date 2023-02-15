import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import NamespaceMemberEdit from "../../../src/components/Namespace/NamespaceMemberEdit.vue";
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

const memberLocalEdit = {
  id: "xxxxxxxy",
  role: "observer",
  selectedRole: "observer",
  username: "user2",
};

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
      namespaceGlobal,
    },
    props: {
      member: members[1],
      show: false,
      notHasAuthorization: false,
    },
    data: {
      showDialog: false,
      errorMessage: "",
      memberLocal: memberLocalEdit,
      items: ["administrator", "operator", "observer"],
    },
    template: {
      "remove-icon": true,
      "edit-title": true,
      "namespaceNewMember-dialog": false,
    },
    templateText: {
      "edit-title": "Edit",
    },
  },
  {
    description: "Dialog opened",
    variables: {
      namespaceGlobal,
    },
    props: {
      member: members[1],
      show: true,
      notHasAuthorization: false,
    },
    data: {
      showDialog: false,
      errorMessage: "",
      memberLocal: memberLocalEdit,
      items: ["administrator", "operator", "observer"],
    },
    template: {
      "remove-icon": true,
      "edit-title": true,
      "namespaceNewMember-dialog": true,
      "text-title": true,
      "close-btn": true,
      "edit-btn": true,
    },
    templateText: {
      "edit-title": "Edit",
      "text-title": "Update member role",
      "close-btn": "Close",
      "edit-btn": "Edit",
    },
  },
];

const store = (namespace: any) => {
  return createStore({
    state: {
      namespace,
    },
    getters: {
      "namespaces/get": (state) => state.namespace,
    },
    actions: {
      "namespaces/adduser": vi.fn(),
      "snackbar/showSnackbarSuccessAction": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
    },
  });
};

describe("NamespaceMemberEdit", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(() => {
        wrapper = mount(NamespaceMemberEdit, {
          global: {
            plugins: [
              [store(test.variables.namespaceGlobal), key],
              routes,
              vuetify,
            ],
          },
          props: {
            member: test.props.member,
            show: test.props.show,
            notHasAuthorization: test.props.notHasAuthorization,
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
        expect(wrapper.vm.show).toEqual(test.props.show);
      });
      it("Compare data with default value", () => {
        expect(wrapper.vm.memberLocal).toEqual(test.data.memberLocal);
        expect(wrapper.vm.errorMessage).toEqual(test.data.errorMessage);
        expect(wrapper.vm.showDialog).toEqual(test.data.showDialog);
      });

      //////
      // HTML validation
      //////

      it("Renders the template with data", async () => {
        expect(wrapper.find(`[data-test="namespace-edit-icon"]`)).toBeTruthy();
        expect(wrapper.find(`[data-test="namespace-edit-title"]`)).toBeTruthy();
      });
    });
  });
});
