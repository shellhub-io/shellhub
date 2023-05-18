import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import FirewallRuleDelete from "../../../src/components/firewall/FirewallRuleDelete.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";
import { IRole } from "@/interfaces/IRole";

const usernameFieldChoices = [
  {
    filterName: "all",
    filterText: "Define rule to all users",
  },
  {
    filterName: "username",
    filterText: "Restrict access using a regexp for username",
  },
];

const filterFieldChoices = [
  {
    filterName: "all",
    filterText: "Define rule to all devices",
  },
  {
    filterName: "hostname",
    filterText: "Restrict rule with a regexp for hostname",
  },
  {
    filterName: "tags",
    filterText: "Restrict rule by device tags",
  },
];
const stateRuleFirewall = [
  {
    id: "allow",
    name: "Allow",
  },
  {
    id: "deny",
    name: "Deny",
  },
];

const ruleFirewall = {
  policy: "allow",
  status: "active",
  priority: "",
  source_ip: "",
  username: "",
};

const tests = [
  {
    description: "Dialog closed",
    role: {
      type: "owner",
      permission: true,
    },
    data: {
      showDialog: false,
      action: "create",
      usernameFieldChoices,
      filterFieldChoices,
      ruleFirewall,
      state: stateRuleFirewall,
    },
    props: {
      id: "5f1996c8",
      notHasAuthorization: false,
    },
    computed: {
      hasAuthorization: true,
    },
    template: {
      "add-btn": true,
      "firewallRuleForm-card": false,
    },
    templateText: {
      "add-btn": "Add Rule",
    },
  },
  {
    description: "Dialog closed",
    role: {
      type: "operator",
      permission: false,
    },
    data: {
      showDialog: false,
      usernameFieldChoices,
      filterFieldChoices,
      action: "create",
      ruleFirewall,
      state: stateRuleFirewall,
    },
    props: {
      id: "5f1996c8",
      notHasAuthorization: false,
    },
    computed: {
      hasAuthorization: false,
    },
    template: {
      "add-btn": true,
      "firewallRuleForm-card": false,
    },
    templateText: {
      "add-btn": "Add Rule",
    },
  },
  {
    description: "Dialog opened",
    role: {
      type: "owner",
      permission: true,
    },
    data: {
      usernameFieldChoices,
      filterFieldChoices,
      showDialog: true,
      action: "create",
      ruleFirewall,
      state: stateRuleFirewall,
    },
    props: {
      id: "5f1996c8",
      notHasAuthorization: false,
    },
    computed: {
      hasAuthorization: true,
    },
    template: {
      "add-btn": true,
      "firewallRuleForm-card": true,
      "text-title": true,
      "cancel-btn": true,
      "create-btn": true,
    },
    templateText: {
      "add-btn": "Add Rule",
      "text-title": "New Firewall Rule",
      "cancel-btn": "Cancel",
      "create-btn": "Create",
    },
  },
];

const store = (currentrole: IRole) => createStore({
  state: {
    currentrole,
  },
  getters: {
    "auth/role": (state) => state.currentrole,
  },
  actions: {
    "firewallrules/post": vi.fn(),
    "firewallrules/put": vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

describe("FirewallRuleFormDialog", () => {
  let wrapper: VueWrapper<InstanceType<typeof FirewallRuleDelete>>;

  tests.forEach((test) => {
    describe(`${test.description} - ${test.role.type}`, () => {
      beforeEach(() => {
        const vuetify = createVuetify();

        const wrapper = mount(FirewallRuleDelete, {
          global: {
            plugins: [[store(test.role), key], vuetify, routes],
          },
          props: {
            id: test.props.id,
            notHasAuthorization: test.props.notHasAuthorization,
          },
          setup() {
            return {
              ...test.data,
              ...test.computed,
            };
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
    });

    ///////
    // Data checking
    //////

    it("Compare data with default value", () => {
      Object.keys(test.data).forEach((key) => {
        expect(wrapper.vm[key]).toEqual(test.data[key]);
      });
    });

    it("Compare computed with default value", () => {
      Object.keys(test.computed).forEach((key) => {
        expect(wrapper.vm[key]).toEqual(test.computed[key]);
      });
    });

    //////
    // HTML validation
    //////

    it("Compare HTML with default value", () => {
      // TODO
    });
  });
});
