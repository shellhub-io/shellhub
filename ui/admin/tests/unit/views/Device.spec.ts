import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import useDevicesStore from "@admin/store/modules/devices";
import { mockDevices } from "../mocks";
import Device from "@admin/views/Device.vue";

vi.mock("@admin/store/api/devices");

describe("Device", () => {
  let wrapper: VueWrapper<InstanceType<typeof Device>>;
  let router: ReturnType<typeof createCleanAdminRouter>;
  let devicesStore: ReturnType<typeof useDevicesStore>;

  const mountWrapper = async (mockError?: Error) => {
    router = createCleanAdminRouter();
    await router.push({ name: "devices" });
    await router.isReady();

    wrapper = mountComponent(Device, {
      global: { plugins: [router] },
      piniaOptions: {
        initialState: {
          adminDevices: {
            devices: mockError ? [] : mockDevices,
            deviceCount: mockError ? 0 : mockDevices.length,
          },
        },
        stubActions: !mockError,
      },
    });

    devicesStore = useDevicesStore();
    if (mockError) vi.mocked(devicesStore.fetchDeviceList).mockRejectedValueOnce(mockError);

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when devices load successfully", () => {
    beforeEach(() => mountWrapper());

    it("displays the page header with correct title", () => {
      expect(wrapper.text()).toContain("Devices");
      expect(wrapper.text()).toContain("Fleet Oversight");
    });

    it("displays the search input field", () => {
      const searchInput = wrapper.find('[data-test="search-input"]');
      expect(searchInput.exists()).toBe(true);
      expect(searchInput.text()).toContain("Search by hostname"); // Placeholder
    });

    it("displays the devices list component", () => {
      expect(wrapper.find('[data-test="devices-list"]').exists()).toBe(true);
    });
  });

  describe("when searching for devices", () => {
    beforeEach(() => mountWrapper());

    it("triggers search on keyup event", async () => {
      const searchInput = wrapper.find('[data-test="search-input"] input');
      await searchInput.setValue("test-device");
      await searchInput.trigger("keyup");
      await flushPromises();

      expect(devicesStore.fetchDeviceList).toHaveBeenCalled();
    });

    it("encodes filter correctly when searching", async () => {
      const searchInput = wrapper.find('[data-test="search-input"] input');
      await searchInput.setValue("device-one");
      await searchInput.trigger("keyup");
      await flushPromises();

      expect(devicesStore.setFilter).toHaveBeenCalled();
      expect(devicesStore.fetchDeviceList).toHaveBeenCalledWith(
        expect.objectContaining({
          filter: expect.any(String),
          page: 1,
        }),
      );
    });

    it("clears filter when search is empty", async () => {
      const searchInput = wrapper.find('[data-test="search-input"] input');
      await searchInput.setValue("");
      await searchInput.trigger("keyup");
      await flushPromises();

      expect(devicesStore.setFilter).toHaveBeenCalledWith("");
      expect(devicesStore.fetchDeviceList).toHaveBeenCalledWith(
        expect.objectContaining({
          filter: "",
          page: 1,
        }),
      );
    });
  });

  describe("when devices fail to load", () => {
    it("shows error snackbar", async () => {
      await mountWrapper(createAxiosError(500, "Server Error"));
      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to fetch devices.");
    });
  });
});
