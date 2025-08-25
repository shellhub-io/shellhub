import { setActivePinia, createPinia } from "pinia";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import FirewallRuleDelete from "@/components/firewall/FirewallRuleDelete.vue";
import { router } from "@/router";
import { rulesApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useFirewallRulesStore from "@/store/modules/firewall_rules";

type FirewallRuleDeleteWrapper = VueWrapper<InstanceType<typeof FirewallRuleDelete>>;

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Firewall Rule Delete", () => {
  let wrapper: FirewallRuleDeleteWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const firewallRulesStore = useFirewallRulesStore();
  const mockRulesApi = new MockAdapter(rulesApi.getAxios());

  beforeEach(async () => {
    wrapper = mount(FirewallRuleDelete, {
      global: {
        plugins: [vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: {
        id: "1000",
        hasAuthorization: true,
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

    expect(wrapper.find('[data-test="remove-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="remove-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-delete-dialog-btn"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="firewall-delete-dialog-btn"]').trigger("click");

    expect(dialog.find('[data-test="firewallRuleDelete-card"]').exists()).toBe(true);
    expect(dialog.find('[data-test="text-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="text-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="remove-btn"]').exists()).toBe(true);
  });

  it("Successful on removing firewall rules", async () => {
    const storeSpy = vi.spyOn(firewallRulesStore, "removeFirewallRule");

    mockRulesApi.onDelete("http://localhost:3000/api/firewall/rules/1000").reply(200);

    await wrapper.findComponent('[data-test="firewall-delete-dialog-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="remove-btn"]').trigger("click");

    expect(storeSpy).toBeCalledWith("1000");
  });

  it("Fails on removing firewall rules", async () => {
    mockRulesApi.onDelete("http://localhost:3000/api/firewall/rules/1000").reply(403);

    await wrapper.findComponent('[data-test="firewall-delete-dialog-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="remove-btn"]').trigger("click");
    await flushPromises();
    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to delete firewall rule.");
  });
});
