import { createVuetify } from "vuetify";
import { flushPromises, mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useFirewallRulesStore from "@admin/store/modules/firewall_rules";
import routes from "@admin/router";
import FirewallRulesDetails from "@admin/views/FirewallRulesDetails.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

const mockFirewallRule = {
  id: "6256b876e5c1d9bbdf954662",
  tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  priority: 1,
  action: "allow" as const,
  active: true,
  source_ip: ".*",
  username: "^[A-a]",
  filter: { hostname: ".*" },
};

const mockFirewallRuleWithTags = {
  id: "6256b876e5c1d9bbdf954663",
  tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  priority: 2,
  action: "deny" as const,
  active: false,
  source_ip: "192.168.1.0/24",
  username: "admin",
  filter: {
    tags: [
      { name: "production" },
      { name: "web-server" },
      { name: "critical-infrastructure" },
    ],
  },
};

const mockRoute = { params: { id: mockFirewallRule.id } };

describe("Firewall Rule Details", async () => {
  const pinia = createPinia();
  setActivePinia(pinia);

  const firewallStore = useFirewallRulesStore();
  firewallStore.fetchFirewallRuleById = vi.fn().mockResolvedValue(mockFirewallRule);

  const vuetify = createVuetify();

  const wrapper = mount(FirewallRulesDetails, {
    global: {
      plugins: [pinia, vuetify, routes, SnackbarPlugin],
      mocks: { $route: mockRoute },
    },
  });

  await flushPromises();

  it("Displays the rule priority in the card title", () => {
    expect(wrapper.find(".text-h6").text()).toBe(`Rule #${mockFirewallRule.priority}`);
  });

  it("Shows active status icon with tooltip", () => {
    const icon = wrapper.find('[data-test="active-icon"]');
    expect(icon.classes()).toContain("text-success");
  });

  it("Shows action chip with correct color", () => {
    const actionChip = wrapper.find('[data-test="firewall-action-chip"]');
    expect(actionChip.text()).toBe(mockFirewallRule.action);
  });

  it("Displays firewall rule ID", () => {
    const idField = wrapper.find('[data-test="firewall-id-field"]');
    expect(idField.text()).toContain("ID:");
    expect(idField.text()).toContain(mockFirewallRule.id);
  });

  it("Displays priority", () => {
    const priorityField = wrapper.find('[data-test="firewall-priority-field"]');
    expect(priorityField.text()).toContain("Priority:");
    expect(priorityField.text()).toContain(String(mockFirewallRule.priority));
  });

  it("Displays namespace with router link", () => {
    const tenantField = wrapper.find('[data-test="firewall-tenant-field"]');
    expect(tenantField.text()).toContain("Namespace:");
    expect(tenantField.find("a").exists()).toBe(true);
  });

  it("Displays formatted source IP", () => {
    const sourceIpField = wrapper.find('[data-test="firewall-source-ip-field"]');
    expect(sourceIpField.text()).toContain("Source IP:");
    expect(sourceIpField.text()).toContain("Any IP");
  });

  it("Displays formatted username", () => {
    const usernameField = wrapper.find('[data-test="firewall-username-field"]');
    expect(usernameField.text()).toContain("Username:");
    expect(usernameField.text()).toContain(mockFirewallRule.username);
  });

  it("Displays filter information", () => {
    const filterField = wrapper.find('[data-test="firewall-filter-field"]');
    expect(filterField.text()).toContain("Filter:");
  });

  it("Shows error message when firewall rule data is empty", async () => {
    firewallStore.fetchFirewallRuleById = vi.fn().mockResolvedValue({});

    const errorWrapper = mount(FirewallRulesDetails, {
      global: {
        plugins: [pinia, vuetify, routes, SnackbarPlugin],
        mocks: { $route: mockRoute },
      },
    });

    await flushPromises();

    expect(errorWrapper.text()).toContain("Something is wrong, try again!");
  });

  it("Displays tags correctly when filter contains tags", async () => {
    firewallStore.fetchFirewallRuleById = vi.fn().mockResolvedValue(mockFirewallRuleWithTags);

    const tagsWrapper = mount(FirewallRulesDetails, {
      global: {
        plugins: [pinia, vuetify, routes, SnackbarPlugin],
        mocks: { $route: { params: { id: mockFirewallRuleWithTags.id } } },
      },
    });

    await flushPromises();

    const filterField = tagsWrapper.find('[data-test="firewall-filter-field"]');

    const chips = filterField.findAll(".v-chip");
    expect(chips).toHaveLength(3);
    expect(chips[0].text()).toContain("production");
    expect(chips[1].text()).toContain("web-server");
    expect(chips[2].text()).toContain("critical-i");
  });
});
