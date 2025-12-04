import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { nextTick } from "vue";
import Home from "@/views/Home.vue";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useNamespacesStore from "@/store/modules/namespaces";
import useDevicesStore from "@/store/modules/devices";
import { INamespace } from "@/interfaces/INamespace";

type HomeWrapper = VueWrapper<InstanceType<typeof Home>>;

describe("Home", () => {
  let wrapper: HomeWrapper;
  let namespacesStore: ReturnType<typeof useNamespacesStore>;
  let devicesStore: ReturnType<typeof useDevicesStore>;
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

  beforeEach(async () => {
    setActivePinia(createPinia());
    namespacesStore = useNamespacesStore();
    devicesStore = useDevicesStore();

    namespacesStore.$patch({
      namespaceList: [namespaceData],
      currentNamespace: namespaceData,
    });

    devicesStore.$patch({
      totalDevicesCount: 5,
      onlineDevicesCount: 3,
      pendingDevicesCount: 1,
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
  });

  describe("Component Rendering", () => {
    it("renders the component successfully", () => {
      expect(wrapper.exists()).toBe(true);
      expect(wrapper.html()).toMatchSnapshot();
    });

    it("renders the no namespace state correctly", async () => {
      wrapper.unmount();
      namespacesStore.$patch({
        namespaceList: [],
        currentNamespace: {},
      });

      wrapper = mount(Home, {
        global: { plugins: [vuetify, router, SnackbarPlugin] },
      });

      await flushPromises();

      expect(wrapper.exists()).toBe(true);
      expect(wrapper.text()).toContain("No Active Namespace");
      expect(wrapper.text()).toContain(
        "A namespace is a logical grouping that isolates your devices, sessions, and configurations from others.",
      );
      expect(wrapper.text()).toContain("You need to create or join a namespace");
    });

    it("displays namespace information", async () => {
      await flushPromises();

      expect(wrapper.text()).toContain("TENANT ID");
      expect(wrapper.text()).toContain(namespaceData.tenant_id);
      expect(wrapper.text()).toContain("This is your active namespace");
    });

    it("displays device stat cards", () => {
      const text = wrapper.text();
      expect(text).toContain("Accepted Devices");
      expect(text).toContain("Online Devices");
      expect(text).toContain("Pending Devices");
    });

    it("displays the add device card", () => {
      const text = wrapper.text();
      expect(text).toContain("Add a new device");
      expect(text).toContain("Register new devices to this namespace");
    });
  });

  describe("Stats Loading (From devicesStore)", () => {
    it("displays correct stat values via computed properties", async () => {
      await flushPromises();
      expect(devicesStore.totalDevicesCount).toBe(5);
      expect(devicesStore.onlineDevicesCount).toBe(3);
      expect(devicesStore.pendingDevicesCount).toBe(1);
    });

    it("updates UI when store changes", async () => {
      devicesStore.$patch({
        totalDevicesCount: 10,
        onlineDevicesCount: 8,
        pendingDevicesCount: 2,
      });

      await nextTick();
      const html = wrapper.html();

      expect(html).toContain("10");
      expect(html).toContain("8");
      expect(html).toContain("2");
    });

    it("handles missing namespace gracefully", async () => {
      wrapper.unmount();
      namespacesStore.namespaceList = [];

      wrapper = mount(Home, {
        global: { plugins: [vuetify, router, SnackbarPlugin] },
      });

      await flushPromises();

      expect(wrapper.exists()).toBe(true);
      expect(wrapper.text()).toContain("No Active Namespace");
    });

    it("reacts when namespace becomes available", async () => {
      wrapper.unmount();
      namespacesStore.namespaceList = [];

      wrapper = mount(Home, {
        global: { plugins: [vuetify, router, SnackbarPlugin] },
      });

      await flushPromises();

      namespacesStore.namespaceList = [namespaceData];
      await nextTick();
      await flushPromises();

      expect(wrapper.text()).toContain(namespaceData.tenant_id);
    });
  });

  describe("Error Handling", () => {
    it("can manually toggle error state", async () => {
      expect(wrapper.find('[data-test="home-failed"]').exists()).toBe(false);

      wrapper.vm.hasError = true;
      await nextTick();

      expect(wrapper.find('[data-test="home-failed"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="home-failed"]').text()).toContain(
        "Something is wrong, try again!",
      );
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
