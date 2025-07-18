import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi, afterEach } from "vitest";
import { nextTick } from "vue";
import { store, key } from "@/store";
import FirewallRuleEdit from "@/components/firewall/FirewallRuleEdit.vue";
import { envVariables } from "@/envVariables";
import { router } from "@/router";
import { namespacesApi, rulesApi, tagsApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import { IFirewallRule } from "@/interfaces/IFirewallRule";

type FirewallRuleEditWrapper = VueWrapper<InstanceType<typeof FirewallRuleEdit>>;

const members = [
  {
    id: "xxxxxxxx",
    username: "test",
    role: "owner",
  },
];

const namespaceData = {
  name: "user",
  owner: "xxxxxxxx",
  tenant_id: "fake-tenant-data",
  members,
  max_devices: 3,
  devices_count: 3,
  devices: 2,
  created_at: "",
};

const authData = {
  status: "",
  token: "",
  user: "test",
  name: "test",
  tenant: "fake-tenant-data",
  email: "test@test.com",
  id: "xxxxxxxx",
  role: "owner",
};

const firewallRule = {
  id: "1000",
  tenant_id: "00000000-0000-4000-0000-000000000000",
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

  const mountWrapper = (firewallRuleProp: IFirewallRule = firewallRule) => mount(FirewallRuleEdit, {
    global: {
      plugins: [[store, key], vuetify, router],
      provide: { [SnackbarInjectionKey]: mockSnackbar },
    },
    props: {
      firewallRule: firewallRuleProp,
      hasAuthorization: true,
    },
  });

  let mockNamespace: MockAdapter;
  let mockTags: MockAdapter;
  let mockFirewall: MockAdapter;

  beforeEach(async () => {
    const el = document.createElement("div");
    document.body.appendChild(el);
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockFirewall = new MockAdapter(rulesApi.getAxios());
    mockTags = new MockAdapter(tagsApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockTags.onGet("http://localhost:3000/api/tags").reply(200, ["tag1", "tag2"]);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);

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
    const storeSpy = vi.spyOn(store, "dispatch");

    mockFirewall.onPut("http://localhost:3000/api/firewall/rules/1000").reply(200);

    await wrapper.findComponent('[data-test="firewall-edit-rule-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="firewall-rule-edit-btn"]').trigger("click");

    expect(storeSpy).toBeCalledWith("firewallRules/put", {
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
    mockFirewall.onPut("http://localhost:3000/api/firewall/rules/1000").reply(403);

    await wrapper.findComponent('[data-test="firewall-edit-rule-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="firewall-rule-edit-btn"]').trigger("click");
    await flushPromises();

    expect(mockSnackbar.showError).toBeCalledWith("Error while updating firewall rule.");
  });
});
