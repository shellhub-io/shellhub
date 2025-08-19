import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import SettingNamespace from "@/components/Setting/SettingNamespace.vue";
import { store, key } from "@/store";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { router } from "@/router";
import { namespacesApi, usersApi } from "@/api/http";
import useAuthStore from "@/store/modules/auth";

type SettingNamespaceWrapper = VueWrapper<InstanceType<typeof SettingNamespace>>;

describe("Setting Namespace", () => {
  let wrapper: SettingNamespaceWrapper;
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const vuetify = createVuetify();
  const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());
  const mockUsersApi = new MockAdapter(usersApi.getAxios());

  const members = [
    {
      id: "507f1f77bcf86cd799439011",
      role: "owner",
    },
  ];

  const billingData = {
    active: false,
    status: "inactive",
    customer_id: "cus_test",
    subscription_id: "sub_test",
    current_period_end: 2068385820,
    created_at: "",
    updated_at: "",
    invoices: [],
  };

  const namespaceData = {
    data: {
      name: "test",
      owner: "test",
      tenant_id: "fake-tenant",
      members,
      billing: billingData,
      settings: {
        session_record: true,
        connection_announcement: "",
      },
      max_devices: 3,
      devices_count: 3,
      created_at: "",
    },
  };

  beforeEach(async () => {
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

    localStorage.setItem("tenant", "fake-tenant");
    mockNamespacesApi.onGet("http://localhost:3000/api/namespaces/fake-tenant").reply(200, namespaceData.data);
    mockUsersApi.onGet("http://localhost:3000/api/users/security").reply(200, true);

    authStore.tenantId = "fake-tenant";

    wrapper = mount(SettingNamespace, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  const dataTests = [
    "card",
    "card-header",
    "card-title",
    "card-subtitle",
    "edit-namespace-btn",
    "profile-details-list",
    "profile-details-item",
    "name-icon",
    "name-title",
    "name-input",
    "tenant-details-item",
    "tenant-icon",
    "tenant-title",
    "tenant-copy-btn",
    "announcement-item",
    "announcement-icon",
    "announcement-title",
    "announcement-subtitle",
    "edit-announcement-btn",
    "record-item",
    "record-icon",
    "record-title",
    "record-description",
    "session-recording-setting-component",
    "delete-leave-item",
    "delete-leave-icon",
    "delete-leave-title",
    "leave-description",
    "leave-namespace-btn",
  ];

  dataTests.forEach((dataTest) => {
    it(`should render the element with data-test="${dataTest}"`, () => {
      const element = wrapper.find(`[data-test="${dataTest}"]`);
      expect(element.exists()).toBe(true);
    });
  });
});
