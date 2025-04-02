import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import FirewallRulesList from "../../../../../src/components/FirewallRules/FirewallRulesList.vue";
import { key } from "../../../../../src/store";
import routes from "../../../../../src/router";

type FirewallRulesListWrapper = VueWrapper<InstanceType<typeof FirewallRulesList>>;

const headers = [
  { text: "Tenant Id", value: "tenant_id" },
  { text: "Priority", value: "priority" },
  { text: "Action", value: "action" },
  { text: "Source Ip", value: "source_ip" },
  { text: "Username", value: "username" },
  { text: "Filter", value: "filter" },
  { text: "Actions", value: "actions" },
];

const firewallRules = [
  {
    action: "allow",
    active: true,
    filter: {
      tag: ["xxxx", "yyyy"],
    },
    id: "5f1996c84d2190a22d5857bb",
    tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    priority: 4,
    source_ip: "127.0.0.1",
    tenant_ip: "127.0.0.1",
    username: "shellhub",
  },
];

const firewallRulesTest = firewallRules;

const store = createStore({
  state: {
    firewallRules,
  },
  getters: {
    "firewallRules/list": (state) => state.firewallRules,
    "firewallRules/numberFirewalls": (state) => state.firewallRules.length,
  },
  actions: {
    "firewallRules/fetch": vi.fn(),
  },
});

describe("Firewall Rules List", () => {
  let wrapper: FirewallRulesListWrapper;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(FirewallRulesList, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", () => {
    const dt = wrapper.find("[data-test]");
    expect(dt.attributes()["data-test"]).toContain("firewallRules-list");
    expect(wrapper.vm.headers).toEqual(headers);
    expect(wrapper.vm.loading).toEqual(false);
    expect(wrapper.vm.itemsPerPage).toEqual(10);
    expect(wrapper.vm.page).toEqual(1);
  });

  it("Renders data in the computed", async () => {
    const firewallRules = await wrapper.vm.firewallRules;
    expect(firewallRules).toEqual(firewallRulesTest);
  });

  it('should show "Any Ip" when column is "source Ip" and regex is ".*"', () => {
    expect(wrapper.vm.formatSourceIP(".*")).toEqual("Any IP");
  });

  it("should show the source ip when a rule is passed to firewall", () => {
    expect(wrapper.vm.formatSourceIP("127.0.0.1")).toEqual("127.0.0.1");
  });

  it('should show "All Users" when column is "Username" and regex is ".*"', () => {
    expect(wrapper.vm.formatUsername(".*")).toEqual("All users");
  });

  it('should show "All Devices" when column is "Filter" and regex is ".*"', () => {
    expect(wrapper.vm.formatHostnameFilter({ hostname: ".*" })).toEqual("All devices");
  });

  it("must show only 10 characters in the tag when it has more than 10 characters", () => {
    expect(wrapper.vm.displayOnlyTenCharacters("very big word in tag")).toEqual("very big w...");
  });

  it('should return "false" when the function receives a valid string and is less than 10 characters', () => {
    expect(wrapper.vm.showTag("test tag")).toBeFalsy();
  });

  it('should return "true" when the function receives a valid string and is more than 10 characters', () => {
    expect(wrapper.vm.showTag("very big word in tag")).toBeTruthy();
  });
});
