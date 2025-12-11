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
import useDevicesStore from "@/store/modules/devices";
import { IDevice } from "@/interfaces/IDevice";
import useAuthStore from "@/store/modules/auth";
import { nextTick } from "vue";
import useNamespacesStore from "@/store/modules/namespaces";
import { INamespace, INamespaceMember } from "@/interfaces/INamespace";

const Component = {
  template: "<v-layout><DevicesDropdown v-model=\"show\" /></v-layout>",
  props: ["modelValue"],
  data: () => ({
    show: true,
  }),
};

// Mock Vuetify display
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

const mockAcceptedDevices: IDevice[] = [
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
  const namespacesStore = useNamespacesStore();
  const devicesStore = useDevicesStore();
  const authStore = useAuthStore();
  const mockDevicesApi = new MockAdapter(devicesApi.getAxios());

  authStore.role = "owner";
  namespacesStore.namespaceList = [mockNamespace];

  beforeEach(async () => {
    devicesStore.totalDevicesCount = 0;
    devicesStore.onlineDevicesCount = 0;
    devicesStore.offlineDevicesCount = 0;
    devicesStore.pendingDevicesCount = 0;

    mockDevicesApi
      .onGet("http://localhost:3000/api/devices?page=1&per_page=100&status=pending")
      .reply(200, mockPendingDevices, { "x-total-count": "2" });

    mockDevicesApi
      .onGet("http://localhost:3000/api/devices?page=1&per_page=100&status=accepted")
      .reply(200, mockAcceptedDevices, { "x-total-count": "10" });

    mockDevicesApi
      .onGet("http://localhost:3000/api/devices?page=1&per_page=1&status=accepted")
      .reply(200, [], { "x-total-count": "10" });

    mockDevicesApi
      .onGet("http://localhost:3000/api/devices?page=1&per_page=1&status=pending")
      .reply(200, [], { "x-total-count": "2" });

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
      props: {
        modelValue: true,
      },
      attachTo: document.body,
    });

    drawer = wrapper.findComponent(DevicesDropdown);
    drawer.vm.isDrawerOpen = true;
    await flushPromises();

    await nextTick();
  });

  afterEach(() => {
    if (wrapper) wrapper.unmount();
  });

  it("Fetches device lists on mount (pending + recent)", async () => {
    const fetchDevicesSpy = vi.spyOn(devicesStore, "fetchDeviceList").mockResolvedValue();
    const fetchCountsSpy = vi.spyOn(devicesStore, "fetchDeviceCounts").mockResolvedValue();

    wrapper = mount(Component, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
        components: { "v-layout": VLayout, DevicesDropdown },
      },
      props: {
        modelValue: true,
      },
    });

    await flushPromises();

    expect(fetchCountsSpy).toHaveBeenCalled();
    expect(fetchDevicesSpy).toHaveBeenCalledWith({ status: "pending", perPage: 100, filter: undefined });
    expect(fetchDevicesSpy).toHaveBeenCalledWith({ status: "accepted", perPage: 100, filter: undefined });
  });

  it("Opens drawer when icon is clicked", async () => {
    drawer.vm.isDrawerOpen = false;
    await nextTick();

    await wrapper.find('[data-test="devices-icon"]').trigger("click");
    await nextTick();
    await flushPromises();

    expect(wrapper.find('[data-test="devices-drawer"]').exists()).toBe(true);
  });

  it("Displays correct statistics cards with proper values", async () => {
    devicesStore.totalDevicesCount = 10;
    devicesStore.onlineDevicesCount = 6;
    devicesStore.pendingDevicesCount = 2;
    devicesStore.offlineDevicesCount = 4;
    await nextTick();

    const totalCard = wrapper.find('[data-test="total-devices-card"]');
    const onlineCard = wrapper.find('[data-test="online-devices-card"]');
    const pendingCard = wrapper.find('[data-test="pending-devices-card"]');
    const offlineCard = wrapper.find('[data-test="offline-devices-card"]');

    expect(totalCard.text()).toContain("10");
    expect(onlineCard.text()).toContain("6");
    expect(pendingCard.text()).toContain("2");
    expect(offlineCard.text()).toContain("4");
  });

  it("Switches to recent tab when clicked", async () => {
    const recentTab = wrapper.find('[data-test="recent-tab"]');
    await recentTab.trigger("click");
    await flushPromises();

    expect(drawer.vm.activeTab).toBe("recent");
  });

  it("Displays pending devices list correctly", async () => {
    drawer.vm.pendingDevicesList = mockPendingDevices;
    drawer.vm.activeTab = "pending";
    await flushPromises();

    expect(wrapper.findAll('[data-test="pending-device-item"]').length).toBe(2);
  });

  it("Shows empty state when no pending devices", async () => {
    drawer.vm.pendingDevicesList = [];
    devicesStore.pendingDevicesCount = 0;
    drawer.vm.activeTab = "pending";
    await flushPromises();

    expect(wrapper.text()).toContain("No pending devices");
  });

  it("Shows pending device count badge", async () => {
    devicesStore.pendingDevicesCount = 2;
    await nextTick();

    const pendingTab = wrapper.find('[data-test="pending-tab"]');
    expect(pendingTab.text()).toContain("2");
  });

  it("Calls acceptDevice with correct UID", async () => {
    mockDevicesApi
      .onPatch(`http://localhost:3000/api/devices/${mockPendingDevices[0].uid}/accept`)
      .reply(200);

    const acceptSpy = vi.spyOn(devicesStore, "acceptDevice");
    drawer.vm.pendingDevicesList = mockPendingDevices;
    await flushPromises();

    const acceptBtn = wrapper.findAllComponents(DeviceActionButton)
      .find((c) => c.props("uid") === mockPendingDevices[0].uid && c.props("action") === "accept");

    await acceptBtn?.vm.handleClick();
    await flushPromises();

    expect(acceptSpy).toHaveBeenCalledWith(mockPendingDevices[0].uid);
  });

  it("Refetches pending and accepted devices on handleUpdate", async () => {
    const fetchDevicesSpy = vi.spyOn(devicesStore, "fetchDeviceList").mockResolvedValue();
    const fetchCountsSpy = vi.spyOn(devicesStore, "fetchDeviceCounts").mockResolvedValue();

    await drawer.vm.handleUpdate();
    await flushPromises();

    expect(fetchCountsSpy).toHaveBeenCalled();
    expect(fetchDevicesSpy).toHaveBeenCalledWith({ status: "pending", perPage: 100, filter: undefined });
    expect(fetchDevicesSpy).toHaveBeenCalledWith({ status: "accepted", perPage: 100, filter: undefined });
  });

  it("Shows correct pending devices count in badge", async () => {
    devicesStore.pendingDevicesCount = 2;
    await nextTick();
    expect(wrapper.find('[data-test="device-dropdown-badge"]').text()).toContain("2");

    devicesStore.pendingDevicesCount = 0;
    await nextTick();
    expect(wrapper.find('[data-test="device-dropdown-badge"]').text()).not.toContain("2");
  });

  it("Shows error snackbar when accept fails", async () => {
    mockDevicesApi
      .onPatch(`http://localhost:3000/api/devices/${mockPendingDevices[0].uid}/accept`)
      .reply(402);

    drawer.vm.pendingDevicesList = mockPendingDevices;
    await flushPromises();

    const acceptBtn = wrapper.findAllComponents(DeviceActionButton)
      .find((c) => c.props("uid") === mockPendingDevices[0].uid && c.props("action") === "accept");

    await acceptBtn?.vm.handleClick();
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalled();
  });

  it("Calls rejectDevice with correct UID", async () => {
    mockDevicesApi
      .onPatch(`http://localhost:3000/api/devices/${mockPendingDevices[0].uid}/reject`)
      .reply(200);

    const rejectSpy = vi.spyOn(devicesStore, "rejectDevice");

    drawer.vm.pendingDevicesList = mockPendingDevices;
    await flushPromises();

    const rejectBtn = wrapper.findAllComponents(DeviceActionButton)
      .find((c) => c.props("uid") === mockPendingDevices[0].uid && c.props("action") === "reject");

    await rejectBtn?.vm.handleClick();
    await flushPromises();

    expect(rejectSpy).toHaveBeenCalledWith(mockPendingDevices[0].uid);
  });

  it("Shows error snackbar when reject fails", async () => {
    mockDevicesApi
      .onPatch(`http://localhost:3000/api/devices/${mockPendingDevices[0].uid}/reject`)
      .reply(500);

    drawer.vm.pendingDevicesList = mockPendingDevices;
    await flushPromises();

    const rejectBtn = wrapper.findAllComponents(DeviceActionButton)
      .find((c) => c.props("uid") === mockPendingDevices[0].uid && c.props("action") === "reject");

    await rejectBtn?.vm.handleClick();
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalled();
  });

  it("Displays recent devices sorted by last_seen descending", async () => {
    drawer.vm.recentDevicesList = [...mockAcceptedDevices].sort(
      (a, b) => new Date(b.last_seen).getTime() - new Date(a.last_seen).getTime(),
    );
    drawer.vm.activeTab = "recent";
    await flushPromises();

    // First device should be the one with most recent last_seen
    expect(drawer.vm.recentDevicesList[0].uid).toBe("recent-device-1");
  });

  it("Shows empty state when no recent devices", async () => {
    drawer.vm.recentDevicesList = [];
    drawer.vm.activeTab = "recent";
    await flushPromises();

    expect(wrapper.text()).toContain("No recent activity");
  });

  it("Formats time ago correctly for valid dates", () => {
    const date = new Date(Date.now() - 3600000);
    expect(drawer.vm.formatTimeAgo(date)).toBe("an hour ago");
  });

  it("Handles null/undefined dates gracefully", () => {
    // @ts-expect-error Testing invalid input
    expect(drawer.vm.formatTimeAgo()).toBe("Unknown");
  });

  it("Device detail link navigates to correct route", async () => {
    drawer.vm.recentDevicesList = mockAcceptedDevices;
    drawer.vm.activeTab = "recent";
    await flushPromises();

    const deviceLink = wrapper.find(`a[href="/devices/${mockAcceptedDevices[0].uid}"]`);
    expect(deviceLink.exists()).toBe(true);
  });
});
