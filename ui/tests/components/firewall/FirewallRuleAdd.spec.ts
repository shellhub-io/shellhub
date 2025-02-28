import { createPinia, setActivePinia } from "pinia";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import FirewallRuleAdd from "@/components/firewall/FirewallRuleAdd.vue";
import { router } from "@/router";
import { rulesApi, tagsApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import { FormFilterOptions } from "@/interfaces/IFilter";
import useAuthStore from "@/store/modules/auth";
import useFirewallRulesStore from "@/store/modules/firewall_rules";

type FirewallRuleAddWrapper = VueWrapper<InstanceType<typeof FirewallRuleAdd>>;

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Firewall Rule Add", () => {
  let wrapper: FirewallRuleAddWrapper;
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const firewallRulesStore = useFirewallRulesStore();
  const vuetify = createVuetify();

  const mockRulesApi = new MockAdapter(rulesApi.getAxios());
  const mockTags = new MockAdapter(tagsApi.getAxios());
  beforeEach(async () => {
    localStorage.setItem("tenant", "fake-tenant-data");

    mockTags
      .onGet("http://localhost:3000/api/namespaces/fake-tenant-data/tags?filter=&page=1&per_page=10")
      .reply(200, [{ name: "1" }, { name: "2" }]);

    authStore.role = "owner";

    wrapper = mount(FirewallRuleAdd, {
      global: {
        plugins: [vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the dialog open button and other key elements", async () => {
    const dialog = new DOMWrapper(document.body);

    expect(wrapper.find('[data-test="firewall-add-rule-btn"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="firewall-add-rule-btn"]').trigger("click");

    expect(dialog.find('[data-test="firewall-rule-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-status"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-priority"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-policy"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-source-ip-select"]').exists()).toBe(true);
    expect(dialog.find('[data-test="username-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="filter-select"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-cancel"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-add-btn"]').exists()).toBe(true);
  });

  it("Conditional rendering components", async () => {
    const dialog = new DOMWrapper(document.body);

    wrapper.vm.selectedIPOption = "restrict";
    wrapper.vm.selectedUsernameOption = "username";
    wrapper.vm.selectedFilterOption = FormFilterOptions.Tags;

    await wrapper.findComponent('[data-test="firewall-add-rule-btn"]').trigger("click");

    expect(dialog.find('[data-test="firewall-rule-source-ip"]').exists()).toBe(true);
    expect(dialog.find('[data-test="firewall-rule-username-restriction"]').exists()).toBe(true);
    expect(dialog.find('[data-test="tags-selector"]').exists()).toBe(true);
  });

  it("Conditional rendering components (Hostname)", async () => {
    const dialog = new DOMWrapper(document.body);

    wrapper.vm.selectedFilterOption = FormFilterOptions.Hostname;

    await wrapper.findComponent('[data-test="firewall-add-rule-btn"]').trigger("click");

    expect(dialog.find('[data-test="firewall-rule-hostname-restriction"]').exists()).toBe(true);
  });

  it("Successful on adding firewall rules", async () => {
    const storeSpy = vi.spyOn(firewallRulesStore, "createFirewallRule");

    mockRulesApi.onPost("http://localhost:3000/api/firewall/rules").reply(200);

    await wrapper.findComponent('[data-test="firewall-add-rule-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="firewall-rule-add-btn"]').trigger("click");

    expect(storeSpy).toBeCalledWith({
      active: true,
      action: "allow",
      priority: 1,
      source_ip: ".*",
      username: ".*",
      filter: {
        hostname: ".*",
      },
    });
  });

  it("Fails on adding firewall rules", async () => {
    mockRulesApi.onPost("http://localhost:3000/api/firewall/rules").reply(404);

    await wrapper.findComponent('[data-test="firewall-add-rule-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="firewall-rule-add-btn"]').trigger("click");
    await flushPromises();

    expect(mockSnackbar.showError).toBeCalledWith("Failed to create a new firewall rule.");
  });
});
