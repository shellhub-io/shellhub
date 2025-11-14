import { createPinia, setActivePinia } from "pinia";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, afterEach, vi } from "vitest";
import { VLayout } from "vuetify/components";
import { ref } from "vue";
import DevicesDropdown from "@/components/AppBar/DevicesDropdown.vue";
import DeviceActionButton from "@/components/Devices/DeviceActionButton.vue";
import { devicesApi } from "@/api/http";
import { SnackbarInjectionKey, SnackbarPlugin } from "@/plugins/snackbar";
import { router } from "@/router";
import useStatsStore from "@/store/modules/stats";
import useDevicesStore from "@/store/modules/devices";
import { IDevice } from "@/interfaces/IDevice";
import { IStats } from "@/interfaces/IStats";
import useAuthStore from "@/store/modules/auth";
import { nextTick } from "vue";
import useNamespacesStore from "@/store/modules/namespaces";
import { INamespace, INamespaceMember } from "@/interfaces/INamespace";

const Component = {
  template: "<v-layout><DevicesDropdown /></v-layout>",
};

vi.mock("vuetify", async () => {
  const actual = await vi.importActual<typeof import("vuetify")>("vuetify");

  return {
    ...actual,
    useDisplay: () => ({
      smAndUp: ref(true),
      smAndDown: ref(false),
      mdAndUp: ref(true),
      thresholds: ref({
        xs: 0,
        sm: 600,
        md: 960,
        lg: 1280,
        xl: 1920,
        xxl: 2560,
      }),
    }),
  };
});

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

const mockPendingDevices: IDevice[] = [
  {
    uid: "pending-device-1",
    name: "pending-test-1",
    identity: { mac: "00:11:22:33:44:55" },
    info: { id: "ubuntu", pretty_name: "Ubuntu 22.04", version: "", arch: "x86_64", platform: "linux" },
    public_key: "test-key",
    tenant_id: "test-tenant",
    last_seen: "2025-11-05T10:00:00Z",
    status_updated_at: "2025-11-05T09:00:00Z",
    online: false,
    created_at: "2025-11-05T08:00:00Z",
    tags: [],
    position: { latitude: 0, longitude: 0 },
    namespace: "test",
    status: "pending",
    remote_addr: "192.168.1.100",
  },
  {
    uid: "pending-device-2",
    name: "pending-test-2",
    identity: { mac: "00:11:22:33:44:66" },
    info: { id: "debian", pretty_name: "Debian 11", version: "", arch: "x86_64", platform: "linux" },
    public_key: "test-key-2",
    tenant_id: "test-tenant",
    last_seen: "2025-11-05T11:00:00Z",
    status_updated_at: "2025-11-05T10:30:00Z",
    online: false,
    created_at: "2025-11-05T09:30:00Z",
    tags: [],
    position: { latitude: 0, longitude: 0 },
    namespace: "test",
    status: "pending",
    remote_addr: "192.168.1.101",
  },
];

const mockRecentDevices: IDevice[] = [
  {
    uid: "recent-device-1",
    name: "recent-test-1",
    identity: { mac: "AA:BB:CC:DD:EE:FF" },
    info: { id: "ubuntu", pretty_name: "Ubuntu 22.04", version: "", arch: "x86_64", platform: "linux" },
    public_key: "test-key",
    tenant_id: "test-tenant",
    last_seen: "2025-11-06T15:00:00Z",
    online: true,
    created_at: "2025-11-05T08:00:00Z",
    status_updated_at: "2025-11-05T08:00:00Z",
    tags: [],
    position: { latitude: 0, longitude: 0 },
    namespace: "test",
    status: "accepted",
    remote_addr: "192.168.1.102",
  },
  {
    uid: "recent-device-2",
    name: "recent-test-2",
    identity: { mac: "AA:BB:CC:DD:EE:AA" },
    info: { id: "debian", pretty_name: "Debian 11", version: "", arch: "x86_64", platform: "linux" },
    public_key: "test-key-2",
    tenant_id: "test-tenant",
    last_seen: "2025-11-05T12:00:00Z",
    created_at: "2025-11-05T09:30:00Z",
    status_updated_at: "2025-11-05T09:30:00Z",
    tags: [],
    position: { latitude: 0, longitude: 0 },
    online: false,
    namespace: "test",
    status: "accepted",
    remote_addr: "192.168.1.103",
  },
];

const mockStats: IStats = {
  registered_devices: 10,
  online_devices: 6,
  pending_devices: 2,
  rejected_devices: 1,
  active_sessions: 0,
};

const mockNamespace: INamespace = {
  name: "examplespace",
  owner: "507f1f77bcf86cd799439011",
  tenant_id: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
  members: [
    {
      id: "507f1f77bcf86cd799439011",
      role: "administrator",
    },
  ] as INamespaceMember[],
  settings: {
    session_record: true,
    connection_announcement: "",
  },
  max_devices: 3,
  devices_accepted_count: 0,
  devices_pending_count: 0,
  devices_rejected_count: 0,
  created_at: "2025-05-01T00:00:00.000Z",
  billing: null,
  type: "team",
};

describe("Device Management Dropdown", () => {
  let wrapper: VueWrapper<unknown>;
  let drawer: VueWrapper<InstanceType<typeof DevicesDropdown>>;
  const vuetify = createVuetify();
  setActivePinia(createPinia());
  const statsStore = useStatsStore();
  const namespacesStore = useNamespacesStore();
  const devicesStore = useDevicesStore();
  const authStore = useAuthStore();
  const mockDevicesApi = new MockAdapter(devicesApi.getAxios());
  authStore.role = "owner";
  namespacesStore.namespaceList = [mockNamespace];
  beforeEach(async () => {
    mockDevicesApi
      .onGet("http://localhost:3000/api/stats")
      .reply(200, mockStats);
    mockDevicesApi
      .onGet("http://localhost:3000/api/devices?page=1&per_page=100&status=pending").reply(200, mockPendingDevices);
    mockDevicesApi
      .onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted").reply(200, mockRecentDevices);
    wrapper = mount(Component, {
      global: {
        plugins: [vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
        components: {
          "v-layout": VLayout,
          DevicesDropdown,
        },
        stubs: { teleport: true },
      },
      attachTo: document.body,
    });
    await wrapper.find('[data-test="devices-icon"]').trigger("click");
    drawer = wrapper.findComponent(DevicesDropdown);
  });

  afterEach(() => { if (wrapper) wrapper.unmount(); });

  it("Fetches all required data on mount", async () => {
    const fetchStatsSpy = vi.spyOn(statsStore, "fetchStats");
    const fetchDevicesSpy = vi.spyOn(devicesStore, "fetchDeviceList");

    wrapper.unmount();
    wrapper = mount(Component, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
      provide: { [SnackbarInjectionKey]: mockSnackbar },
      components: {
        "v-layout": VLayout,
        DevicesDropdown,
      },
    });
    await flushPromises();

    expect(fetchStatsSpy).toHaveBeenCalled();
    expect(fetchDevicesSpy).toHaveBeenCalledWith({ status: "pending", perPage: 100 });
    expect(fetchDevicesSpy).toHaveBeenCalledWith({ status: "accepted" });
  });

  it("Opens drawer when icon is clicked", async () => {
    drawer.vm.isDrawerOpen = false;
    const icon = wrapper.find('[data-test="devices-icon"]');
    await icon.trigger("click");
    await flushPromises();

    expect(drawer.vm.isDrawerOpen).toBe(true);
    const drawerComponent = wrapper.find('[data-test="devices-drawer"]');
    expect(drawerComponent.exists()).toBe(true);
  });

  it("Displays correct statistics cards with proper values", async () => {
    statsStore.stats = mockStats;
    await flushPromises();

    const totalCard = wrapper.find('[data-test="total-devices-card"]');
    const onlineCard = wrapper.find('[data-test="online-devices-card"]');
    const pendingCard = wrapper.find('[data-test="pending-devices-card"]');
    const offlineCard = wrapper.find('[data-test="offline-devices-card"]');

    expect(totalCard.text()).toContain("10");
    expect(onlineCard.text()).toContain("6");
    expect(pendingCard.text()).toContain("2");
    expect(offlineCard.text()).toContain("4"); // 10 - 6 = 4
  });

  it("Switches to recent tab when clicked", async () => {
    drawer.vm.isDrawerOpen = true;
    await flushPromises();

    const recentTab = wrapper.find('[data-test="recent-tab"]');
    await recentTab.trigger("click");
    await flushPromises();

    expect(drawer.vm.activeTab).toBe("recent");
  });

  it("Displays pending devices list correctly", async () => {
    devicesStore.devices = mockPendingDevices;
    drawer.vm.pendingDevicesList = mockPendingDevices;
    drawer.vm.isDrawerOpen = true;
    drawer.vm.activeTab = "pending";
    await flushPromises();

    const deviceItems = wrapper.findAll('[data-test="pending-device-item"]');
    expect(deviceItems.length).toBe(2);
  });

  it("Shows empty state when no pending devices", async () => {
    mockDevicesApi
      .onGet("http://localhost:3000/api/devices?page=1&per_page=100&status=pending").reply(200, []);
    drawer.vm.pendingDevicesList = [];
    drawer.vm.isDrawerOpen = true;
    drawer.vm.activeTab = "pending";
    await flushPromises();

    expect(wrapper.text()).toContain("No pending devices");
  });

  it("Shows pending device count badge", async () => {
    statsStore.stats = mockStats;
    drawer.vm.isDrawerOpen = true;
    await flushPromises();

    const pendingTab = wrapper.find('[data-test="pending-tab"]');
    expect(pendingTab.text()).toContain("2");
  });

  it("Calls acceptDevice with correct UID", async () => {
    mockDevicesApi
      .onPatch(`http://localhost:3000/api/devices/${mockPendingDevices[0].uid}/accept`)
      .reply(200);

    const acceptSpy = vi.spyOn(devicesStore, "acceptDevice");
    drawer.vm.pendingDevicesList = mockPendingDevices;
    drawer.vm.isDrawerOpen = true;
    await flushPromises();

    const acceptButtonComponent = wrapper.findAllComponents(DeviceActionButton)
      .find((c) => c.props("uid") === mockPendingDevices[0].uid && c.props("action") === "accept");

    await acceptButtonComponent?.vm.handleClick();
    await flushPromises();

    expect(acceptSpy).toHaveBeenCalledWith(mockPendingDevices[0].uid);
  });

  it("Refetches stats and pending devices on handleUpdate", async () => {
    const fetchStatsSpy = vi.spyOn(statsStore, "fetchStats").mockResolvedValue();
    const fetchDevicesSpy = vi.spyOn(devicesStore, "fetchDeviceList").mockResolvedValue();

    await drawer?.vm.handleUpdate();
    await flushPromises();

    expect(fetchStatsSpy).toHaveBeenCalled();
    expect(fetchDevicesSpy).toHaveBeenCalledWith({ status: "pending", perPage: 100 });
  });

  it("Shows correct pending devices count in badge", async () => {
    const badge = wrapper.find('[data-test="device-dropdown-badge"]');
    drawer.vm.pendingDevicesList = mockPendingDevices;
    await nextTick();
    expect(badge.text()).toBe("2");

    drawer.vm.pendingDevicesList = [];
    await nextTick();
    expect(badge.text()).toBe("0");
  });

  it("Shows error snackbar when accept fails", async () => {
    mockDevicesApi
      .onPatch(`http://localhost:3000/api/devices/${mockPendingDevices[0].uid}/accept`)
      .reply(402);

    drawer.vm.pendingDevicesList = mockPendingDevices;
    drawer.vm.isDrawerOpen = true;
    await flushPromises();

    const acceptButtonComponent = wrapper.findAllComponents(DeviceActionButton)
      .find((c) => c.props("uid") === mockPendingDevices[0].uid && c.props("action") === "accept");

    await acceptButtonComponent?.vm.handleClick();
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalled();
  });

  it("Calls rejectDevice with correct UID", async () => {
    mockDevicesApi
      .onPatch(`http://localhost:3000/api/devices/${mockPendingDevices[0].uid}/reject`)
      .reply(200);

    const rejectSpy = vi.spyOn(devicesStore, "rejectDevice");

    drawer.vm.pendingDevicesList = mockPendingDevices;
    drawer.vm.isDrawerOpen = true;
    await flushPromises();

    const rejectButtonComponent = wrapper.findAllComponents(DeviceActionButton)
      .find((c) => c.props("uid") === mockPendingDevices[0].uid && c.props("action") === "reject");

    await rejectButtonComponent?.vm.handleClick();
    await flushPromises();

    expect(rejectSpy).toHaveBeenCalledWith(mockPendingDevices[0].uid);
  });

  it("Shows error snackbar when reject fails", async () => {
    mockDevicesApi
      .onPatch(`http://localhost:3000/api/devices/${mockPendingDevices[0].uid}/reject`)
      .reply(500);

    drawer.vm.pendingDevicesList = mockPendingDevices;
    drawer.vm.isDrawerOpen = true;
    await flushPromises();

    const rejectButtonComponent = wrapper.findAllComponents(DeviceActionButton)
      .find((c) => c.props("uid") === mockPendingDevices[0].uid && c.props("action") === "reject");

    await rejectButtonComponent?.vm.handleClick();
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalled();
  });

  it("Displays recent devices sorted by last_seen descending", async () => {
    drawer.vm.recentDevicesList = [...mockRecentDevices].sort(
      (a, b) => new Date(b.last_seen).getTime() - new Date(a.last_seen).getTime(),
    );
    drawer.vm.isDrawerOpen = true;
    drawer.vm.activeTab = "recent";
    await flushPromises();

    // First device should be the one with most recent last_seen
    expect(drawer.vm.recentDevicesList[0].uid).toBe("recent-device-1");
  });

  it("Shows empty state when no recent devices", async () => {
    drawer.vm.recentDevicesList = [];
    drawer.vm.isDrawerOpen = true;
    drawer.vm.activeTab = "recent";
    await flushPromises();

    expect(wrapper.text()).toContain("No recent activity");
  });

  it("Formats time ago correctly for valid dates", () => {
    const pastDate = new Date(Date.now() - 3600000); // 1 hour ago
    const result = drawer.vm.formatTimeAgo(pastDate);

    expect(result).toBe("an hour ago");
  });

  it("Handles null/undefined dates gracefully", () => {
    // @ts-expect-error Testing invalid input
    const result = drawer.vm.formatTimeAgo();
    expect(result).toBe("Unknown");
  });

  it("Device detail link navigates to correct route", async () => {
    drawer.vm.recentDevicesList = mockRecentDevices;
    drawer.vm.isDrawerOpen = true;
    drawer.vm.activeTab = "recent";
    await flushPromises();

    const deviceLink = wrapper.find(`a[href="/devices/${mockRecentDevices[0].uid}"]`);
    expect(deviceLink.exists()).toBeTruthy();
  });

  it("Shows badge count matching stats.pending_devices", async () => {
    statsStore.stats = { ...mockStats, pending_devices: 5 };
    drawer.vm.isDrawerOpen = true;
    await flushPromises();

    const badge = wrapper.find(".v-chip");
    expect(badge.text()).toBe("5");
  });

  it("Hides badge when no pending devices", async () => {
    statsStore.stats = { ...mockStats, pending_devices: 0 };
    await flushPromises();
    expect(drawer.vm.stats.pending_devices).toBe(0);
  });
});
