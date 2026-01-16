import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import useDevicesStore from "@admin/store/modules/devices";
import DeviceList from "@admin/components/Device/DeviceList.vue";
import { mockDevices } from "../../mocks";
import { Router } from "vue-router";

describe("DeviceList", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceList>>;
  let router: Router;
  let devicesStore: ReturnType<typeof useDevicesStore>;

  const mountWrapper = (mockDeviceCount?: number) => {
    router = createCleanAdminRouter();

    wrapper = mountComponent(DeviceList, {
      global: { plugins: [router] },
      piniaOptions: {
        initialState: {
          adminDevices: {
            devices: mockDevices,
            deviceCount: mockDeviceCount ?? mockDevices.length,
          },
        },
      },
    });

    devicesStore = useDevicesStore();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders the data table", () => {
      expect(wrapper.find('[data-test="devices-list"]').exists()).toBe(true);
    });

    it("displays device names", () => {
      expect(wrapper.text()).toContain(mockDevices[0].name);
      expect(wrapper.text()).toContain(mockDevices[1].name);
    });

    it("displays device info", () => {
      expect(wrapper.text()).toContain(mockDevices[0].info.pretty_name);
      expect(wrapper.text()).toContain(mockDevices[1].info.pretty_name);
    });

    it("displays device namespaces", () => {
      const namespaceLinks = wrapper.findAll('[data-test="namespace-link"]');
      expect(namespaceLinks).toHaveLength(mockDevices.length);
      expect(namespaceLinks[0].text()).toBe(mockDevices[0].namespace);
    });

    it("displays online status icons", () => {
      const onlineIcons = wrapper.findAll('[data-test="success-icon"]');
      const offlineIcons = wrapper.findAll('[data-test="error-icon"]');

      const onlineCount = mockDevices.filter((d) => d.online).length;
      const offlineCount = mockDevices.filter((d) => !d.online).length;

      expect(onlineIcons).toHaveLength(onlineCount);
      expect(offlineIcons).toHaveLength(offlineCount);
    });

    it("displays device status", () => {
      expect(wrapper.text()).toContain(mockDevices[0].status);
    });

    it("displays info buttons for each device", () => {
      const infoButtons = wrapper.findAll('[data-test="info-button"]');
      expect(infoButtons).toHaveLength(mockDevices.length);
    });
  });

  describe("fetching devices", () => {
    it("fetches devices on mount", () => {
      mountWrapper();

      expect(devicesStore.fetchDeviceList).toHaveBeenCalledWith(
        expect.objectContaining({
          perPage: 10,
          page: 1,
        }),
      );
    });

    it("refetches devices when page changes", async () => {
      mountWrapper(11); // Mock total count to 11 to enable pagination

      // Click next page button
      const nextPageBtn = wrapper.find('[data-test="pager-next"]');
      await nextPageBtn.trigger("click");
      await flushPromises();

      expect(devicesStore.fetchDeviceList).toHaveBeenCalledWith(
        expect.objectContaining({
          page: 2,
        }),
      );
    });

    it("refetches devices when items per page changes", async () => {
      mountWrapper(20);

      // Change items per page via combobox
      const ippCombo = wrapper.find('[data-test="ipp-combo"] input');
      await ippCombo.setValue(20);
      await flushPromises();

      expect(devicesStore.fetchDeviceList).toHaveBeenCalledWith(
        expect.objectContaining({
          perPage: 20,
        }),
      );
    });
  });

  describe("navigating to device details", () => {
    it("navigates when clicking info button", async () => {
      mountWrapper();

      const pushSpy = vi.spyOn(router, "push");
      const infoButton = wrapper.findAll('[data-test="info-button"]')[0];

      await infoButton.trigger("click");

      expect(pushSpy).toHaveBeenCalledWith({
        name: "deviceDetails",
        params: { id: mockDevices[0].uid },
      });
    });
  });

  describe("navigating to namespace details", () => {
    it("navigates when clicking namespace link", async () => {
      mountWrapper();

      const pushSpy = vi.spyOn(router, "push");
      const namespaceLink = wrapper.findAll('[data-test="namespace-link"]')[0];

      await namespaceLink.trigger("click");

      expect(pushSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "namespaceDetails",
          params: { id: mockDevices[0].tenant_id },
        }),
      );
    });
  });

  describe("error handling", () => {
    it("shows error snackbar when fetching devices fails", async () => {
      mountWrapper(11);
      vi.mocked(devicesStore.fetchDeviceList).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );

      // Trigger refetch by changing page
      const nextPageBtn = wrapper.find('[data-test="pager-next"]');
      await nextPageBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to fetch devices.");
    });
  });
});
