import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import FirewallRuleAdd from "../../../src/components/firewall/FirewallRuleAdd.vue";
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
    props: {
      size: "default",
    },
    data: {
      dialog: false,
      action: "create",
      usernameFieldChoices,
      filterFieldChoices,
      ruleFirewall,
      state: stateRuleFirewall,
      choiceUsername: "all",
      choiceIP: "all",
      choiceFilter: "all",
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
    props: {
      size: "default",
    },
    data: {
      dialog: false,
      usernameFieldChoices,
      filterFieldChoices,
      action: "create",
      ruleFirewall,
      state: stateRuleFirewall,
      choiceUsername: "all",
      choiceIP: "all",
      choiceFilter: "all",
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
    props: {
      size: "default",
    },
    data: {
      usernameFieldChoices,
      filterFieldChoices,
      dialog: true,
      action: "create",
      ruleFirewall,
      state: stateRuleFirewall,
      choiceUsername: "all",
      choiceIP: "all",
      choiceFilter: "all",
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

const store = (currentrole: IRole | string) => createStore({
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
  let wrapper: VueWrapper<InstanceType<typeof FirewallRuleAdd>>;

  tests.forEach((test) => {
    describe(`${test.description} - ${test.role.type}`, () => {
      beforeEach(() => {
        const vuetify = createVuetify();
        wrapper = mount(FirewallRuleAdd, {
          global: {
            plugins: [[store(test.role), key], vuetify, routes],
          },
          props: test.props,
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

      it("Renders the template with data", () => {
        // TODO
      });
    });
  });

  describe("Update data checks", () => {
    beforeEach(() => {
      const vuetify = createVuetify();
      wrapper = mount(FirewallRuleAdd, {
        global: {
          plugins: [[store("owner"), key], vuetify, routes],
        },
        props: {
          size: "default",
        },
        shallow: true,
      });
    });

    it("Should construct filter object for hostname", async () => {
      const rf = wrapper.vm.ruleFirewall;

      wrapper.vm.choiceFilter = "hostname";
      wrapper.vm.filterField = "another";
      await wrapper.vm.$nextTick();

      await flushPromises();
      wrapper.vm.constructFilterObject();

      expect(wrapper.vm.ruleFirewall).toStrictEqual({
        ...rf,
        filter: { hostname: "another" },
      });
    });

    it("Should construct filter object for tags", async () => {
      const rf = wrapper.vm.ruleFirewall;

      const tags = ["tag1", "tag2"];

      wrapper.vm.choiceFilter = "tags";
      wrapper.vm.tagChoices = tags;

      await wrapper.vm.$nextTick();
      await flushPromises();

      wrapper.vm.constructFilterObject();

      expect(wrapper.vm.ruleFirewall).toStrictEqual({ ...rf, filter: { tags } });
    });

    it("Should call constructFilterObject when choiceFilter changes", async () => {
      const spy = vi.spyOn(wrapper.vm, "constructFilterObject");

      wrapper.vm.choiceFilter = "hostname";
      await wrapper.vm.$nextTick();
      wrapper.vm.constructFilterObject();
      await flushPromises();

      expect(spy).toHaveBeenCalled();
    });
  });
});
