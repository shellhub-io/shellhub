import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import FirewallRule from "../../../src/components/firewall/FirewallRuleList.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";
import { IRole } from "@/interfaces/IRole";

const statusGlobal = true;

const numberFirewallsGlobal = 2;

const firewallsGlobal = [
  {
    id: "5f1996c8",
    tenant_id: "xxxxxxxx",
    priority: 4,
    action: "allow",
    active: true,
    source_ip: "00.00.00",
    username: "shellhub",
    filter: {
      hostname: "shellhub",
    },
  },
  {
    id: "5f1996c8",
    tenant_id: "xxxxxxxx",
    priority: 3,
    action: "allow",
    active: false,
    source_ip: "00.00.00",
    username: "shellhub",
    filter: {
      tags: ["tag1", "tag2"],
    },
  },
];
const headers = [
  {
    text: "Active",
    value: "active",
  },
  {
    text: "Priority",
    value: "priority",
  },
  {
    text: "Action",
    value: "action",
  },
  {
    text: "Source IP",
    value: "source_ip",
  },
  {
    text: "Username",
    value: "username",
  },
  {
    text: "Filter",
    value: "filter",
  },
  {
    text: "Actions",
    value: "actions",
  },
];

const tests = [
  {
    description: "List data when user has owner role",
    role: {
      type: "owner",
      permission: true,
    },
    variables: {
      firewallsGlobal,
      numberFirewallsGlobal,
      statusGlobal,
    },
    data: {
      firewallRuleEditShow: [false, false],
      firewallRuleDeleteShow: [false, false],
      editAction: "edit",
      removeAction: "remove",
      headers,
      getNumberFirewallRules: 2,
      page: 1,
      itemsPerPage: 10,
      next: vi.fn(),
      prev: vi.fn(),
      nextPage: vi.fn(),
      previousPage: vi.fn(),
      changeItemsPerPage: vi.fn(),
      refreshFirewallRules: vi.fn(),
      loading: false,
    },
    computed: {
      firewallRules: firewallsGlobal,
      getNumberFirewallRules: numberFirewallsGlobal,
      hasAuthorizationFormDialogEdit: true,
      hasAuthorizationFormDialogRemove: true,
    },
  },
  {
    description: "List data when user has operator role",
    role: {
      type: "operator",
      permission: false,
    },
    variables: {
      firewallsGlobal,
      numberFirewallsGlobal,
      statusGlobal,
    },
    data: {
      firewallRuleEditShow: [false, false],
      firewallRuleDeleteShow: [false, false],
      editAction: "edit",
      removeAction: "remove",
      headers,
      getNumberFirewallRules: 2,
      page: 1,
      itemsPerPage: 10,
      next: vi.fn(),
      prev: vi.fn(),
      nextPage: vi.fn(),
      previousPage: vi.fn(),
      changeItemsPerPage: vi.fn(),
      refreshFirewallRules: vi.fn(),
      loading: false,
    },
    computed: {
      firewallRules: firewallsGlobal,
      getNumberFirewallRules: numberFirewallsGlobal,
      hasAuthorizationFormDialogEdit: false,
      hasAuthorizationFormDialogRemove: false,
    },
  },
];

const store = (
  firewalls: typeof firewallsGlobal,
  numberFirewalls: number,
  status: boolean,
  currentrole: IRole,
) => createStore({
  state: {
    firewalls,
    numberFirewalls,
    status,
    currentrole,
  },
  getters: {
    "firewallrules/list": (state) => state.firewalls,
    "firewallrules/getNumberFirewalls": (state) => state.numberFirewalls,
    "boxs/getStatus": (state) => state.status,
    "auth/role": (state) => state.currentrole,
  },
  actions: {
    "firewallrules/fetch": vi.fn(),
    "boxs/setStatus": vi.fn(),
    "snackbar/showSnackbarErrorAssociation": vi.fn(),
    "snackbar/showSnackbarErrorLoading": vi.fn(),
  },
});

tests.forEach((test) => {
  describe(`${test.description}`, () => {
    let wrapper: VueWrapper<InstanceType<typeof FirewallRule>>;

    beforeEach(() => {
      const vuetify = createVuetify();
      wrapper = mount(FirewallRule, {
        global: {
          plugins: [
            [
              store(
                test.variables.firewallsGlobal,
                test.variables.numberFirewallsGlobal,
                test.variables.statusGlobal,
                test.role,
              ),
              key,
            ],
            routes,
            vuetify,
          ],
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

    ///////
    // Data and Props checking
    //////

    it("Data is defined", () => {
      expect(wrapper.vm.$data).toBeDefined();
    });

    it("Compare data with default value", () => {
      expect(wrapper.vm.headers).toEqual(test.data.headers);
      expect(wrapper.vm.getNumberFirewallRules).toEqual(
        test.data.getNumberFirewallRules,
      );
      expect(wrapper.vm.page).toEqual(test.data.page);
      expect(wrapper.vm.itemsPerPage).toEqual(test.data.itemsPerPage);
      expect(wrapper.vm.next).toEqual(test.data.next);
      expect(wrapper.vm.prev).toEqual(test.data.prev);
      expect(wrapper.vm.nextPage).toEqual(test.data.nextPage);
      expect(wrapper.vm.previousPage).toEqual(test.data.previousPage);
      expect(wrapper.vm.changeItemsPerPage).toEqual(
        test.data.changeItemsPerPage,
      );
      expect(wrapper.vm.refreshFirewallRules).toEqual(
        test.data.refreshFirewallRules,
      );
      expect(wrapper.vm.loading).toEqual(test.data.loading);
    });

    it("Compare the computed with the default value", () => {
      expect(wrapper.vm.firewallRules).toEqual(test.computed.firewallRules);
      expect(wrapper.vm.getNumberFirewallRules).toEqual(
        test.computed.getNumberFirewallRules,
      );
      expect(wrapper.vm.hasAuthorizationFormDialogEdit).toEqual(
        test.computed.hasAuthorizationFormDialogEdit,
      );
      expect(wrapper.vm.hasAuthorizationFormDialogRemove).toEqual(
        test.computed.hasAuthorizationFormDialogRemove,
      );
    });

    //////
    // HTML validation
    //////

    it("Renders the template with data", () => {
      const dt = wrapper.find('[data-test="firewallRules-list"]');
      const dataTableProps = dt.attributes();

      expect(dt.exists()).toBeTruthy();
      expect(+dataTableProps.totalcount).toBe(test.variables.numberFirewallsGlobal);
    });
  });
});
