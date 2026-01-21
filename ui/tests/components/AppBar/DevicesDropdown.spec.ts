import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { flushPromises, VueWrapper } from "@vue/test-utils";
import { createCleanRouter } from "@tests/utils/router";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import DevicesDropdown from "@/components/AppBar/DevicesDropdown.vue";
import { Router } from "vue-router";
import useDevicesStore from "@/store/modules/devices";
import type { IDevice } from "@/interfaces/IDevice";
import { mockDevice } from "@tests/mocks";
import { VBadge, VLayout } from "vuetify/components";
import { createAxiosError } from "@tests/utils/axiosError";

vi.mock("@/store/api/devices");

const Component = {
  template: "<v-layout><DevicesDropdown v-model=\"show\" /></v-layout>",
  data: () => ({
    show: true,
  }),
};

const mockPendingDevices: IDevice[] = [
  {
    ...mockDevice[0],
    uid: "pending-1",
    name: "pending-device-1",
    status: "pending",
    online: false,
    last_seen: "2026-01-20T10:00:00Z",
    status_updated_at: "2026-01-20T09:00:00Z",
    identity: { mac: "00:11:22:33:44:55" },
    remote_addr: "192.168.1.100",
  },
  {
    ...mockDevice[1],
    uid: "pending-2",
    name: "pending-device-2",
    status: "pending",
    online: false,
    last_seen: "2026-01-20T11:00:00Z",
    status_updated_at: "2026-01-20T10:30:00Z",
    identity: { mac: "00:11:22:33:44:66" },
    remote_addr: "192.168.1.101",
  },
];

const mockRecentDevices: IDevice[] = [
  {
    ...mockDevice[0],
    uid: "recent-1",
    name: "recent-device-1",
    status: "accepted",
    online: true,
    last_seen: "2026-01-22T15:00:00Z",
    identity: { mac: "AA:BB:CC:DD:EE:FF" },
  },
  {
    ...mockDevice[1],
    uid: "recent-2",
    name: "recent-device-2",
    status: "accepted",
    online: false,
    last_seen: "2026-01-21T12:00:00Z",
    identity: { mac: "AA:BB:CC:DD:EE:AA" },
  },
];

describe("DevicesDropdown", () => {
  let wrapper: VueWrapper;
  let drawer: VueWrapper<InstanceType<typeof DevicesDropdown>>;
  let router: Router;
  let devicesStore: ReturnType<typeof useDevicesStore>;

  const mountWrapper = (mockDevices = mockPendingDevices) => {
    router = createCleanRouter();

    wrapper = mountComponent(Component, {
      global: {
        plugins: [router],
        stubs: { teleport: true },
        components: { DevicesDropdown, "v-layout": VLayout },
      },
      piniaOptions: {
        initialState: {
          devices: {
            devices: mockDevices,
            totalDevicesCount: 12,
            onlineDevicesCount: 7,
            offlineDevicesCount: 5,
            pendingDevicesCount: mockDevices.some((device) => device.status === "pending") ? 2 : 0,
          },
        },
      },
      attachTo: document.body,
    });
    drawer = wrapper.findComponent(DevicesDropdown);

    devicesStore = useDevicesStore();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders the device icon button", () => {
      const icon = wrapper.find('[data-test="devices-icon"]');
      expect(icon.exists()).toBe(true);
    });

    it("displays badge with pending devices count when there are pending devices", () => {
      const badge = wrapper.find('[data-test="device-dropdown-badge"]');
      expect(badge.exists()).toBe(true);
      expect(badge.text()).toContain("2");
    });

    it("renders the navigation drawer when open", () => {
      expect(drawer.exists()).toBe(true);
    });

    it("displays device management title", () => {
      expect(drawer.text()).toContain("Device Management");
    });
  });

  describe("statistics cards", () => {
    beforeEach(() => mountWrapper());

    it("displays total devices count", () => {
      const totalCard = drawer.find('[data-test="total-devices-card"]');
      expect(totalCard.text()).toContain("12");
      expect(totalCard.text()).toContain("Total");
    });

    it("displays online devices count", () => {
      const onlineCard = drawer.find('[data-test="online-devices-card"]');
      expect(onlineCard.text()).toContain("7");
      expect(onlineCard.text()).toContain("Online");
    });

    it("displays pending devices count", () => {
      const pendingCard = drawer.find('[data-test="pending-devices-card"]');
      expect(pendingCard.text()).toContain("2");
      expect(pendingCard.text()).toContain("Pending");
    });

    it("displays offline devices count", () => {
      const offlineCard = drawer.find('[data-test="offline-devices-card"]');
      expect(offlineCard.text()).toContain("5");
      expect(offlineCard.text()).toContain("Offline");
    });

    it("updates statistics when store values change", async () => {
      devicesStore.totalDevicesCount = 20;
      devicesStore.onlineDevicesCount = 10;
      devicesStore.pendingDevicesCount = 5;
      devicesStore.offlineDevicesCount = 5;
      await flushPromises();

      expect(drawer.find('[data-test="total-devices-card"]').text()).toContain("20");
      expect(drawer.find('[data-test="online-devices-card"]').text()).toContain("10");
      expect(drawer.find('[data-test="pending-devices-card"]').text()).toContain("5");
      expect(drawer.find('[data-test="offline-devices-card"]').text()).toContain("5");
    });
  });

  describe("tabs", () => {
    beforeEach(() => mountWrapper());

    it("displays pending tab by default", () => {
      const pendingTab = drawer.find('[data-test="pending-tab"]');
      expect(pendingTab.exists()).toBe(true);
      expect(pendingTab.text()).toContain("Pending Approval");
    });

    it("displays recent tab", () => {
      const recentTab = drawer.find('[data-test="recent-tab"]');
      expect(recentTab.exists()).toBe(true);
      expect(recentTab.text()).toContain("Recent Activity");
    });

    it("shows pending count badge on pending tab when devices exist", () => {
      const pendingTab = drawer.find('[data-test="pending-tab"]');
      expect(pendingTab.text()).toContain("2");
    });

    it("switches to recent tab when clicked", async () => {
      const recentTab = drawer.find('[data-test="recent-tab"]');
      await recentTab.trigger("click");
      await flushPromises();

      expect(drawer.vm.activeTab).toBe("recent");
    });

    it("switches back to pending tab when clicked", async () => {
      await drawer.find('[data-test="recent-tab"]').trigger("click");
      await flushPromises();

      const pendingTab = drawer.find('[data-test="pending-tab"]');
      await pendingTab.trigger("click");
      await flushPromises();

      expect(drawer.vm.activeTab).toBe("pending");
    });
  });

  describe("pending devices list", () => {
    beforeEach(() => {
      mountWrapper();
      drawer.vm.activeTab = "pending";
    });

    it("displays pending devices list", async () => {
      await flushPromises();
      const items = drawer.findAll('[data-test="pending-device-item"]');
      expect(items).toHaveLength(2);
    });

    it("shows device name for each pending device", async () => {
      await flushPromises();
      expect(drawer.text()).toContain("pending-device-1");
      expect(drawer.text()).toContain("pending-device-2");
    });

    it("shows device MAC address or UID", async () => {
      await flushPromises();
      expect(drawer.text()).toContain("00:11:22:33:44:55");
      expect(drawer.text()).toContain("00:11:22:33:44:66");
    });

    it("shows device remote address", async () => {
      await flushPromises();
      expect(drawer.text()).toContain("192.168.1.100");
      expect(drawer.text()).toContain("192.168.1.101");
    });

    it("displays accept button for each pending device", async () => {
      await flushPromises();
      const acceptBtn = drawer.find('[data-test="accept-pending-1"]');
      expect(acceptBtn.exists()).toBe(true);
    });

    it("displays reject button for each pending device", async () => {
      await flushPromises();
      const rejectBtn = drawer.find('[data-test="reject-pending-1"]');
      expect(rejectBtn.exists()).toBe(true);
    });

    it("shows time ago for status update", async () => {
      await flushPromises();
      const timeAgo = drawer.vm.formatTimeAgo(mockPendingDevices[0].status_updated_at);
      expect(timeAgo).toBeTruthy();
      expect(timeAgo).not.toBe("Unknown");
    });
  });

  describe("pending devices - empty state", () => {
    beforeEach(() => {
      mountWrapper([]);
      drawer.vm.activeTab = "pending";
    });

    it("shows empty state when no pending devices", async () => {
      await flushPromises();
      expect(drawer.text()).toContain("No pending devices");
      expect(drawer.text()).toContain("All devices have been approved");
    });

    it("displays success icon in empty state", async () => {
      await flushPromises();
      const icon = drawer.find(".mdi-check-circle");
      expect(icon.exists()).toBe(true);
    });

    it("does not show pending device items", async () => {
      await flushPromises();
      const items = drawer.findAll('[data-test="pending-device-item"]');
      expect(items).toHaveLength(0);
    });
  });

  describe("recent devices list", () => {
    beforeEach(() => {
      mountWrapper(mockRecentDevices);
      drawer.vm.activeTab = "recent";
    });

    it("displays recent devices sorted by last seen", async () => {
      await flushPromises();
      const sortedDevices = [...mockRecentDevices].sort(
        (a, b) => new Date(b.last_seen).getTime() - new Date(a.last_seen).getTime(),
      );
      expect(drawer.vm.recentDevicesList[0].uid).toBe(sortedDevices[0].uid);
    });

    it("shows online badge for online devices", async () => {
      await flushPromises();
      const badges = drawer.findAll(".v-badge");
      expect(badges.length).toBeGreaterThan(0);
    });

    it("shows 'Active now' for online devices", async () => {
      await flushPromises();
      expect(drawer.text()).toContain("Active now");
    });

    it("shows time ago for offline devices", async () => {
      await flushPromises();
      const offlineDevice = mockRecentDevices.find((d) => !d.online);
      if (offlineDevice) {
        const timeAgo = drawer.vm.formatTimeAgo(offlineDevice.last_seen);
        expect(timeAgo).toBeTruthy();
      }
    });

    it("creates navigation links to device details", async () => {
      await flushPromises();
      const link = drawer.find(`a[href="/devices/${mockRecentDevices[0].uid}"]`);
      expect(link.exists()).toBe(true);
    });
  });

  describe("recent devices - empty state", () => {
    beforeEach(() => {
      mountWrapper([]);
      drawer.vm.activeTab = "recent";
    });

    it("shows empty state when no recent devices", async () => {
      await flushPromises();
      expect(drawer.text()).toContain("No recent activity");
    });

    it("displays history icon in empty state", async () => {
      await flushPromises();
      const icon = drawer.find(".mdi-history");
      expect(icon.exists()).toBe(true);
    });
  });

  describe("drawer interactions", () => {
    it("opens drawer when icon is clicked", async () => {
      mountWrapper();
      const icon = wrapper.find('[data-test="devices-icon"]');

      await icon.trigger("click"); // Close first
      await flushPromises();
      await icon.trigger("click");
      await flushPromises();

      expect(drawer.emitted("update:modelValue")).toBeTruthy();
      expect(drawer.emitted("update:modelValue")?.[1]).toEqual([true]);
    });

    it("closes drawer when already open", async () => {
      mountWrapper();
      const icon = wrapper.find('[data-test="devices-icon"]');

      await icon.trigger("click");
      await flushPromises();

      expect(drawer.emitted("update:modelValue")).toBeTruthy();
      expect(drawer.emitted("update:modelValue")?.[0]).toEqual([false]);
    });

    it("toggles drawer state multiple times", async () => {
      mountWrapper();
      const icon = wrapper.find('[data-test="devices-icon"]');

      await icon.trigger("click");
      await icon.trigger("click");
      await icon.trigger("click");
      await flushPromises();

      expect(drawer.emitted("update:modelValue")).toHaveLength(3);
    });
  });

  describe("badge visibility", () => {
    it("shows badge when there are pending devices", () => {
      mountWrapper();
      const badge = wrapper.findComponent('[data-test="device-dropdown-badge"]') as VueWrapper<VBadge>;
      expect(badge.props("modelValue")).toBe(true);
      expect(badge.text()).toContain("2");
    });

    it("hides badge when no pending devices", () => {
      mountWrapper([]);
      const badge = wrapper.findComponent('[data-test="device-dropdown-badge"]') as VueWrapper<VBadge>;
      expect(badge.props("modelValue")).toBe(false);
    });

    it("updates badge count when pending devices change", async () => {
      mountWrapper();

      devicesStore.pendingDevicesCount = 5;
      await flushPromises();

      const badge = wrapper.find('[data-test="device-dropdown-badge"]');
      expect(badge.text()).toContain("5");
    });
  });

  describe("data fetching", () => {
    beforeEach(() => mountWrapper());

    it("fetches device list on mount", () => {
      expect(devicesStore.fetchDeviceList).toHaveBeenCalledWith(
        expect.objectContaining({ status: "pending", perPage: 100 }),
      );
      expect(devicesStore.fetchDeviceList).toHaveBeenCalledWith(
        expect.objectContaining({ status: "accepted", perPage: 100 }),
      );
    });

    it("fetches device counts on mount", () => {
      expect(devicesStore.fetchDeviceCounts).toHaveBeenCalled();
    });

    it("refetches data when handleUpdate is called", async () => {
      vi.clearAllMocks();

      await drawer.vm.handleUpdate();
      await flushPromises();

      expect(devicesStore.fetchDeviceCounts).toHaveBeenCalled();
      expect(devicesStore.fetchDeviceList).toHaveBeenCalledWith(
        expect.objectContaining({ status: "pending" }),
      );
      expect(devicesStore.fetchDeviceList).toHaveBeenCalledWith(
        expect.objectContaining({ status: "accepted" }),
      );
    });
  });

  describe("error handling", () => {
    it("shows error snackbar when fetching pending devices fails", async () => {
      mountWrapper();
      vi.mocked(devicesStore.fetchDeviceList).mockRejectedValueOnce(createAxiosError(500, "Internal Server Error"));
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load pending devices");
    });

    it("shows error snackbar when fetching recent devices fails", async () => {
      mountWrapper();
      vi.mocked(devicesStore.fetchDeviceList).mockResolvedValueOnce().mockRejectedValueOnce(createAxiosError(500, "Internal Server Error"));
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load recent devices");
    });

    it("shows error snackbar when handleUpdate fails", async () => {
      mountWrapper();
      vi.mocked(devicesStore.fetchDeviceCounts).mockRejectedValueOnce(new Error("Network error"));

      await drawer.vm.handleUpdate();
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to update device data");
    });
  });

  describe("utility functions", () => {
    beforeEach(() => mountWrapper());

    it("formats time ago correctly for valid dates", () => {
      const oneHourAgo = new Date(Date.now() - 3600000);
      const result = drawer.vm.formatTimeAgo(oneHourAgo);
      expect(result).toBe("an hour ago");
    });

    it("formats time ago for recent dates", () => {
      const fiveMinutesAgo = new Date(Date.now() - 300000);
      const result = drawer.vm.formatTimeAgo(fiveMinutesAgo);
      expect(result).toContain("minutes ago");
    });

    it("returns Unknown for null date", () => {
      // @ts-expect-error Testing null input
      const result = drawer.vm.formatTimeAgo(null);
      expect(result).toBe("Unknown");
    });

    it("returns Unknown for undefined date", () => {
      // @ts-expect-error Testing undefined input
      const result = drawer.vm.formatTimeAgo(undefined);
      expect(result).toBe("Unknown");
    });

    it("handles date strings correctly", () => {
      const dateString = "2026-01-20T10:00:00Z";
      const result = drawer.vm.formatTimeAgo(dateString);
      expect(result).toBeTruthy();
      expect(result).not.toBe("Unknown");
    });
  });

  describe("view all devices button", () => {
    beforeEach(() => mountWrapper());

    it("displays view all devices button", () => {
      const button = drawer.find('[data-test="view-all-devices-btn"]');
      expect(button.exists()).toBe(true);
      expect(button.text()).toContain("View all devices");
    });

    it("links to devices page", () => {
      const button = drawer.find('[data-test="view-all-devices-btn"]');
      expect(button.attributes("href")).toBe("/devices");
    });
  });

  describe("responsive behavior", () => {
    beforeEach(() => mountWrapper());

    it("renders drawer with responsive width", () => {
      const drawer = wrapper.find('[data-test="devices-drawer"]');
      expect(drawer.exists()).toBe(true);
    });

    it("displays tab toggle for mobile and desktop", () => {
      const tabToggle = wrapper.find('[data-test="tab-toggle"]');
      expect(tabToggle.exists()).toBe(true);
    });
  });
});
