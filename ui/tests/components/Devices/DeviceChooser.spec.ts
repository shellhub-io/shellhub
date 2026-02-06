import { describe, expect, it, afterEach, vi, beforeEach } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import DeviceChooser from "@/components/Devices/DeviceChooser.vue";
import useDevicesStore from "@/store/modules/devices";
import { mockDevice, mockDeviceForSession } from "@tests/mocks/device";
import { createCleanRouter } from "@tests/utils/router";

vi.mock("@/utils/permission", () => ({
  default: vi.fn(() => true),
}));

const mockDevices = [mockDevice, mockDeviceForSession];

describe("DeviceChooser", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceChooser>>;
  let dialog: DOMWrapper<Element>;
  let devicesStore: ReturnType<typeof useDevicesStore>;

  const mountWrapper = (hasSuggestedDevices = false) => {
    wrapper = mountComponent(DeviceChooser, {
      global: { plugins: [createCleanRouter()] },
      piniaOptions: {
        initialState: {
          devices: {
            showDeviceChooser: true,
            devices: mockDevices,
            suggestedDevices: hasSuggestedDevices ? mockDevices : [],
            selectedDevices: [],
          },
        },
      },
    });

    devicesStore = useDevicesStore();
    dialog = new DOMWrapper(document.body);
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("dialog rendering", () => {
    it("renders FormDialog when showDeviceChooser is true", () => {
      mountWrapper();

      expect(wrapper.findComponent({ name: "FormDialog" }).exists()).toBe(true);
    });

    it("displays correct dialog title", () => {
      mountWrapper();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("title")).toBe("Update account or select three devices");
    });

    it("shows explanation text with billing links", () => {
      mountWrapper();

      expect(dialog.text()).toContain("free version is limited to 3 devices");
      expect(dialog.text()).toContain("premium plan");
    });

    it("renders tabs for Suggested and All devices", () => {
      mountWrapper(true);

      expect(dialog.find('[data-test="Suggested-tab"]').exists()).toBe(true);
      expect(dialog.find('[data-test="All-tab"]').exists()).toBe(true);
    });

    it("disables Suggested tab when no suggested devices", () => {
      mountWrapper();

      const suggestedTab = dialog.find('[data-test="Suggested-tab"]');
      expect(suggestedTab.attributes("disabled")).toBeDefined();
    });

    it("renders DeviceListChooser component", () => {
      mountWrapper();

      expect(wrapper.findComponent({ name: "DeviceListChooser" }).exists()).toBe(true);
    });
  });

  describe("search functionality", () => {
    beforeEach(() => mountWrapper());

    it("shows search field when on All tab", async () => {
      await dialog.find('[data-test="All-tab"]').trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="search-text"]').exists()).toBe(true);
    });

    it("filters devices when search input changes", async () => {
      await dialog.find('[data-test="All-tab"]').trigger("click");
      await flushPromises();

      const searchInput = dialog.find('[data-test="search-text"] input');
      await searchInput.setValue("test-device");
      await searchInput.trigger("keyup");
      await flushPromises();

      expect(devicesStore.fetchDeviceList).toHaveBeenCalled();
    });
  });

  describe("tab switching", () => {
    beforeEach(() => mountWrapper(true));

    it("fetches most used devices when switching to Suggested tab", async () => {
      await dialog.find('[data-test="All-tab"]').trigger("click");
      await flushPromises();

      await dialog.find('[data-test="Suggested-tab"]').trigger("click");
      await flushPromises();

      expect(devicesStore.fetchMostUsedDevices).toHaveBeenCalled();
    });

    it("fetches device list when switching to All tab", async () => {
      await dialog.find('[data-test="All-tab"]').trigger("click");
      await flushPromises();

      expect(devicesStore.fetchDeviceList).toHaveBeenCalled();
    });
  });

  describe("device selection", () => {
    it("disables Accept button when no devices selected on All tab", async () => {
      mountWrapper();
      devicesStore.selectedDevices = [];

      await dialog.find('[data-test="All-tab"]').trigger("click");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(true);
    });

    it("disables Accept button when more than 3 devices selected", async () => {
      mountWrapper();
      devicesStore.selectedDevices = [mockDevices[0], mockDevices[1], mockDevices[0], mockDevices[1]];

      await dialog.find('[data-test="All-tab"]').trigger("click");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(true);
    });

    it("enables Accept button when 1-3 devices selected on All tab", async () => {
      mountWrapper();
      devicesStore.selectedDevices = [mockDevices[0]];

      await dialog.find('[data-test="All-tab"]').trigger("click");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(false);
    });

    it("enables Accept button on Suggested tab regardless of selection", () => {
      mountWrapper(true);

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("confirmDisabled")).toBe(false);
    });
  });

  describe("accepting device choices", () => {
    it("sends suggested devices when accepting from Suggested tab", async () => {
      mountWrapper(true);
      devicesStore.suggestedDevices = mockDevices;

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      await formDialog.vm.$emit("confirm");
      await flushPromises();

      expect(devicesStore.sendDeviceChoices).toHaveBeenCalledWith(mockDevices);
    });

    it("sends selected devices when accepting from All tab", async () => {
      mountWrapper();
      devicesStore.selectedDevices = [mockDevices[0]];

      await dialog.find('[data-test="All-tab"]').trigger("click");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      await formDialog.vm.$emit("confirm");
      await flushPromises();

      expect(devicesStore.sendDeviceChoices).toHaveBeenCalledWith([mockDevices[0]]);
    });

    it("closes dialog after successful device choice", async () => {
      mountWrapper(true);

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      await formDialog.vm.$emit("confirm");
      await flushPromises();

      expect(devicesStore.showDeviceChooser).toBe(false);
    });

    it("refreshes device list after successful device choice", async () => {
      mountWrapper(true);

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      await formDialog.vm.$emit("confirm");
      await flushPromises();

      expect(devicesStore.fetchDeviceList).toHaveBeenCalledWith();
    });

    it("keeps dialog open on error", async () => {
      mountWrapper(true);
      vi.spyOn(devicesStore, "sendDeviceChoices").mockRejectedValue(createAxiosError(500, "Internal server error"));

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      await formDialog.vm.$emit("confirm");
      await flushPromises();

      expect(devicesStore.showDeviceChooser).toBe(true);
    });
  });

  describe("dialog closing", () => {
    beforeEach(() => mountWrapper());

    it("closes dialog when cancel is clicked", async () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      await formDialog.vm.$emit("cancel");
      await flushPromises();

      expect(devicesStore.showDeviceChooser).toBe(false);
    });

    it("closes dialog when close is clicked", async () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      await formDialog.vm.$emit("close");
      await flushPromises();

      expect(devicesStore.showDeviceChooser).toBe(false);
    });
  });
});
