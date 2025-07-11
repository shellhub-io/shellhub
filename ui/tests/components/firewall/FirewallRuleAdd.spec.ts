import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { store, key } from "@/store";
import FirewallRuleAdd from "@/components/firewall/FirewallRuleAdd.vue";
import { router } from "@/router";
import { namespacesApi, rulesApi, tagsApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import { FormFilterOptions } from "@/interfaces/IFilter";

type FirewallRuleAddWrapper = VueWrapper<InstanceType<typeof FirewallRuleAdd>>;

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Firewall Rule Add", () => {
  const node = document.createElement("div");
  node.setAttribute("id", "app");
  document.body.appendChild(node);

  let wrapper: FirewallRuleAddWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;
  let mockFirewall: MockAdapter;
  let mockTags: MockAdapter;

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

  beforeEach(async () => {
    const el = document.createElement("div");
    document.body.appendChild(el);
    vi.useFakeTimers();
    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockFirewall = new MockAdapter(rulesApi.getAxios());
    mockTags = new MockAdapter(tagsApi.getAxios());

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockFirewall = new MockAdapter(rulesApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockTags.onGet("http://localhost:3000/api/tags").reply(200, []);

    store.commit("auth/authSuccess", authData);
    store.commit("namespaces/setNamespace", namespaceData);

    wrapper = mount(FirewallRuleAdd, {
      global: {
        plugins: [[store, key], vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
        config: {
          errorHandler: () => { /* ignore global error handler */ },
        },
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
    expect(dialog.find('[data-test="firewall-rule-source-ip"]').exists()).toBe(true);
    expect(dialog.find('[data-test="username-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="device-field"]').exists()).toBe(true);
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
    const storeSpy = vi.spyOn(store, "dispatch");

    mockFirewall.onPost("http://localhost:3000/api/firewall/rules").reply(200);

    await wrapper.findComponent('[data-test="firewall-add-rule-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="firewall-rule-add-btn"]').trigger("click");

    expect(storeSpy).toBeCalledWith("firewallRules/post", {
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
    mockFirewall.onPost("http://localhost:3000/api/firewall/rules").reply(400);

    await wrapper.findComponent('[data-test="firewall-add-rule-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="firewall-rule-add-btn"]').trigger("click");
    await flushPromises();

    expect(mockSnackbar.showError).toBeCalledWith("Failed to create a new firewall rule.");
  });
});
