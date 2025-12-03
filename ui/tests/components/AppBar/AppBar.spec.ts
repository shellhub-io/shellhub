import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { VLayout } from "vuetify/components";
import { createPinia, setActivePinia } from "pinia";
import { devicesApi, namespacesApi, systemApi } from "@/api/http";
import AppBar from "@/components/AppBar/AppBar.vue";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";
import useBillingStore from "@/store/modules/billing";
import useSupportStore from "@/store/modules/support";

const Component = {
  template: "<v-layout><AppBar /></v-layout>",
};

vi.mock("@productdevbook/chatwoot/vue", () => ({
  useChatWoot: () => ({
    setUser: vi.fn(),
    setConversationCustomAttributes: vi.fn(),
    toggle: vi.fn(),
    reset: vi.fn(),
  }),
}));

const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());
const mockSystemApi = new MockAdapter(systemApi.getAxios());
const mockDevicesApi = new MockAdapter(devicesApi.getAxios());

const billingData = {
  id: "sub_test",
  active: true,
  status: "active",
  customer_id: "cus_test",
  subscription_id: "sub_test",
  current_period_end: 999999999999,
  end_at: 999999999999,
  created_at: "",
  updated_at: "",
  invoices: [],
};

const authStoreData = {
  id: "507f1f77bcf86cd799439011",
  username: "test",
  email: "test@example.com",
  tenantId: "fake-tenant-data",
};

const systemInfo = {
  version: "v0.19.2",
  endpoints:
  {
    ssh: "localhost:2222",
    api: "localhost:8080",
  },
  setup: true,
  authentication:
  {
    local: true,
    saml: false,
  },
};

// eslint-disable-next-line vue/max-len
const mockInvitationsUrl = "http://localhost:3000/api/users/invitations?filter=W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJzdGF0dXMiLCJvcGVyYXRvciI6ImVxIiwidmFsdWUiOiJwZW5kaW5nIn19XQ%3D%3D&page=1&per_page=100";

describe("AppBar Component", () => {
  let wrapper: VueWrapper<unknown>;
  const vuetify = createVuetify();
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const billingStore = useBillingStore();
  const supportStore = useSupportStore();

  beforeEach(() => {
    window.matchMedia = vi.fn().mockImplementation((query) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }));

    envVariables.isCloud = true;
    localStorage.setItem("tenant", "fake-tenant-data");

    mockSystemApi.onGet("http://localhost:3000/info").reply(200, systemInfo);
    mockDevicesApi.onGet("http://localhost:3000/api/devices?page=1&per_page=100&status=pending").reply(200, []);
    mockDevicesApi.onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted").reply(200, []);
    mockDevicesApi.onGet("http://localhost:3000/api/stats").reply(200, {});
    mockNamespacesApi.onGet("http://localhost:3000/api/namespaces?page=1&per_page=30").reply(200, []);
    mockNamespacesApi.onGet(mockInvitationsUrl).reply(200, []);
    authStore.$patch(authStoreData);
    billingStore.billing = billingData;

    wrapper = mount(Component, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
        components: {
          "v-layout": VLayout,
          AppBar,
        },
      },
    });
  });

  afterEach(() => { wrapper.unmount(); });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders internal components", () => {
    expect(wrapper.find('[data-test="app-bar"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="menu-toggle"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="breadcrumbs"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="support-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="user-menu-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="user-icon"]').exists()).toBe(true);
  });

  it("Opens the ShellHub help page when the support button is clicked", async () => {
    mockNamespacesApi.onGet("http://localhost:3000/api/namespaces/fake-tenant-data/support").reply(200, { identifier: "fake-identifier" });

    const drawer = wrapper.findComponent(AppBar);

    const openShellhubHelpMock = vi.spyOn(drawer.vm, "openShellhubHelp");
    openShellhubHelpMock.mockImplementation(vi.fn());
    const supportBtn = wrapper.find('[data-test="support-btn"]');

    await supportBtn.trigger("click");
    expect(openShellhubHelpMock).toHaveBeenCalled();
  });

  it("Renders the logout btn", async () => {
    const drawer = wrapper.findComponent(AppBar);

    const userMenuBtn = drawer.find('[data-test="user-menu-btn"]');
    expect(userMenuBtn.exists()).toBe(true);

    await userMenuBtn.trigger("click");

    const logoutItem = drawer.findComponent('[data-test="Logout"]');
    expect(logoutItem.exists()).toBe(true);
  });

  it("Displays the correct breadcrumb titles", () => {
    const drawer = wrapper.findComponent(AppBar);

    const breadcrumbItems = drawer.findAll('[data-test="breadcrumbs"] v-breadcrumbs-item');

    const expectedBreadcrumbs = drawer.vm.breadcrumbItems;
    breadcrumbItems.forEach((item, index) => {
      expect(item.text()).toBe(expectedBreadcrumbs[index].title);
    });
  });

  it("Opens the paywall if instance is community", async () => {
    envVariables.isCloud = false;
    envVariables.isCommunity = true;
    wrapper.unmount();
    wrapper = mount(Component, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
        components: {
          "v-layout": VLayout,
          AppBar,
        },
      },
    });

    await flushPromises();
    const drawer = wrapper.findComponent(AppBar);

    await drawer.vm.openShellhubHelp();

    await flushPromises();

    expect(drawer.vm.chatSupportPaywall).toBeTruthy();
  });

  it("Uses Chatwoot if identifier is set", async () => {
    mockNamespacesApi.onGet("http://localhost:3000/api/namespaces/fake-tenant-data/support").reply(200, { identifier: "fake-identifier" });

    const drawer = wrapper.findComponent(AppBar);

    const supportBtn = wrapper.find('[data-test="support-btn"]');

    vi.spyOn(drawer.vm, "identifier", "get").mockReturnValue("mocked_identifier");

    const windowOpenMock = vi.spyOn(window, "open").mockImplementation(() => null);
    const storeSpy = vi.spyOn(supportStore, "getIdentifier");

    await supportBtn.trigger("click");

    await flushPromises();

    expect(windowOpenMock).not.toHaveBeenCalled();
    expect(storeSpy).toHaveBeenCalledWith("fake-tenant-data");
  });
});
