import { setActivePinia, createPinia } from "pinia";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import FirewallRuleDelete from "@/components/Firewall/FirewallRuleDelete.vue";
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

  beforeEach(() => {
    vi.clearAllMocks(); // important: reset spies between tests

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

  it("Successful on removing firewall rules", async () => {
    const storeSpy = vi.spyOn(firewallRulesStore, "removeFirewallRule").mockResolvedValue();

    mockRulesApi.onDelete("http://localhost:3000/api/firewall/rules/1000").reply(200);

    await wrapper.find('[data-test="firewall-delete-dialog-btn"]').trigger("click");

    const dialog = wrapper.findComponent({ name: "MessageDialog" });
    await dialog.vm.$emit("confirm");

    expect(storeSpy).toBeCalledWith("1000");
  });

  it("Fails on removing firewall rules", async () => {
    vi.spyOn(firewallRulesStore, "removeFirewallRule").mockRejectedValue(new Error("403"));

    mockRulesApi.onDelete("http://localhost:3000/api/firewall/rules/1000").reply(403);

    await wrapper.find('[data-test="firewall-delete-dialog-btn"]').trigger("click");

    const dialog = wrapper.findComponent({ name: "MessageDialog" });
    await dialog.vm.$emit("confirm");

    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to delete firewall rule.");
  });
});
