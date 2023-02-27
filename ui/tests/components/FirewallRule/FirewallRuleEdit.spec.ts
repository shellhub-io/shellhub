import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import FirewallRuleEdit from "../../../src/components/firewall/FirewallRuleEdit.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

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
    name: "allow",
  },
  {
    id: "deny",
    name: "deny",
  },
];

const firewallRuleProps = {
  id: "5f1996c8",
  tenant_id: "xxxxxxxx",
  priority: 4,
  source_ip: "00.00.00",
  username: ".*",
  filter: {
    hostname: "hostname",
  },
  action: "allow",
  active: true,
};

const firewallRuleData = {
  priority: 0,
  source_ip: "",
  username: "",
  filter: {},
  policy: "",
  status: "",
};

const tests = [
  {
    description: "Dialog closed",
    props: {
      firewallRule: firewallRuleProps,
      show: false,
    },
    data: {
      showDialog: false,
      hostnameField: "",
      state: stateRuleFirewall,
      choiceFilter: "all",
      choiceUsername: "all",
      usernameFieldChoices,
      filterFieldChoices,
      ruleFirewallLocal: firewallRuleData,
    },
    template: {
      "edit-icon": true,
      "edit-title": true,
      "firewallRuleForm-card": false,
    },
    templateText: {
      "edit-title": "Edit",
    },
  },
  {
    description: "Dialog opened",
    props: {
      firewallRule: firewallRuleProps,
      show: true,
    },
    data: {
      showDialog: true,
      state: stateRuleFirewall,
      hostnameField: "",
      choiceFilter: "all",
      choiceUsername: "all",
      usernameFieldChoices,
      filterFieldChoices,
      ruleFirewallLocal: firewallRuleData,
    },
    template: {
      "edit-icon": true,
      "edit-title": true,
      "firewallRuleForm-card": true,
      "text-title": true,
      "priority-field": true,
      "action-field": true,
      "source_ip-field": true,
      "username-field": true,
      "filter-field": true,
      "cancel-btn": true,
      "edit-btn": true,
    },
    templateText: {
      "edit-title": "Edit",
      "text-title": "Edit Firewall Rule",
      "cancel-btn": "Cancel",
      "edit-btn": "Edit",
    },
  },
];

const store = createStore({
  state: {},
  getters: {},
  actions: {
    "firewallrules/post": vi.fn(),
    "firewallrules/put": vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

describe("FirewallRuleFormDialog", () => {
  let wrapper: VueWrapper<InstanceType<typeof FirewallRuleEdit>>;

  tests.forEach((test) => {
    describe(`${test.description}`, () => {
      beforeEach(async () => {
        const vuetify = createVuetify();
        wrapper = mount(FirewallRuleEdit, {
          global: {
            plugins: [[store, key], routes, vuetify],
          },
          setup() {
            return {
              ...test.data,
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

      ///////
      // Data checking
      //////

      ///////
      // Data checking
      //////

      it("Compare data with default value", () => {
        Object.keys(test.data).forEach((key) => {
          expect(wrapper.vm[key]).toEqual(test.data[key]);
        });
      });

      //////
      // HTML validation
      //////

      it("Renders the template with data", () => {
        // TODO
      });
    });
  });

  describe("Update data checks", () => {
    beforeEach(() => {
      const vuetify = createVuetify();
      wrapper = mount(FirewallRuleEdit, {
        global: {
          plugins: [[store, key], vuetify, routes],
        },
        props: {
          firewallRule: firewallRuleProps,
          show: true,
        },
        shallow: true,
      });
    });
    it("Should setLocalVariable populate ruleFirewallLocal for tag filter", async () => {
      wrapper.setProps({
        firewallRule: {
          priority: 1,
          source_ip: "2",
          filter: {
            hostname: "name",
          },
          username: ".*",
          status: "active",
          policy: "allow",
        },
      });
      await flushPromises();
      wrapper.vm.setLocalVariable();

      expect(wrapper.vm.choiceFilter).toBe("hostname");
      expect(wrapper.vm.choiceUsername).toBe("all");
    });

    it("Should select restriction update ruleFirewallLocal filter tags", async () => {
      const rfl = wrapper.vm.ruleFirewallLocal;

      const tags = ["tag1", "tag2"];
      wrapper.vm.choiceFilter = "tags";
      wrapper.vm.tagChoices = tags;
      wrapper.vm.selectRestriction();

      await flushPromises();
      expect(wrapper.vm.ruleFirewallLocal).toStrictEqual({
        ...rfl,
        filter: { tags },
      });
    });

    it("Should select restriction update ruleFirewallLocal filter ip_address", async () => {
      const ipAddr = "00.00.00";

      wrapper.vm.choiceIP = "ipDetails";
      wrapper.vm.ipField = ipAddr;

      wrapper.vm.selectRestriction();
      wrapper.vm.setLocalVariable();
      await flushPromises();
      const rfl = wrapper.vm.ruleFirewallLocal;

      expect(wrapper.vm.ruleFirewallLocal).toStrictEqual({
        ...rfl,
        source_ip: ipAddr,
      });
    });

    it("Should select restriction update ruleFirewallLocal for field username", async () => {
      wrapper.vm.setLocalVariable();
      const rfl = wrapper.vm.ruleFirewallLocal;
      const uf = "user";
      wrapper.vm.choiceUsername = "username";
      wrapper.vm.username = uf;

      wrapper.vm.selectRestriction();
      await flushPromises();

      expect(wrapper.vm.ruleFirewallLocal).toStrictEqual({
        ...rfl,
        username: uf,
      });
    });

    it("Should select restriction update ruleFirewallLocal filter hostname", async () => {
      wrapper.vm.setLocalVariable();
      const rfl = wrapper.vm.ruleFirewallLocal;
      const { filter } = wrapper.props("firewallRule");

      const hostname = "hostname";
      wrapper.vm.choiceFilter = "hostname";
      filter.hostname = hostname;
      await flushPromises();
      wrapper.vm.selectRestriction();
      expect(wrapper.vm.ruleFirewallLocal).toStrictEqual({
        ...rfl,
        filter: { hostname },
      });
    });

    it("Should edit method call hasErros", async () => {
      // TODO
    });
  });
});
