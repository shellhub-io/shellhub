import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useFirewallRulesStore from "@admin/store/modules/firewall_rules";
import { SnackbarPlugin } from "@/plugins/snackbar";
import routes from "../../../../src/router";
import FirewallRules from "../../../../src/views/FirewallRules.vue";

type FirewallRulesWrapper = VueWrapper<InstanceType<typeof FirewallRules>>;

describe("Firewall Rules", () => {
  let wrapper: FirewallRulesWrapper;

  beforeEach(() => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const firewallStore = useFirewallRulesStore();
    vi.spyOn(firewallStore, "getNumberFirewalls", "get").mockReturnValue(1);
    firewallStore.fetch = vi.fn();

    const snackbarStore = useSnackbarStore();
    snackbarStore.showSnackbarErrorAction = vi.fn();

    const vuetify = createVuetify();

    wrapper = mount(FirewallRules, {
      global: {
        plugins: [pinia, vuetify, routes, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Should render all the components in the screen", () => {
    expect(wrapper.find("h1").text()).toContain("Firewall Rules");
    expect(wrapper.find("[data-test='firewallRules-list']").exists()).toBe(true);
  });
});
