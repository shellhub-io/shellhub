import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import FirewallRules from "@admin/views/FirewallRules.vue";

describe("FirewallRules", () => {
  let wrapper: VueWrapper<InstanceType<typeof FirewallRules>>;

  const mountWrapper = async () => {
    const router = createCleanAdminRouter();
    await router.push({ name: "firewall-rules" });
    await router.isReady();

    wrapper = mountComponent(FirewallRules, { global: { plugins: [router] } });
  };

  beforeEach(() => mountWrapper());

  afterEach(() => { wrapper?.unmount(); });

  it("displays the page header with correct title", () => {
    expect(wrapper.text()).toContain("Firewall Rules");
    expect(wrapper.text()).toContain("Security Controls");
  });

  it("displays the page header description", () => {
    expect(wrapper.text()).toContain("Review every policy applied across namespaces and confirm access is locked down.");
  });

  it("displays the firewall rules list component", () => {
    expect(wrapper.find('[data-test="firewall-rules-list"]').exists()).toBe(true);
  });
});
