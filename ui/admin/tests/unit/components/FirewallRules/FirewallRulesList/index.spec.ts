import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useFirewallRulesStore from "@admin/store/modules/firewall_rules";
import { SnackbarPlugin } from "@/plugins/snackbar";
import FirewallRulesList from "../../../../../src/components/FirewallRules/FirewallRulesList.vue";
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
    action: "allow" as const,
    active: true,
    filter: {
      tags: [
        {
          tenant_id: "fake-tenant-data",
          name: "test-tag",
          created_at: "",
          updated_at: "",
        },
      ],
    },
    id: "5f1996c84d2190a22d5857bb",
    tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    priority: 4,
    source_ip: "127.0.0.1",
    username: "shellhub",
  },
];

describe("Firewall Rules List", () => {
  let wrapper: FirewallRulesListWrapper;

  beforeEach(() => {
    setActivePinia(createPinia());

    const vuetify = createVuetify();
    const firewallRulesStore = useFirewallRulesStore();

    firewallRulesStore.firewalls = firewallRules;
    firewallRulesStore.numberFirewalls = firewallRules.length;
    firewallRulesStore.fetch = vi.fn();

    wrapper = mount(FirewallRulesList, {
      global: {
        plugins: [vuetify, routes, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", () => {
    const dt = wrapper.find("[data-test]");
    expect(dt.attributes()["data-test"]).toContain("firewall-rules-list");
    expect(wrapper.vm.headers).toEqual(headers);
    expect(wrapper.vm.loading).toEqual(false);
    expect(wrapper.vm.itemsPerPage).toEqual(10);
    expect(wrapper.vm.page).toEqual(1);
  });

  it("Renders data in the computed", async () => {
    expect(wrapper.vm.firewallRules).toEqual(firewallRules);
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
