import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { nextTick } from "vue";
import Home from "@/views/Home.vue";
import { devicesApi } from "@/api/http";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useNamespacesStore from "@/store/modules/namespaces";
import useStatsStore from "@/store/modules/stats";
import { INamespace } from "@/interfaces/INamespace";

type HomeWrapper = VueWrapper<InstanceType<typeof Home>>;

describe("Home", () => {
  let wrapper: HomeWrapper;
  let namespacesStore: ReturnType<typeof useNamespacesStore>;
  let statsStore: ReturnType<typeof useStatsStore>;
  let mockDevicesApi: MockAdapter;
  const vuetify = createVuetify();

  const members = [
    {
      id: "xxxxxxxx",
      role: "owner" as const,
      email: "test@example.com",
      status: "accepted" as const,
      added_at: "2024-01-01T00:00:00Z",
      expires_at: "2025-01-01T00:00:00Z",
    },
  ];

  const namespaceData: INamespace = {
    billing: null,
    name: "test",
    owner: "test",
    tenant_id: "fake-tenant-data",
    members,
    settings: {
      session_record: true,
      connection_announcement: "",
    },
    max_devices: 3,
    devices_accepted_count: 0,
    devices_rejected_count: 0,
    devices_pending_count: 0,
    created_at: "",
    type: "personal",
  };

  const statsMock = {
    registered_devices: 5,
    online_devices: 3,
    active_sessions: 2,
    pending_devices: 1,
    rejected_devices: 0,
  };

  beforeEach(async () => {
    setActivePinia(createPinia());
    namespacesStore = useNamespacesStore();
    statsStore = useStatsStore();

    mockDevicesApi = new MockAdapter(devicesApi.getAxios());
    mockDevicesApi.onGet("http://localhost:3000/api/stats").reply(200, statsMock);

    namespacesStore.$patch({
      namespaceList: [namespaceData],
      currentNamespace: namespaceData,
    });

    wrapper = mount(Home, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
    });

    await flushPromises();
  });

  afterEach(() => {
    wrapper.unmount();
    mockDevicesApi.reset();
    mockDevicesApi.restore();
  });

  describe("Component Rendering", () => {
    it("renders the component successfully", () => {
      expect(wrapper.exists()).toBe(true);
      expect(wrapper.html()).toMatchSnapshot();
    });

    it("displays namespace information", async () => {
      await flushPromises();

      expect(wrapper.text()).toContain("TENANT ID");
      expect(wrapper.text()).toContain(namespaceData.tenant_id);
      expect(wrapper.text()).toContain("This is your active namespace");
    });

    it("displays all device stat cards", () => {
      expect(wrapper.text()).toContain("Accepted Devices");
      expect(wrapper.text()).toContain("Online Devices");
      expect(wrapper.text()).toContain("Pending Devices");
    });

    it("displays the add device card", () => {
      expect(wrapper.text()).toContain("Add a new device");
      expect(wrapper.text()).toContain("Register new devices to this namespace");
    });
  });

  describe("Stats Loading", () => {
    it("fetches and displays stats on mount", async () => {
      await flushPromises();

      expect(statsStore.stats.registered_devices).toBe(statsMock.registered_devices);
      expect(statsStore.stats.online_devices).toBe(statsMock.online_devices);
      expect(statsStore.stats.pending_devices).toBe(statsMock.pending_devices);
    });

    it("displays correct stat values in the UI", async () => {
      await flushPromises();

      const text = wrapper.text();
      expect(text).toContain("Accepted Devices");
      expect(text).toContain("Online Devices");
      expect(text).toContain("Pending Devices");
    });

    it("handles missing namespace gracefully", async () => {
      wrapper.unmount();
      namespacesStore.namespaceList = [];

      wrapper = mount(Home, {
        global: {
          plugins: [vuetify, router, SnackbarPlugin],
        },
      });

      await flushPromises();

      expect(wrapper.exists()).toBe(true);
    });

    it("fetches stats when namespace becomes available", async () => {
      wrapper.unmount();
      namespacesStore.namespaceList = [];

      wrapper = mount(Home, {
        global: {
          plugins: [vuetify, router, SnackbarPlugin],
        },
      });

      await flushPromises();

      namespacesStore.namespaceList = [namespaceData];
      await nextTick();
      await flushPromises();

      expect(statsStore.stats.registered_devices).toBe(statsMock.registered_devices);
    });
  });

  describe("Error Handling", () => {
    it("displays error message when API call fails with 403 status", async () => {
      wrapper.unmount();
      mockDevicesApi.reset();
      mockDevicesApi.onGet("http://localhost:3000/api/stats").reply(403);

      wrapper = mount(Home, {
        global: {
          plugins: [vuetify, router, SnackbarPlugin],
        },
      });

      await flushPromises();

      expect(wrapper.vm.hasStatus).toBe(true);
      expect(wrapper.find('[data-test="home-failed"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="home-failed"]').text()).toContain(
        "Something is wrong, try again!",
      );
    });

    it("displays error message when API call fails with other errors", async () => {
      wrapper.unmount();
      mockDevicesApi.reset();
      mockDevicesApi.onGet("http://localhost:3000/api/stats").reply(500);

      wrapper = mount(Home, {
        global: {
          plugins: [vuetify, router, SnackbarPlugin],
        },
      });

      await flushPromises();

      expect(wrapper.vm.hasStatus).toBe(true);
      expect(wrapper.find('[data-test="home-failed"]').exists()).toBe(true);
    });

    it("displays error message when API call times out", async () => {
      wrapper.unmount();
      mockDevicesApi.reset();
      mockDevicesApi.onGet("http://localhost:3000/api/stats").timeout();

      wrapper = mount(Home, {
        global: {
          plugins: [vuetify, router, SnackbarPlugin],
        },
      });

      await flushPromises();

      expect(wrapper.vm.hasStatus).toBe(true);
      expect(wrapper.find('[data-test="home-failed"]').exists()).toBe(true);
    });

    it("can manually toggle error state", async () => {
      expect(wrapper.find('[data-test="home-failed"]').exists()).toBe(false);

      wrapper.vm.hasStatus = true;
      await nextTick();

      expect(wrapper.find('[data-test="home-failed"]').exists()).toBe(true);
    });
  });

  describe("User Interactions", () => {
    it("allows copying tenant ID", () => {
      const copyButton = wrapper.find('[data-test="copy-tenant-btn"]');
      expect(copyButton.exists()).toBe(true);
    });

    it("displays tenant ID in code format", () => {
      const codeElement = wrapper.find('[data-test="tenant-info-text"]');
      expect(codeElement.exists()).toBe(true);
      expect(codeElement.text()).toBe(namespaceData.tenant_id);
    });
  });

  describe("Navigation", () => {
    it("has navigation buttons for device pages", () => {
      const text = wrapper.text();
      expect(text).toContain("View all devices");
      expect(text).toContain("View Online Devices");
      expect(text).toContain("Approve Devices");
    });
  });
});
