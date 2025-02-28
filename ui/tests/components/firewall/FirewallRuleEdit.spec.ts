import { setActivePinia, createPinia } from "pinia";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi, afterEach } from "vitest";
import { nextTick } from "vue";
import FirewallRuleEdit from "@/components/firewall/FirewallRuleEdit.vue";
import { rulesApi, tagsApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import { IFirewallRule } from "@/interfaces/IFirewallRule";
import useFirewallRulesStore from "@/store/modules/firewall_rules";

type FirewallRuleEditWrapper = VueWrapper<InstanceType<typeof FirewallRuleEdit>>;

const firewallRule = {
  id: "1000",
  tenant_id: "fake-tenant-data",
  priority: 1,
  action: "allow" as const,
  active: true,
  source_ip: ".*",
  username: ".*",
  status: "active",
  filter: {
    hostname: ".*",
  },
};

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Firewall Rule Edit", () => {
  let wrapper: FirewallRuleEditWrapper;
  const vuetify = createVuetify();
  localStorage.setItem("tenant", "fake-tenant-data");

  setActivePinia(createPinia());
  const firewallRulesStore = useFirewallRulesStore();
  const mountWrapper = (firewallRuleProp: IFirewallRule = firewallRule) => mount(FirewallRuleEdit, {
    global: {
      plugins: [vuetify],
      provide: { [SnackbarInjectionKey]: mockSnackbar },
    },
    props: {
      firewallRule: firewallRuleProp,
      hasAuthorization: true,
    },
  });

  const mockTagsApi = new MockAdapter(tagsApi.getAxios());
  const mockRulesApi = new MockAdapter(rulesApi.getAxios());

  beforeEach(async () => {
    mockTagsApi
      .onGet("http://localhost:3000/api/namespaces/fake-tenant-data/tags?filter=&page=1&per_page=10")
      .reply(200, [{ name: "1" }, { name: "2" }]);

    wrapper = mountWrapper();
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the dialog open button and other key elements", async () => {
    const dialog = new DOMWrapper(document.body);

    expect(wrapper.find('[data-test="firewall-edit-rule-btn"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="firewall-edit-rule-btn"]').trigger("click");

    expect(dialog.find('[data-test="firewall-edit-rule-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-status"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-priority"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-policy"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-source-ip-select"]').exists()).toBe(true);
    expect(dialog.find('[data-test="username-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="filter-select"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-cancel"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-edit-btn"]').exists()).toBe(true);
  });

  it("Conditional rendering components", async () => {
    wrapper = mountWrapper({ ...firewallRule, source_ip: "127.0.0.1", username: "ossystems", filter: { tags: ["tag1", "tag2"] } });

    const dialog = new DOMWrapper(document.body);
    await wrapper.findComponent('[data-test="firewall-edit-rule-btn"]').trigger("click");
    await nextTick();

    expect(dialog.find('[data-test="firewall-rule-source-ip"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-username-restriction"]').exists()).toBe(true);
    expect(dialog.find('[data-test="tags-selector"]').exists()).toBe(true);
  });

  it("Conditional rendering components (Hostname)", async () => {
    wrapper = mountWrapper({ ...firewallRule, filter: { hostname: "host" } });

    const dialog = new DOMWrapper(document.body);
    await wrapper.findComponent('[data-test="firewall-edit-rule-btn"]').trigger("click");
    await nextTick();
    expect(dialog.find('[data-test="firewall-rule-hostname-restriction"]').exists()).toBe(true);
  });

  it("Successful on editing firewall rules", async () => {
    const storeSpy = vi.spyOn(firewallRulesStore, "updateFirewallRule");

    mockRulesApi.onPut("http://localhost:3000/api/firewall/rules/1000").reply(200);

    await wrapper.findComponent('[data-test="firewall-edit-rule-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="firewall-rule-edit-btn"]').trigger("click");

    expect(storeSpy).toBeCalledWith({
      id: "1000",
      action: "allow",
      priority: 1,
      active: true,
      source_ip: ".*",
      username: ".*",
      filter: {
        hostname: ".*",
      },
    });
  });

  it("Fails on editing firewall rules", async () => {
    mockRulesApi.onPut("http://localhost:3000/api/firewall/rules/1000").reply(403);

    await wrapper.findComponent('[data-test="firewall-edit-rule-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="firewall-rule-edit-btn"]').trigger("click");
    await flushPromises();

    expect(mockSnackbar.showError).toBeCalledWith("Error while updating firewall rule.");
  });
});
